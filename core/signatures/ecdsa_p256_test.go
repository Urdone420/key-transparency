// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package signatures

import (
	"bytes"
	"crypto/rand"
	"math"
	"testing"
)

func TestConsistentKeyIDs(t *testing.T) {
	// Verify that the ID generated from from pub and from priv are the same.
	for _, tc := range []struct {
		priv string
		pub  string
	}{
		{testPrivKey, testPubKey},
	} {
		signer, err := SignerFromPEM(rand.Reader, []byte(tc.priv))
		if err != nil {
			t.Fatalf("SignerFromPEM(): %v", err)
		}
		verifier, err := VerifierFromPEM([]byte(tc.pub))
		if err != nil {
			t.Fatalf("VerifierFromPEM(): %v", err)
		}

		if got, want := signer.KeyID(), verifier.KeyID(); got != want {
			t.Errorf("signer.KeyID(): %v, want %v", got, want)
		}
	}
}

type env struct {
	signer   Signer
	verifier Verifier
}

func newEnv(t *testing.T) *env {
	signer, err := SignerFromPEM(rand.Reader, []byte(testPrivKey))
	if err != nil {
		t.Fatalf("SignerFromPEM(): %v", err)
	}
	verifier, err := VerifierFromPEM([]byte(testPubKey))
	if err != nil {
		t.Fatalf("VerifierFromPEM(): %v", err)
	}

	return &env{signer, verifier}
}

func TestSignVerifier(t *testing.T) {
	e := newEnv(t)
	for _, tc := range []struct {
		data interface{}
	}{
		{struct{ Foo string }{"bar"}},
	} {
		sig, err := e.signer.Sign(tc.data)
		if err != nil {
			t.Errorf("Sign(%v): %v", tc.data, err)
		}
		if err := e.verifier.Verify(tc.data, sig); err != nil {
			t.Errorf("Verify(%v, %v): %v", tc.data, sig, err)
		}
	}
}

func TestRightTruncateSignature(t *testing.T) {
	e := newEnv(t)
	data := struct{ Foo string }{"bar"}

	// Truncate bytes from the end of sig and try to verify.
	sig, err := e.signer.Sign(data)
	if err != nil {
		t.Errorf("Sign(%v): %v", data, err)
	}
	sigLen := len(sig.Signature)
	for i := 0; i < sigLen; i++ {
		sig.Signature = sig.Signature[:len(sig.Signature)-1]
		if err := e.verifier.Verify(data, sig); err == nil {
			t.Errorf("Verify unexpectedly succeeded after truncating %v bytes from the end of the signature", i)
		}
	}
}

func TestLeftTruncateSignature(t *testing.T) {
	e := newEnv(t)
	data := struct{ Foo string }{"bar"}

	// Truncate bytes from the beginning of sig and try to verify.
	sig, err := e.signer.Sign(data)
	if err != nil {
		t.Errorf("Sign(%v): %v", data, err)
	}
	sigLen := len(sig.Signature)
	for i := 0; i < sigLen; i++ {
		sig.Signature = sig.Signature[1:]
		if err := e.verifier.Verify(data, sig); err == nil {
			t.Errorf("Verify unexpectedly succeeded after truncating %v bytes from the beginning of the signature", i)
		}
	}
}

func TestBitFlipSignature(t *testing.T) {
	e := newEnv(t)
	data := struct{ Foo string }{"bar"}

	// Truncate bytes from the beginning of sig and try to verify.
	sig, err := e.signer.Sign(data)
	if err != nil {
		t.Errorf("Sign(%v): %v", data, err)
	}
	for i := 0; i < len(sig.Signature)*8; i++ {
		// Flip bit in position i.
		flippedSig := *sig
		flippedSig.Signature = flipBit(sig.Signature, uint(i))

		// Verify signature
		if err := e.verifier.Verify(data, &flippedSig); err == nil {
			t.Errorf("Verify unexpectedly succeeded after flipping bit %v of the signature", i)
		}
	}
}

func flipBit(a []byte, pos uint) []byte {
	index := int(math.Floor(float64(pos) / 8))
	b := byte(a[index])
	b ^= (1 << uint(math.Mod(float64(pos), 8.0)))

	var buf bytes.Buffer
	buf.Write(a[:index])
	buf.Write([]byte{b})
	buf.Write(a[index+1:])
	return buf.Bytes()
}