package mapasci

import (
	"fmt"
	"math"
	"strings"
)

const (
	VerticalPaddingRows = 2
)

type Marker struct {
	Lon        float64
	Lat        float64
	Center     rune
	Horizontal rune
	Vertical   rune
	ArmX       int
	ArmY       int
}

func CharForLandFraction(fraction float64) (byte, error) {
	if !isFinite(fraction) {
		return 0, fmt.Errorf("land fraction must be finite")
	}
	if fraction < 0.0 || fraction > 1.0 {
		return 0, fmt.Errorf("land fraction must be in [0, 1], got %v", fraction)
	}

	if fraction < 0.12 {
		return ' ', nil
	}
	if fraction < 0.38 {
		return '.', nil
	}
	if fraction < 0.62 {
		return '*', nil
	}
	if fraction < 0.86 {
		return '@', nil
	}
	return '#', nil
}

func RenderWorldASCII(mask *LandMask, size int, supersample int, charAspect float64, marker *Marker) (string, error) {
	if err := validateMask(mask); err != nil {
		return "", err
	}
	if size <= 0 {
		return "", fmt.Errorf("size must be > 0, got %d", size)
	}
	if supersample <= 0 {
		return "", fmt.Errorf("supersample must be > 0, got %d", supersample)
	}
	if !isFinite(charAspect) || charAspect <= 0.0 {
		return "", fmt.Errorf("char_aspect must be > 0, got %v", charAspect)
	}

	mapWidth := size
	mapHeight := int(math.Round(float64(mapWidth) / (2.0 * charAspect)))
	if mapHeight <= 0 {
		return "", fmt.Errorf("size=%d with char_aspect=%v produces zero map height", size, charAspect)
	}

	subsamplesPerCell := supersample * supersample
	totalRows := mapHeight + (2 * VerticalPaddingRows)
	lines := make([][]byte, 0, totalRows)

	for row := 0; row < totalRows; row++ {
		line := make([]byte, mapWidth)
		if row < VerticalPaddingRows || row >= totalRows-VerticalPaddingRows {
			for i := range line {
				line[i] = ' '
			}
			lines = append(lines, line)
			continue
		}

		for col := 0; col < mapWidth; col++ {
			landSum := 0.0
			for sy := 0; sy < supersample; sy++ {
				for sx := 0; sx < supersample; sx++ {
					x := float64(col) + (float64(sx)+0.5)/float64(supersample)
					y := float64(row) + (float64(sy)+0.5)/float64(supersample)

					lon := (x/float64(mapWidth))*360.0 - 180.0
					yActive := y - float64(VerticalPaddingRows)
					t := yActive / float64(mapHeight)
					lat := 90.0 - (180.0 * t)

					landSum += sampleLandValueUnchecked(mask, lon, lat)
				}
			}

			landFraction := landSum / float64(subsamplesPerCell)
			ch, err := CharForLandFraction(landFraction)
			if err != nil {
				return "", err
			}
			line[col] = ch
		}

		lines = append(lines, line)
	}

	if marker != nil {
		if err := applyMarker(lines, mapWidth, mapHeight, *marker); err != nil {
			return "", err
		}
	}

	var b strings.Builder
	for idx, line := range lines {
		if _, err := b.Write(line); err != nil {
			return "", err
		}
		if idx != len(lines)-1 {
			if err := b.WriteByte('\n'); err != nil {
				return "", err
			}
		}
	}

	return b.String(), nil
}

func applyMarker(lines [][]byte, mapWidth int, mapHeight int, marker Marker) error {
	if !isFinite(marker.Lon) || !isFinite(marker.Lat) {
		return fmt.Errorf("marker lon and lat must be finite")
	}
	if marker.ArmX < -1 {
		return fmt.Errorf("marker ArmX must be >= -1, got %d", marker.ArmX)
	}
	if marker.ArmY < -1 {
		return fmt.Errorf("marker ArmY must be >= -1, got %d", marker.ArmY)
	}

	center, err := markerRuneOrDefault(marker.Center, 'O', "marker center")
	if err != nil {
		return err
	}
	horizontal, err := markerRuneOrDefault(marker.Horizontal, '-', "marker horizontal")
	if err != nil {
		return err
	}
	vertical, err := markerRuneOrDefault(marker.Vertical, '|', "marker vertical")
	if err != nil {
		return err
	}

	u := math.Mod((marker.Lon+180.0)/360.0, 1.0)
	if u < 0.0 {
		u += 1.0
	}
	v := clamp((90.0-marker.Lat)/180.0, 0.0, 1.0)

	xCenter := int(math.Round(u * float64(mapWidth-1)))
	yCenterActive := int(math.Round(v * float64(mapHeight-1)))
	yCenter := yCenterActive + VerticalPaddingRows

	xStart := 0
	xEnd := mapWidth - 1
	if marker.ArmX >= 0 {
		xStart = max(0, xCenter-marker.ArmX)
		xEnd = min(mapWidth-1, xCenter+marker.ArmX)
	}

	yStart := VerticalPaddingRows
	yEnd := VerticalPaddingRows + mapHeight - 1
	if marker.ArmY >= 0 {
		yStart = max(VerticalPaddingRows, yCenter-marker.ArmY)
		yEnd = min(VerticalPaddingRows+mapHeight-1, yCenter+marker.ArmY)
	}

	for y := yStart; y <= yEnd; y++ {
		lines[y][xCenter] = byte(vertical)
	}
	for x := xStart; x <= xEnd; x++ {
		lines[yCenter][x] = byte(horizontal)
	}
	lines[yCenter][xCenter] = byte(center)

	return nil
}

func markerRuneOrDefault(value rune, fallback rune, name string) (rune, error) {
	if value == 0 {
		value = fallback
	}
	if value > 127 {
		return 0, fmt.Errorf("%s must be ASCII", name)
	}
	return value, nil
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
