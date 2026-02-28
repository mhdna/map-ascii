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

func StreamWorldASCIIAnimation(
	ctx context.Context,
	mask *LandMask,
	size int,
	supersample int,
	charAspect float64,
	marker *Marker,
	renderOpts *RenderOptions,
	animOpts *AnimationOptions,
	emit func(Frame) error,
) error {
	if emit == nil {
		return fmt.Errorf("emit callback must not be nil")
	}
	if marker == nil {
		return fmt.Errorf("marker must not be nil")
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

		frameText, err := renderAnimationFrame(mask, size, supersample, charAspect, marker, renderOpts, effectiveStyle, frameIdx, pulseA, pulseB)
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
