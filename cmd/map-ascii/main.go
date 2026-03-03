package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	mapascii "github.com/Kivayan/map-ascii"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	size := flag.Int("size", 60, "Map width in characters")
	supersample := flag.Int("supersample", 3, "NxN supersampling per ASCII cell")
	charAspect := flag.Float64("char-aspect", 2.0, "Character height/width ratio used for output height")
	marginY := mapascii.DefaultVerticalMarginRows
	continentChoices := mapascii.ContinentNamesCSV()
	flag.IntVar(&marginY, "margin-y", mapascii.DefaultVerticalMarginRows, "Empty rows above and below the map (outside the frame)")
	flag.IntVar(&marginY, "padding-y", mapascii.DefaultVerticalMarginRows, "Deprecated alias for --margin-y")
	frame := flag.Bool("frame", false, "Draw an ASCII frame around the map")
	colorMode := flag.String("color", "auto", "Color output mode: never, auto, always")
	mapColor := flag.String("map-color", "", "Map color name (16 ANSI colors)")
	frameColor := flag.String("frame-color", "", "Frame color name (16 ANSI colors)")
	markerColor := flag.String("marker-color", "", "Marker color name (16 ANSI colors)")
	outputPath := flag.String("output", "", "Optional output text file")
	maskPath := flag.String("mask", "", "Path to land mask PNG (optional; embedded default is used when omitted)")
	continent := flag.String("continent", "", "Continent preset viewport: "+continentChoices)
	bbox := flag.String("bbox", "", "Viewport bounding box as minLon,minLat,maxLon,maxLat")

	markerLon := flag.Float64("marker-lon", math.NaN(), "Marker longitude")
	markerLat := flag.Float64("marker-lat", math.NaN(), "Marker latitude")
	markerCenter := flag.String("marker-center", "O", "Marker center character")
	markerHorizontal := flag.String("marker-horizontal", "-", "Marker horizontal character")
	markerVertical := flag.String("marker-vertical", "|", "Marker vertical character")
	markerArmX := flag.Int("marker-arm-x", -1, "Horizontal arm length in characters (-1 for full width)")
	markerArmY := flag.Int("marker-arm-y", -1, "Vertical arm length in characters (-1 for full map height)")
	animateMarker := flag.Bool("animate-marker", false, "Animate marker output in terminal")
	animateFPS := flag.Int("animate-fps", mapascii.DefaultAnimationFPS, "Animation refresh rate in frames per second")
	animateStyle := flag.String("animate-style", string(mapascii.AnimationStylePulseColor), "Animation style: pulse-color, blink")
	animateDuration := flag.Duration("animate-duration", 0, "Animation duration (for example: 10s, 2m). 0 means run until interrupted")

	flag.Parse()

	if *size <= 0 {
		return fmt.Errorf("size must be > 0, got %d", *size)
	}
	if *supersample <= 0 {
		return fmt.Errorf("supersample must be > 0, got %d", *supersample)
	}
	if !isFinite(*charAspect) || *charAspect <= 0.0 {
		return fmt.Errorf("char-aspect must be > 0, got %v", *charAspect)
	}
	if marginY < 0 {
		return fmt.Errorf("margin-y must be >= 0, got %d", marginY)
	}

	mask, err := loadMask(*maskPath)
	if err != nil {
		return err
	}

	marker, err := buildMarker(*markerLon, *markerLat, *markerCenter, *markerHorizontal, *markerVertical, *markerArmX, *markerArmY)
	if err != nil {
		return err
	}

	viewport, err := buildViewport(*continent, *bbox)
	if err != nil {
		return err
	}

	renderColorMode := *colorMode
	if *outputPath != "" && strings.EqualFold(renderColorMode, "auto") {
		renderColorMode = "never"
	}
	renderOptions := &mapascii.RenderOptions{
		VerticalMarginRows: marginY,
		Frame:              *frame,
		ColorMode:          renderColorMode,
		MapColor:           *mapColor,
		FrameColor:         *frameColor,
		MarkerColor:        *markerColor,
		Viewport:           viewport,
	}

	if *animateMarker {
		if *outputPath != "" {
			return fmt.Errorf("animate-marker cannot be used with output file")
		}
		if marker == nil {
			return fmt.Errorf("animate-marker requires marker-lon and marker-lat")
		}
		if !stdoutIsTTY() {
			return fmt.Errorf("animate-marker requires a TTY stdout")
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		writer := &terminalFrameWriter{}
		defer writer.Close()

		err := mapascii.StreamWorldASCIIAnimation(
			ctx,
			mask,
			*size,
			*supersample,
			*charAspect,
			marker,
			renderOptions,
			&mapascii.AnimationOptions{
				FPS:      *animateFPS,
				Style:    mapascii.AnimationStyle(*animateStyle),
				Duration: *animateDuration,
			},
			writer.Emit,
		)
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}

	asciiMap, err := mapascii.RenderWorldASCIIWithOptions(mask, *size, *supersample, *charAspect, marker, renderOptions)
	if err != nil {
		return err
	}

	if *outputPath == "" {
		fmt.Println(asciiMap)
		return nil
	}

	parent := filepath.Dir(*outputPath)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	if err := os.WriteFile(*outputPath, []byte(asciiMap+"\n"), 0o644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	fmt.Printf("Wrote ASCII map to %s\n", *outputPath)
	return nil
}

func buildMarker(
	lon float64,
	lat float64,
	centerRaw string,
	horizontalRaw string,
	verticalRaw string,
	armX int,
	armY int,
) (*mapascii.Marker, error) {
	hasLon := isFinite(lon)
	hasLat := isFinite(lat)

	if hasLon != hasLat {
		return nil, fmt.Errorf("marker-lon and marker-lat must be provided together")
	}
	if !hasLon {
		return nil, nil
	}

	center, err := parseSingleRune(centerRaw, "marker-center")
	if err != nil {
		return nil, err
	}
	horizontal, err := parseSingleRune(horizontalRaw, "marker-horizontal")
	if err != nil {
		return nil, err
	}
	vertical, err := parseSingleRune(verticalRaw, "marker-vertical")
	if err != nil {
		return nil, err
	}

	if armX < -1 {
		return nil, fmt.Errorf("marker-arm-x must be >= -1, got %d", armX)
	}
	if armY < -1 {
		return nil, fmt.Errorf("marker-arm-y must be >= -1, got %d", armY)
	}

	return &mapascii.Marker{
		Lon:        lon,
		Lat:        lat,
		Center:     center,
		Horizontal: horizontal,
		Vertical:   vertical,
		ArmX:       armX,
		ArmY:       armY,
	}, nil
}

func parseSingleRune(raw string, flagName string) (rune, error) {
	if raw == "" {
		return 0, fmt.Errorf("%s must contain one ASCII character", flagName)
	}
	if utf8.RuneCountInString(raw) != 1 {
		return 0, fmt.Errorf("%s must contain exactly one character", flagName)
	}

	r, _ := utf8.DecodeRuneInString(raw)
	if r > 127 {
		return 0, fmt.Errorf("%s must be ASCII", flagName)
	}
	return r, nil
}

func loadMask(maskPath string) (*mapascii.LandMask, error) {
	if maskPath != "" {
		mask, err := mapascii.LoadLandMask(maskPath)
		if err != nil {
			return nil, err
		}
		return mask, nil
	}

	mask, err := mapascii.LoadLandMask("data/landmask_3600x1800.png")
	if err == nil {
		return mask, nil
	}

	mask, embeddedErr := mapascii.LoadEmbeddedDefaultLandMask()
	if embeddedErr != nil {
		return nil, fmt.Errorf("load default mask: %w", embeddedErr)
	}
	return mask, nil
}

func buildViewport(continentRaw string, bboxRaw string) (*mapascii.Viewport, error) {
	continent := strings.TrimSpace(continentRaw)
	bbox := strings.TrimSpace(bboxRaw)

	if continent != "" && bbox != "" {
		return nil, fmt.Errorf("continent and bbox cannot be used together")
	}

	if bbox != "" {
		viewport, err := parseBBox(bbox)
		if err != nil {
			return nil, err
		}
		return viewport, nil
	}

	if continent == "" {
		return nil, nil
	}

	viewport, err := mapascii.ViewportForContinent(continent)
	if err != nil {
		return nil, err
	}

	return &viewport, nil
}

func parseBBox(raw string) (*mapascii.Viewport, error) {
	parts := strings.Split(raw, ",")
	if len(parts) != 4 {
		return nil, fmt.Errorf("bbox must be minLon,minLat,maxLon,maxLat")
	}

	values := make([]float64, 4)
	for idx, part := range parts {
		parsed, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return nil, fmt.Errorf("bbox value %q is not a valid number: %w", strings.TrimSpace(part), err)
		}
		values[idx] = parsed
	}

	return &mapascii.Viewport{
		MinLon: values[0],
		MinLat: values[1],
		MaxLon: values[2],
		MaxLat: values[3],
	}, nil
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

type terminalFrameWriter struct {
	initialized bool
	lineCount   int
}

func (w *terminalFrameWriter) Emit(frame mapascii.Frame) error {
	if !w.initialized {
		if _, err := fmt.Fprint(os.Stdout, "\x1b[?25l"); err != nil {
			return err
		}
		w.initialized = true
	} else if w.lineCount > 1 {
		if _, err := fmt.Fprintf(os.Stdout, "\r\x1b[%dA", w.lineCount-1); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprint(os.Stdout, "\r"); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprint(os.Stdout, frame.Text); err != nil {
		return err
	}

	w.lineCount = strings.Count(frame.Text, "\n") + 1
	return nil
}

func (w *terminalFrameWriter) Close() {
	if !w.initialized {
		return
	}
	_, _ = fmt.Fprint(os.Stdout, "\x1b[0m\x1b[?25h\n")
}

func stdoutIsTTY() bool {
	stdoutInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return (stdoutInfo.Mode() & os.ModeCharDevice) != 0
}
