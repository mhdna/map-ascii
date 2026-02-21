package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"unicode/utf8"

	mapasci "map-asci-go/internal"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	defaultMaskPath := resolveDefaultMaskPath()

	size := flag.Int("size", 60, "Map width in characters")
	supersample := flag.Int("supersample", 3, "NxN supersampling per ASCII cell")
	charAspect := flag.Float64("char-aspect", 2.0, "Character height/width ratio used for output height")
	outputPath := flag.String("output", "", "Optional output text file")
	maskPath := flag.String("mask", defaultMaskPath, "Path to land mask PNG")

	markerLon := flag.Float64("marker-lon", math.NaN(), "Marker longitude")
	markerLat := flag.Float64("marker-lat", math.NaN(), "Marker latitude")
	markerCenter := flag.String("marker-center", "O", "Marker center character")
	markerHorizontal := flag.String("marker-horizontal", "-", "Marker horizontal character")
	markerVertical := flag.String("marker-vertical", "|", "Marker vertical character")
	markerArmX := flag.Int("marker-arm-x", -1, "Horizontal arm length in characters (-1 for full width)")
	markerArmY := flag.Int("marker-arm-y", -1, "Vertical arm length in characters (-1 for full map height)")

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

	mask, err := mapasci.LoadLandMask(*maskPath)
	if err != nil {
		return err
	}

	marker, err := buildMarker(*markerLon, *markerLat, *markerCenter, *markerHorizontal, *markerVertical, *markerArmX, *markerArmY)
	if err != nil {
		return err
	}

	asciiMap, err := mapasci.RenderWorldASCII(mask, *size, *supersample, *charAspect, marker)
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
) (*mapasci.Marker, error) {
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

	return &mapasci.Marker{
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

func resolveDefaultMaskPath() string {
	candidates := []string{
		"data/masks/landmask_3600x1800.png",
		"../data/masks/landmask_3600x1800.png",
		"../../data/masks/landmask_3600x1800.png",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return candidates[0]
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
