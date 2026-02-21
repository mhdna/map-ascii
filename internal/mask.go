package mapasci

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type LandMask struct {
	Width  int
	Height int
	Data   []float64
}

func LoadLandMask(maskPath string) (*LandMask, error) {
	if !strings.EqualFold(filepath.Ext(maskPath), ".png") {
		return nil, fmt.Errorf("mask file must be a PNG: %s", maskPath)
	}

	file, err := os.Open(maskPath)
	if err != nil {
		return nil, fmt.Errorf("open mask file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decode mask PNG: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width < 2 || height < 2 {
		return nil, fmt.Errorf("mask must be at least 2x2 pixels, got %dx%d", width, height)
	}

	data := make([]float64, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			gray := color.GrayModel.Convert(img.At(bounds.Min.X+x, bounds.Min.Y+y)).(color.Gray)
			data[y*width+x] = float64(gray.Y) / 255.0
		}
	}

	return &LandMask{Width: width, Height: height, Data: data}, nil
}

func SampleLandValue(mask *LandMask, lon float64, lat float64) (float64, error) {
	if err := validateMask(mask); err != nil {
		return 0, err
	}
	if !isFinite(lon) || !isFinite(lat) {
		return 0, fmt.Errorf("lon and lat must be finite")
	}

	return sampleLandValueUnchecked(mask, lon, lat), nil
}

func sampleLandValueUnchecked(mask *LandMask, lon float64, lat float64) float64 {
	u := math.Mod((lon+180.0)/360.0, 1.0)
	if u < 0.0 {
		u += 1.0
	}
	v := (90.0 - lat) / 180.0
	v = clamp(v, 0.0, 1.0)

	x := min(int(u*float64(mask.Width)), mask.Width-1)
	y := min(int(v*float64(mask.Height)), mask.Height-1)

	return mask.Data[y*mask.Width+x]
}

func validateMask(mask *LandMask) error {
	if mask == nil {
		return fmt.Errorf("mask must not be nil")
	}
	if mask.Width < 2 || mask.Height < 2 {
		return fmt.Errorf("mask must be at least 2x2, got %dx%d", mask.Width, mask.Height)
	}
	if len(mask.Data) != mask.Width*mask.Height {
		return fmt.Errorf("mask data length mismatch: got %d, expected %d", len(mask.Data), mask.Width*mask.Height)
	}

	for _, value := range mask.Data {
		if !isFinite(value) {
			return fmt.Errorf("mask contains non-finite values")
		}
		if value < 0.0 || value > 1.0 {
			return fmt.Errorf("mask values must be in [0, 1]")
		}
	}

	return nil
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func clamp(value float64, lo float64, hi float64) float64 {
	if value < lo {
		return lo
	}
	if value > hi {
		return hi
	}
	return value
}
