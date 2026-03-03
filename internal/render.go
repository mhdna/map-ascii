package mapasci

import (
	"fmt"
	"math"
	"os"
	"strings"
)

const (
	DefaultVerticalMarginRows  = 2
	DefaultVerticalPaddingRows = DefaultVerticalMarginRows
	ansiReset                  = "\x1b[0m"
	colorModeNever             = "never"
	colorModeAuto              = "auto"
	colorModeAlways            = "always"
)

type RenderOptions struct {
	VerticalMarginRows  int
	VerticalPaddingRows int
	Frame               bool
	ColorMode           string
	MapColor            string
	FrameColor          string
	MarkerColor         string
	Viewport            *Viewport
}

type Viewport struct {
	MinLon float64
	MinLat float64
	MaxLon float64
	MaxLat float64
}

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
	return RenderWorldASCIIWithOptions(mask, size, supersample, charAspect, marker, nil)
}

func RenderWorldASCIIWithOptions(mask *LandMask, size int, supersample int, charAspect float64, marker *Marker, options *RenderOptions) (string, error) {
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

	verticalMarginRows := DefaultVerticalMarginRows
	frame := false
	colorMode := colorModeNever
	mapColorName := ""
	frameColorName := ""
	markerColorName := ""
	viewport := defaultWorldViewport()
	if options != nil {
		if options.VerticalMarginRows < 0 {
			return "", fmt.Errorf("vertical margin rows must be >= 0, got %d", options.VerticalMarginRows)
		}
		if options.VerticalPaddingRows < 0 {
			return "", fmt.Errorf("vertical padding rows must be >= 0, got %d", options.VerticalPaddingRows)
		}
		if options.VerticalMarginRows != 0 {
			verticalMarginRows = options.VerticalMarginRows
		} else {
			verticalMarginRows = options.VerticalPaddingRows
		}
		frame = options.Frame
		if options.ColorMode != "" {
			colorMode = options.ColorMode
		}
		mapColorName = options.MapColor
		frameColorName = options.FrameColor
		markerColorName = options.MarkerColor
		if options.Viewport != nil {
			viewport = *options.Viewport
		}
	}

	if err := validateViewport(viewport); err != nil {
		return "", err
	}

	colorEnabled, err := shouldColorize(colorMode)
	if err != nil {
		return "", err
	}

	mapColor, err := colorSequenceForName(mapColorName, "map color")
	if err != nil {
		return "", err
	}
	frameColor, err := colorSequenceForName(frameColorName, "frame color")
	if err != nil {
		return "", err
	}
	markerColor, err := colorSequenceForName(markerColorName, "marker color")
	if err != nil {
		return "", err
	}

	mapWidth := size
	mapHeight := int(math.Round((float64(mapWidth) * viewport.latSpan() / viewport.lonSpan()) / charAspect))
	if mapHeight <= 0 {
		return "", fmt.Errorf("size=%d with char_aspect=%v and viewport produces zero map height", size, charAspect)
	}

	subsamplesPerCell := supersample * supersample
	lines := make([][]byte, 0, mapHeight)

	for row := 0; row < mapHeight; row++ {
		line := make([]byte, mapWidth)
		for col := 0; col < mapWidth; col++ {
			landSum := 0.0
			for sy := 0; sy < supersample; sy++ {
				for sx := 0; sx < supersample; sx++ {
					x := float64(col) + (float64(sx)+0.5)/float64(supersample)
					y := float64(row) + (float64(sy)+0.5)/float64(supersample)

					lon := viewport.MinLon + (x/float64(mapWidth))*viewport.lonSpan()
					t := y / float64(mapHeight)
					lat := viewport.MaxLat - (viewport.latSpan() * t)

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

	var markerMask []bool
	if marker != nil {
		markerMask, err = applyMarker(lines, mapWidth, mapHeight, *marker, viewport)
		if err != nil {
			return "", err
		}
	}

	if frame {
		lines = frameLines(lines, mapWidth)
	}
	if verticalMarginRows > 0 {
		lines = addVerticalMargins(lines, verticalMarginRows)
	}

	useColor := colorEnabled && (mapColor != "" || frameColor != "" || markerColor != "")
	if useColor {
		return buildColoredOutput(lines, mapWidth, mapHeight, verticalMarginRows, frame, markerMask, mapColor, frameColor, markerColor)
	}

	return buildPlainOutput(lines)
}

func defaultWorldViewport() Viewport {
	return Viewport{
		MinLon: -180.0,
		MinLat: -90.0,
		MaxLon: 180.0,
		MaxLat: 90.0,
	}
}

func (v Viewport) lonSpan() float64 {
	return v.MaxLon - v.MinLon
}

func (v Viewport) latSpan() float64 {
	return v.MaxLat - v.MinLat
}

func validateViewport(v Viewport) error {
	if !isFinite(v.MinLon) || !isFinite(v.MinLat) || !isFinite(v.MaxLon) || !isFinite(v.MaxLat) {
		return fmt.Errorf("viewport values must be finite")
	}
	if v.MinLon < -180.0 || v.MinLon > 180.0 || v.MaxLon < -180.0 || v.MaxLon > 180.0 {
		return fmt.Errorf("viewport longitude must be in [-180, 180]")
	}
	if v.MinLat < -90.0 || v.MinLat > 90.0 || v.MaxLat < -90.0 || v.MaxLat > 90.0 {
		return fmt.Errorf("viewport latitude must be in [-90, 90]")
	}
	if v.MinLon >= v.MaxLon {
		return fmt.Errorf("viewport min lon must be less than max lon")
	}
	if v.MinLat >= v.MaxLat {
		return fmt.Errorf("viewport min lat must be less than max lat")
	}

	return nil
}

func buildPlainOutput(lines [][]byte) (string, error) {
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

func buildColoredOutput(
	lines [][]byte,
	mapWidth int,
	mapHeight int,
	verticalMarginRows int,
	frame bool,
	markerMask []bool,
	mapColor string,
	frameColor string,
	markerColor string,
) (string, error) {
	var b strings.Builder
	for rowIdx, line := range lines {
		currentColor := ""
		for colIdx, ch := range line {
			nextColor := colorForCell(rowIdx, colIdx, mapWidth, mapHeight, verticalMarginRows, frame, markerMask, mapColor, frameColor, markerColor)
			if nextColor != currentColor {
				if nextColor == "" {
					if _, err := b.WriteString(ansiReset); err != nil {
						return "", err
					}
				} else {
					if _, err := b.WriteString(nextColor); err != nil {
						return "", err
					}
				}
				currentColor = nextColor
			}

			if err := b.WriteByte(ch); err != nil {
				return "", err
			}
		}

		if currentColor != "" {
			if _, err := b.WriteString(ansiReset); err != nil {
				return "", err
			}
		}

		if rowIdx != len(lines)-1 {
			if err := b.WriteByte('\n'); err != nil {
				return "", err
			}
		}
	}

	return b.String(), nil
}

func colorForCell(
	row int,
	col int,
	mapWidth int,
	mapHeight int,
	verticalMarginRows int,
	frame bool,
	markerMask []bool,
	mapColor string,
	frameColor string,
	markerColor string,
) string {
	mapRowStart := verticalMarginRows
	mapColStart := 0
	if frame {
		mapRowStart++
		mapColStart = 1
	}

	mapRow := row - mapRowStart
	mapCol := col - mapColStart
	inMap := mapRow >= 0 && mapRow < mapHeight && mapCol >= 0 && mapCol < mapWidth

	if inMap && len(markerMask) == mapWidth*mapHeight && markerMask[(mapRow*mapWidth)+mapCol] {
		if markerColor != "" {
			return markerColor
		}
		return mapColor
	}

	if frame {
		frameTopRow := verticalMarginRows
		frameBottomRow := verticalMarginRows + mapHeight + 1
		frameRightCol := mapWidth + 1
		if row >= frameTopRow && row <= frameBottomRow {
			if row == frameTopRow || row == frameBottomRow || col == 0 || col == frameRightCol {
				return frameColor
			}
		}
	}

	if inMap {
		return mapColor
	}

	return ""
}

func shouldColorize(mode string) (bool, error) {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	switch normalized {
	case "", colorModeNever:
		return false, nil
	case colorModeAlways:
		return true, nil
	case colorModeAuto:
		if os.Getenv("NO_COLOR") != "" {
			return false, nil
		}

		term := strings.TrimSpace(os.Getenv("TERM"))
		if term == "" || term == "dumb" {
			return false, nil
		}

		stdoutInfo, err := os.Stdout.Stat()
		if err != nil {
			return false, nil
		}

		return (stdoutInfo.Mode() & os.ModeCharDevice) != 0, nil
	default:
		return false, fmt.Errorf("color mode must be one of: never, auto, always")
	}
}

func colorSequenceForName(name string, objectName string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", nil
	}

	code, ok := ansi16ColorCodes[strings.ToLower(strings.TrimSpace(name))]
	if !ok {
		return "", fmt.Errorf("%s must be one of: %s", objectName, ansi16ColorNamesCSV)
	}

	return "\x1b[" + code + "m", nil
}

var ansi16ColorCodes = map[string]string{
	"black":          "30",
	"red":            "31",
	"green":          "32",
	"yellow":         "33",
	"blue":           "34",
	"magenta":        "35",
	"cyan":           "36",
	"white":          "37",
	"bright-black":   "90",
	"bright-red":     "91",
	"bright-green":   "92",
	"bright-yellow":  "93",
	"bright-blue":    "94",
	"bright-magenta": "95",
	"bright-cyan":    "96",
	"bright-white":   "97",
}

const ansi16ColorNamesCSV = "black, red, green, yellow, blue, magenta, cyan, white, bright-black, bright-red, bright-green, bright-yellow, bright-blue, bright-magenta, bright-cyan, bright-white"

func applyMarker(lines [][]byte, mapWidth int, mapHeight int, marker Marker, viewport Viewport) ([]bool, error) {
	if !isFinite(marker.Lon) || !isFinite(marker.Lat) {
		return nil, fmt.Errorf("marker lon and lat must be finite")
	}
	if marker.ArmX < -1 {
		return nil, fmt.Errorf("marker ArmX must be >= -1, got %d", marker.ArmX)
	}
	if marker.ArmY < -1 {
		return nil, fmt.Errorf("marker ArmY must be >= -1, got %d", marker.ArmY)
	}

	center, err := markerRuneOrDefault(marker.Center, 'O', "marker center")
	if err != nil {
		return nil, err
	}
	horizontal, err := markerRuneOrDefault(marker.Horizontal, '-', "marker horizontal")
	if err != nil {
		return nil, err
	}
	vertical, err := markerRuneOrDefault(marker.Vertical, '|', "marker vertical")
	if err != nil {
		return nil, err
	}

	markerMask := make([]bool, mapWidth*mapHeight)

	u := normalizeLongitude(marker.Lon, viewport)
	v := clamp((viewport.MaxLat-marker.Lat)/viewport.latSpan(), 0.0, 1.0)

	xCenter := int(math.Round(u * float64(mapWidth-1)))
	yCenterActive := int(math.Round(v * float64(mapHeight-1)))
	yCenter := yCenterActive

	xStart := 0
	xEnd := mapWidth - 1
	if marker.ArmX >= 0 {
		xStart = max(0, xCenter-marker.ArmX)
		xEnd = min(mapWidth-1, xCenter+marker.ArmX)
	}

	yStart := 0
	yEnd := mapHeight - 1
	if marker.ArmY >= 0 {
		yStart = max(0, yCenter-marker.ArmY)
		yEnd = min(mapHeight-1, yCenter+marker.ArmY)
	}

	for y := yStart; y <= yEnd; y++ {
		lines[y][xCenter] = byte(vertical)
		markerMask[(y*mapWidth)+xCenter] = true
	}
	for x := xStart; x <= xEnd; x++ {
		lines[yCenter][x] = byte(horizontal)
		markerMask[(yCenter*mapWidth)+x] = true
	}
	lines[yCenter][xCenter] = byte(center)
	markerMask[(yCenter*mapWidth)+xCenter] = true

	return markerMask, nil
}

func normalizeLongitude(lon float64, viewport Viewport) float64 {
	u := (lon - viewport.MinLon) / viewport.lonSpan()
	if viewport.lonSpan() >= 360.0 {
		u = math.Mod(u, 1.0)
		if u < 0.0 {
			u += 1.0
		}
		return u
	}

	return clamp(u, 0.0, 1.0)
}

func frameLines(lines [][]byte, width int) [][]byte {
	framed := make([][]byte, 0, len(lines)+2)

	top := make([]byte, width+2)
	top[0] = '+'
	top[len(top)-1] = '+'
	for i := 1; i < len(top)-1; i++ {
		top[i] = '-'
	}
	framed = append(framed, top)

	for _, line := range lines {
		framedLine := make([]byte, width+2)
		framedLine[0] = '|'
		copy(framedLine[1:], line)
		framedLine[len(framedLine)-1] = '|'
		framed = append(framed, framedLine)
	}

	bottom := make([]byte, width+2)
	copy(bottom, top)
	framed = append(framed, bottom)

	return framed
}

func addVerticalMargins(lines [][]byte, marginRows int) [][]byte {
	withMargins := make([][]byte, 0, len(lines)+(2*marginRows))
	for i := 0; i < marginRows; i++ {
		withMargins = append(withMargins, []byte{})
	}
	withMargins = append(withMargins, lines...)
	for i := 0; i < marginRows; i++ {
		withMargins = append(withMargins, []byte{})
	}

	return withMargins
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
