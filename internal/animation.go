package mapasci

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	DefaultAnimationFPS = 2
)

type AnimationStyle string

const (
	AnimationStylePulseColor AnimationStyle = "pulse-color"
	AnimationStyleBlink      AnimationStyle = "blink"
)

type AnimationOptions struct {
	FPS      int
	Style    AnimationStyle
	Duration time.Duration
}

type Frame struct {
	Text string
}

// overlayMarkerLayer finds cells that differ between base and layer,
// and applies those differences onto dst.
// base: world with no markers
// dst:  accumulated result so far (may already have markers stamped)
// layer: world with one new marker rendered
func overlayMarkerLayer(base, dst, layer string) string {
	baseRunes := []rune(base)
	dstRunes := []rune(dst)
	layerRunes := []rune(layer)

	if len(baseRunes) != len(layerRunes) || len(baseRunes) != len(dstRunes) {
		// Dimensions don't match, return dst unchanged
		return dst
	}

	out := make([]rune, len(dstRunes))
	copy(out, dstRunes)

	for i := range baseRunes {
		if layerRunes[i] != baseRunes[i] {
			out[i] = layerRunes[i]
		}
	}
	return string(out)
}

func renderAnimationFrameMulti(
	mask *LandMask,
	size int,
	supersample int,
	charAspect float64,
	markers []*Marker,
	renderOpts *RenderOptions,
	style AnimationStyle,
	frameIdx int,
	pulseA string,
	pulseB string,
) (string, error) {
	// --- Resolve options (mirrors RenderWorldASCIIWithOptions logic) ---
	verticalMarginRows := DefaultVerticalMarginRows
	frame := false
	colorMode := colorModeNever
	mapColorName := ""
	frameColorName := ""
	viewport := defaultWorldViewport()

	if renderOpts != nil {
		if renderOpts.VerticalMarginRows < 0 {
			return "", fmt.Errorf("vertical margin rows must be >= 0, got %d", renderOpts.VerticalMarginRows)
		}
		if renderOpts.VerticalPaddingRows < 0 {
			return "", fmt.Errorf("vertical padding rows must be >= 0, got %d", renderOpts.VerticalPaddingRows)
		}
		if renderOpts.VerticalMarginRows != 0 {
			verticalMarginRows = renderOpts.VerticalMarginRows
		} else {
			verticalMarginRows = renderOpts.VerticalPaddingRows
		}
		frame = renderOpts.Frame
		if renderOpts.ColorMode != "" {
			colorMode = renderOpts.ColorMode
		}
		mapColorName = renderOpts.MapColor
		frameColorName = renderOpts.FrameColor
		if renderOpts.Viewport != nil {
			viewport = *renderOpts.Viewport
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

	// --- Resolve marker color for this frame ---
	markerColorName := ""
	switch style {
	case AnimationStyleBlink:
		// color stays as-is; visibility toggled by whether we stamp markers
		if renderOpts != nil {
			markerColorName = renderOpts.MarkerColor
		}
	case AnimationStylePulseColor:
		if frameIdx%2 == 0 {
			markerColorName = pulseA
		} else {
			markerColorName = pulseB
		}
	default:
		return "", fmt.Errorf("unsupported animation style: %s", style)
	}

	markerColor, err := colorSequenceForName(markerColorName, "marker color")
	if err != nil {
		return "", err
	}

	// --- Render base grid ---
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

	// --- Stamp all markers, accumulate a single combined mask ---
	combinedMask := make([]bool, mapWidth*mapHeight)

	if style != AnimationStyleBlink || frameIdx%2 == 0 {
		for _, m := range markers {
			markerMask, err := applyMarker(lines, mapWidth, mapHeight, *m, viewport)
			if err != nil {
				return "", err
			}
			for i, v := range markerMask {
				if v {
					combinedMask[i] = true
				}
			}
		}
	}

	// --- Frame + margins ---
	if frame {
		lines = frameLines(lines, mapWidth)
	}
	if verticalMarginRows > 0 {
		lines = addVerticalMargins(lines, verticalMarginRows)
	}

	// --- Serialize ---
	useColor := colorEnabled && (mapColor != "" || frameColor != "" || markerColor != "")
	if useColor {
		return buildColoredOutput(lines, mapWidth, mapHeight, verticalMarginRows, frame, combinedMask, mapColor, frameColor, markerColor)
	}
	return buildPlainOutput(lines)
}

func StreamWorldASCIIAnimation(
	ctx context.Context,
	mask *LandMask,
	size int,
	supersample int,
	charAspect float64,
	markers []*Marker,
	renderOpts *RenderOptions,
	animOpts *AnimationOptions,
	emit func(Frame) error,
) error {
	if emit == nil {
		return fmt.Errorf("emit callback must not be nil")
	}
	if len(markers) == 0 {
		return fmt.Errorf("at least one marker must be provided")
	}
	for i, m := range markers {
		if m == nil {
			return fmt.Errorf("marker at index %d must not be nil", i)
		}
	}

	if ctx == nil {
		ctx = context.Background()
	}

	normalized, err := normalizeAnimationOptions(animOpts)
	if err != nil {
		return err
	}

	effectiveStyle := normalized.Style
	if normalized.Style == AnimationStylePulseColor {
		colorEnabled, err := colorAnimationEnabled(renderOpts)
		if err != nil {
			return err
		}
		if !colorEnabled {
			effectiveStyle = AnimationStyleBlink
		}
	}

	maxFrames := 0
	if normalized.Duration > 0 {
		maxFrames = int(math.Ceil(normalized.Duration.Seconds() * float64(normalized.FPS)))
		if maxFrames < 1 {
			maxFrames = 1
		}
	}

	ticker := time.NewTicker(time.Second / time.Duration(normalized.FPS))
	defer ticker.Stop()

	pulseA, pulseB := pulseMarkerColors(renderOpts)

	for frameIdx := 0; ; frameIdx++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Render base frame once, then composite all markers onto it
		frameText, err := renderAnimationFrameMulti(mask, size, supersample, charAspect, markers, renderOpts, effectiveStyle, frameIdx, pulseA, pulseB)
		if err != nil {
			return err
		}
		if err := emit(Frame{Text: frameText}); err != nil {
			return err
		}

		if maxFrames > 0 && frameIdx+1 >= maxFrames {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func normalizeAnimationOptions(raw *AnimationOptions) (*AnimationOptions, error) {
	normalized := &AnimationOptions{
		FPS:      DefaultAnimationFPS,
		Style:    AnimationStylePulseColor,
		Duration: 0,
	}
	if raw == nil {
		return normalized, nil
	}

	if raw.FPS < 0 {
		return nil, fmt.Errorf("animation FPS must be >= 0, got %d", raw.FPS)
	}
	if raw.FPS > 0 {
		normalized.FPS = raw.FPS
	}

	if raw.Duration < 0 {
		return nil, fmt.Errorf("animation duration must be >= 0, got %v", raw.Duration)
	}
	normalized.Duration = raw.Duration

	if raw.Style != "" {
		normalized.Style = AnimationStyle(strings.ToLower(strings.TrimSpace(string(raw.Style))))
	}

	switch normalized.Style {
	case AnimationStylePulseColor, AnimationStyleBlink:
		return normalized, nil
	default:
		return nil, fmt.Errorf("animation style must be one of: pulse-color, blink")
	}
}

func colorAnimationEnabled(options *RenderOptions) (bool, error) {
	mode := colorModeNever
	if options != nil && strings.TrimSpace(options.ColorMode) != "" {
		mode = options.ColorMode
	}

	return shouldColorize(mode)
}

func renderAnimationFrame(
	mask *LandMask,
	size int,
	supersample int,
	charAspect float64,
	marker *Marker,
	renderOpts *RenderOptions,
	style AnimationStyle,
	frameIdx int,
	pulseA string,
	pulseB string,
) (string, error) {
	switch style {
	case AnimationStyleBlink:
		if frameIdx%2 == 0 {
			return RenderWorldASCIIWithOptions(mask, size, supersample, charAspect, marker, renderOpts)
		}
		return RenderWorldASCIIWithOptions(mask, size, supersample, charAspect, nil, renderOpts)
	case AnimationStylePulseColor:
		frameOptions := cloneRenderOptions(renderOpts)
		if frameIdx%2 == 0 {
			frameOptions.MarkerColor = pulseA
		} else {
			frameOptions.MarkerColor = pulseB
		}
		return RenderWorldASCIIWithOptions(mask, size, supersample, charAspect, marker, frameOptions)
	default:
		return "", fmt.Errorf("unsupported animation style: %s", style)
	}
}

func cloneRenderOptions(options *RenderOptions) *RenderOptions {
	if options == nil {
		return &RenderOptions{}
	}

	cloned := *options
	return &cloned
}

func pulseMarkerColors(renderOpts *RenderOptions) (string, string) {
	if renderOpts == nil {
		return "red", "bright-red"
	}

	base := strings.ToLower(strings.TrimSpace(renderOpts.MarkerColor))
	if base == "" {
		return "red", "bright-red"
	}

	if strings.HasPrefix(base, "bright-") {
		normal := strings.TrimPrefix(base, "bright-")
		if _, ok := ansi16ColorCodes[normal]; ok {
			return normal, base
		}
	}

	bright := "bright-" + base
	if _, ok := ansi16ColorCodes[bright]; ok {
		return base, bright
	}

	return base, base
}
