package mapascii

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"

	internal "github.com/Kivayan/map-ascii/internal"
)

//go:embed data/landmask_3600x1800.png
var embeddedDefaultMaskPNG []byte

type LandMask = internal.LandMask

type Marker = internal.Marker

type RenderOptions = internal.RenderOptions

const DefaultVerticalMarginRows = internal.DefaultVerticalMarginRows
const VerticalMarginRows = internal.DefaultVerticalMarginRows
const DefaultVerticalPaddingRows = internal.DefaultVerticalPaddingRows
const VerticalPaddingRows = internal.DefaultVerticalPaddingRows

func LoadLandMask(maskPath string) (*LandMask, error) {
	return internal.LoadLandMask(maskPath)
}

func LoadEmbeddedDefaultLandMask() (*LandMask, error) {
	img, _, err := image.Decode(bytes.NewReader(embeddedDefaultMaskPNG))
	if err != nil {
		return nil, fmt.Errorf("decode embedded mask PNG: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width < 2 || height < 2 {
		return nil, fmt.Errorf("embedded mask must be at least 2x2 pixels, got %dx%d", width, height)
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
	return internal.SampleLandValue(mask, lon, lat)
}

func CharForLandFraction(fraction float64) (byte, error) {
	return internal.CharForLandFraction(fraction)
}

func RenderWorldASCII(mask *LandMask, size int, supersample int, charAspect float64, marker *Marker) (string, error) {
	return internal.RenderWorldASCII(mask, size, supersample, charAspect, marker)
}

func RenderWorldASCIIWithOptions(mask *LandMask, size int, supersample int, charAspect float64, marker *Marker, options *RenderOptions) (string, error) {
	return internal.RenderWorldASCIIWithOptions(mask, size, supersample, charAspect, marker, options)
}
