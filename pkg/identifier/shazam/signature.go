// This file is vendored from https://github.com/lukechampine/barbershop
// Original copyright (c) 2024 Luke Champine, MIT licensed (see LICENSE).
//
// Modifications for rig:
//   - Dropped the optional CollectSample helper that depended on faiface/beep.
//   - Dropped the unused Signature.decode round-trip method.
//   - Added explicit error handling on binary writes where required by linters.

package shazam

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"math"

	"gonum.org/v1/gonum/dsp/fourier"
)

func convertSampleRate(x int) int {
	return map[int]int{
		1: 8000,
		2: 11025,
		3: 16000,
		4: 32000,
		5: 44100,

		8000:  1,
		11025: 2,
		16000: 3,
		32000: 4,
		44100: 5,
	}[x]
}

type frequencyPeak struct {
	pass      int
	magnitude int
	bin       int
}

// A Signature is a unique fingerprint of an audio sample.
type Signature struct {
	sampleRate  int
	numSamples  int
	peaksByBand [5][]frequencyPeak
}

// encode serialises the signature into Shazam's wire format.
// Writes to bytes.Buffer never fail, so binary.Write errors are intentionally
// ignored. Integer truncations on this path are deliberate bit packing for
// Shazam's wire format, so gosec G115 is suppressed throughout.
//
//nolint:gosec // G115: intentional bit packing for Shazam wire format
func (s Signature) encode() []byte {
	var buf []byte
	write := func(u uint32) {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], u)
		buf = append(buf, b[:]...)
	}

	// header
	write(0xcafe2580)
	write(0) // checksum
	write(0) // length
	write(0x94119c00)
	write(0)
	write(0)
	write(0)
	write(uint32(convertSampleRate(s.sampleRate)) << 27)
	write(0)
	write(0)
	write(uint32(s.numSamples) + uint32(float64(s.sampleRate)*0.24))
	write(0x007c0000)
	write(uint32(0x40000000))
	write(0) // length2

	// peaks
	for band, peaks := range s.peaksByBand {
		if len(peaks) == 0 {
			continue
		}
		var peakBuf bytes.Buffer
		pass := 0
		for _, peak := range peaks {
			if peak.pass-pass >= 255 {
				peakBuf.WriteByte(0xFF)
				_ = binary.Write(&peakBuf, binary.LittleEndian, uint32(peak.pass))
				pass = peak.pass
			}
			_ = binary.Write(&peakBuf, binary.LittleEndian, uint8(peak.pass-pass))
			_ = binary.Write(&peakBuf, binary.LittleEndian, uint16(peak.magnitude))
			_ = binary.Write(&peakBuf, binary.LittleEndian, uint16(peak.bin))
			pass = peak.pass
		}
		write(uint32(0x60030040 + band))
		write(uint32(peakBuf.Len()))
		for peakBuf.Len()%4 != 0 {
			peakBuf.WriteByte(0x00)
		}
		buf = append(buf, peakBuf.Bytes()...)
	}

	binary.LittleEndian.PutUint32(buf[8:12], uint32(len(buf[48:])))
	binary.LittleEndian.PutUint32(buf[52:56], uint32(len(buf[48:])))
	binary.LittleEndian.PutUint32(buf[4:8], crc32.ChecksumIEEE(buf[8:]))
	return buf
}

type ring[T any] struct {
	buf   []T
	index int
}

func (r ring[T]) mod(i int) int {
	for i < 0 {
		i += len(r.buf)
	}
	return i % len(r.buf)
}

func (r ring[T]) At(i int) *T {
	return &r.buf[r.mod(r.index+i)]
}

func (r ring[T]) Append(x ...T) ring[T] {
	for len(x) > 0 {
		n := copy(r.buf[r.index:], x)
		x = x[n:]
		r.index = (r.index + n) % len(r.buf)
	}
	return r
}

func (r ring[T]) Slice(s []T, offset int) {
	offset = r.mod(offset + r.index)
	for len(s) > 0 {
		n := copy(s, r.buf[offset:])
		s = s[n:]
		offset = (offset + n) % len(r.buf)
	}
}

