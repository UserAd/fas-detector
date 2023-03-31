package main

import (
	"io"
	"math"
	"math/cmplx"

	"github.com/mjibson/go-dsp/fft"
	"github.com/mjibson/go-dsp/window"
	"github.com/pkg/errors"
	"github.com/youpy/go-riff"
	"github.com/youpy/go-wav"
)

const (
	DETECTOR_WINDOW      = 2.0 // 1/2 of second
	DETECTOR_SENSIVITY   = 2.0 // twice
	DETECTOR_NOISE_LEVEL = -40 // dbm detection
)

type Detector struct {
	SampleRate  int
	File        riff.RIFFReader
	Detected    bool
	Reader      *wav.Reader
	SkipCounter int
	Indices     []int
}

func NewDetector(file riff.RIFFReader) *Detector {
	return &Detector{
		File:    file,
		Indices: []int{},
	}
}

func (d *Detector) Detect() (float32, bool, error) {
	reader := wav.NewReader(d.File)
	fm, err := reader.Format()

	if err != nil {
		return 0, false, errors.Wrap(err, "Detect")
	}

	d.SampleRate = int(fm.SampleRate)
	d.SkipCounter = d.SampleRate // Skip 1 second

	fftSize := int(math.Pow(2, math.Ceil(math.Log2(float64(d.SampleRate)))))
	spectralWidth := float64(d.SampleRate) / float64(fftSize)

	frequencies := []int{}

	for _, central := range []int{400, 425, 440, 450, 480} {
		for f := central - 5; f <= central+5; f++ {
			frequencies = append(frequencies, int(float64(f)/spectralWidth))
		}
	}
	d.Indices = frequencies

	maxDuration := d.SampleRate * 5
	count := 0

	w1 := make([]float64, fftSize)

	total := 0

	for {
		samples, err := reader.ReadSamples()
		if err == io.EOF {
			break
		}

		for _, sample := range samples {
			total += 1
			if d.SkipCounter > 0 {
				d.SkipCounter = d.SkipCounter - 1
			} else {
				if maxDuration > 0 {
					maxDuration = maxDuration - 1

					if count == d.SampleRate/DETECTOR_WINDOW {
						count = 0

						if d.checkWindow(w1) {
							return (float32(total)/float32(d.SampleRate) - 1/DETECTOR_WINDOW), true, nil
						}
						w1 = make([]float64, fftSize)
					}

					w1[count] = reader.FloatValue(sample, 0)

					count = count + 1
				} else {
					if d.checkWindow(w1) {
						return (float32(total)/float32(d.SampleRate) - 1/DETECTOR_WINDOW), true, nil
					}
					return 0, false, nil
				}
			}
		}
	}

	return 0, false, nil
}

func (d *Detector) checkWindow(w []float64) bool {
	window.Apply(w, window.Hamming)
	spectrogram := fft.FFTReal(w)

	var spectrum2 []float64
	length := float64(len(w))

	for i := range spectrogram {
		spectrum2 = append(spectrum2, cmplx.Abs(spectrogram[i])/length)
	}

	spectrum1 := spectrum2[0 : len(spectrum2)/2]

	dbm := make([]float64, len(spectrum1))
	for i := range spectrum1 {
		spectrum1[i] = spectrum1[i] * 2
		dbm[i] = 20 * math.Log10(spectrum1[i])
		if dbm[i] == math.Inf(-1) {
			dbm[i] = 0
		}
	}

	maxPositive := math.Inf(-1)
	maxNegative := math.Inf(-1)

	for index, value := range spectrum1 {
		if index > 100 {
			if value > DETECTOR_NOISE_LEVEL {
				if inArray(index, d.Indices) {
					if value > maxPositive {
						maxPositive = value
					}

				} else {
					if value > maxNegative {
						maxNegative = value
					}
				}
			}
		}
	}

	return maxPositive > maxNegative*DETECTOR_SENSIVITY
}

func inArray(val int, arr []int) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}
