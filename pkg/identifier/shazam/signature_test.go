// This file is vendored from https://github.com/lukechampine/barbershop
// Original copyright (c) 2024 Luke Champine, MIT licensed (see LICENSE).

package shazam

import (
	"crypto/sha256"
	"fmt"
	"math"
	"testing"
)

func TestSignature(t *testing.T) {
	samples := make([]float64, 128)
	sig := ComputeSignature(16000, samples)
	h := sha256.Sum256(sig.encode())
	if fmt.Sprintf("%x", h) != "4ae7d1ae7a4787a7d6cda559db6e17026f60369b3485b762759b7a07ff24fab9" {
		t.Fatalf("bad signature: %x", h)
	}

	samples = make([]float64, 1024)
	for i := range samples {
		samples[i] = float64(i)
	}
	sig = ComputeSignature(16000, samples)
	h = sha256.Sum256(sig.encode())
	if fmt.Sprintf("%x", h) != "073022772a4bc617a855adfb6265316f23ae6a25045e670e0904a2b11f132a75" {
		t.Fatalf("bad signature: %x", h)
	}

	samples = make([]float64, 16*1024)
	for i := range samples {
		samples[i] = math.Sin(float64(i) * 2 * math.Pi / 256)
	}
	sig = ComputeSignature(16000, samples)
	h = sha256.Sum256(sig.encode())
	if fmt.Sprintf("%x", h) != "c8c055411ec845f6d57b27baf7fc5735fdaf51f2a6026dd12f09d0eb17652c02" {
		t.Fatalf("bad signature: %x", h)
	}

	samples = make([]float64, 7*1024+55)
	for i := range samples {
		samples[i] = math.Cos(float64(i+12) * math.E)
	}
	sig = ComputeSignature(16000, samples)
	h = sha256.Sum256(sig.encode())
	if fmt.Sprintf("%x", h) != "e399475137268c73d7e6665479358370c0979f7d2c3860f71b1e035105b3a1d8" {
		t.Fatalf("bad signature: %x", h)
	}
}