func newRing[T any](size int) ring[T] {
	return ring[T]{buf: make([]T, size)}
}

// ComputeSignature computes the audio signature of the provided mono samples
// at the given sample rate. The sample rate must be one of 8000, 11025, 16000,
// 32000, or 44100 Hz.
func ComputeSignature(sampleRate int, samples []float64) Signature {
	maxNeighbor := func(spreadOutputs ring[[1025]float64], i int) (neighbor float64) {
		for _, off := range []int{-10, -7, -4, -3, 1, 2, 5, 8} {
			neighbor = max(neighbor, spreadOutputs.At(-49)[(i+off)])
		}
		for _, off := range []int{-53, -45, 165, 172, 179, 186, 193, 200, 214, 221, 228, 235, 242, 249} {
			neighbor = max(neighbor, spreadOutputs.At(off)[i-1])
		}
		return neighbor
	}
	normalizePeak := func(x float64) float64 {
		return math.Log(max(x, 1.0/64))*1477.3 + 6144
	}
	peakBand := func(bin int) (int, bool) {
		hz := (bin * sampleRate) / (2 * 1024 * 64)
		band, ok := map[bool]int{
			250 <= hz && hz < 520:    0,
			520 <= hz && hz < 1450:   1,
			1450 <= hz && hz < 3500:  2,
			3500 <= hz && hz <= 5500: 3,
		}[true]
		return band, ok
	}

	fft := fourier.NewFFT(2048)
	samplesRing := newRing[float64](2048)
	fftOutputs := newRing[[1025]float64](256)
	spreadOutputs := newRing[[1025]float64](256)
	var peaksByBand [5][]frequencyPeak
	for i := 0; i*128+128 < len(samples); i++ {
		samplesRing = samplesRing.Append(samples[i*128:][:128]...)

		// Perform FFT.
		reorderedSamples := make([]float64, 2048)
		samplesRing.Slice(reorderedSamples, 0)
		for j, m := range &hanningMultipliers {
			reorderedSamples[j] = math.Round(reorderedSamples[j]*1024*64) * m
		}
		var outputs [1025]float64
		for k, c := range fft.Coefficients(nil, reorderedSamples) {
			outputs[k] = max((real(c)*real(c)+imag(c)*imag(c))/(1<<17), 0.0000000001)
		}
		fftOutputs = fftOutputs.Append(outputs)

		// Spread peaks, both in the frequency domain...
		for j := 0; j < len(outputs)-2; j++ {
			outputs[j] = max(outputs[j], outputs[j+1], outputs[j+2])
		}
		spreadOutputs = spreadOutputs.Append(outputs)
		// ... and in the time domain.
		for _, off := range []int{-2, -4, -7} {
			prev := spreadOutputs.At(off)
			for j := range prev {
				prev[j] = max(prev[j], outputs[j])
			}
		}

		// Accumulate samples until we have enough...
		if i < 45 {
			continue
		}
		// ...then recognise peaks.
		fftOutput := fftOutputs.At(-46)
		for bin := 10; bin < 1015; bin++ {
			// Ensure that this is a frequency- and time-domain local maximum.
			if fftOutput[bin] <= maxNeighbor(spreadOutputs, bin) {
				continue
			}
			// Normalise and compute frequency band.
			before := normalizePeak(fftOutput[bin-1])
			peak := normalizePeak(fftOutput[bin])
			after := normalizePeak(fftOutput[bin+1])
			variation := int((32 * (after - before)) / (2*peak - after - before))
			peakBin := bin*64 + variation
			band, ok := peakBand(peakBin)
			if !ok {
				continue
			}
			peaksByBand[band] = append(peaksByBand[band], frequencyPeak{
				pass:      i - 45,
				magnitude: int(peak),
				bin:       peakBin,
			})
		}
	}
	return Signature{
		sampleRate:  sampleRate,
		numSamples:  len(samples),
		peaksByBand: peaksByBand,
	}
}
