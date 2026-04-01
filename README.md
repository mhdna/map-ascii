# map-ascii

Go CLI for generating an ASCII world map from a pre-generated PNG land mask.

Module path: `github.com/mhdna/map-ascii`

## Install CLI

```bash
go install github.com/mhdna/map-ascii/cmd/map-ascii@latest
```

Then run:

```bash
map-ascii --size 60 --supersample 3
```

## Library usage

```go
package main

import (
	"fmt"
	"log"

	mapascii "github.com/mhdna/map-ascii"
)

func main() {
	mask, err := mapascii.LoadEmbeddedDefaultLandMask()
	if err != nil {
		log.Fatal(err)
	}

	out, err := mapascii.RenderWorldASCII(mask, 80, 3, 2.0, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(out)
}
```

You can also load a PNG mask from disk with `mapascii.LoadLandMask("path/to/mask.png")`.

For custom margin and optional framing:

```go
options := &mapascii.RenderOptions{
	VerticalMarginRows: 2,
	Frame:              true,
}

out, err := mapascii.RenderWorldASCIIWithOptions(mask, 80, 3, 2.0, nil, options)
```

Render a geographic subset with a viewport bounding box:

```go
options := &mapascii.RenderOptions{
	Viewport: &mapascii.Viewport{
		MinLon: -20,
		MinLat: -36,
		MaxLon: 53,
		MaxLat: 38,
	},
}

out, err := mapascii.RenderWorldASCIIWithOptions(mask, 80, 3, 2.0, nil, options)
```

Use a built-in continent preset (library API):

```go
viewport, err := mapascii.ViewportForContinent(string(mapascii.ContinentAfrica))
if err != nil {
	log.Fatal(err)
}

options := &mapascii.RenderOptions{Viewport: &viewport}
out, err := mapascii.RenderWorldASCIIWithOptions(mask, 80, 3, 2.0, nil, options)
```

List accepted continent names at runtime:

```go
fmt.Println(mapascii.ContinentNames())
// [africa antarctica asia europe north-america south-america oceania]
```

Stream animated frames with a callback (import `context` in your program):

```go
animOptions := &mapascii.AnimationOptions{
	FPS:   2,
	Style: mapascii.AnimationStylePulseColor,
}

err = mapascii.StreamWorldASCIIAnimation(
	context.Background(),
	mask,
	80,
	3,
	2.0,
	&mapascii.Marker{Lon: -73.9857, Lat: 40.7484},
	&mapascii.RenderOptions{ColorMode: "always"},
	animOptions,
	func(frame mapascii.Frame) error {
		fmt.Print(frame.Text)
		return nil
	},
)
```

## Quick start

Render to stdout:

```bash
go run ./cmd/map-ascii --size 60 --supersample 3
```

Render to a file:

```bash
go run ./cmd/map-ascii --size 120 --supersample 3 --output out/world_120.txt
```

Use a specific mask:

```bash
go run ./cmd/map-ascii --mask data/landmask_3600x1800.png --size 120
```

Render a continent preset:

```bash
go run ./cmd/map-ascii --size 90 --continent africa
```

Render a custom bounding box:

```bash
go run ./cmd/map-ascii --size 90 --bbox=-20,-36,53,38
```

Render with no top/bottom margin and a frame:

```bash
go run ./cmd/map-ascii --size 80 --margin-y 0 --frame
```

Render with forced terminal colors:

```bash
go run ./cmd/map-ascii --size 80 --frame --color always --map-color green --frame-color bright-white --marker-lon -73.9857 --marker-lat 40.7484 --marker-color bright-red
```

Animate a marker in the terminal (Ctrl+C to stop):

```bash
go run ./cmd/map-ascii --size 80 --frame --color auto --marker-lon -73.9857 --marker-lat 40.7484 --animate-marker --animate-style pulse-color --animate-fps 2
```

## Marker overlay

Add a crosshair marker centered on a coordinate:

```bash
go run ./cmd/map-ascii --size 120 --marker-lon -73.9857 --marker-lat 40.7484
```

Marker options:

- `--marker-center` (default `O`)
- `--marker-horizontal` (default `-`)
- `--marker-vertical` (default `|`)
- `--marker-arm-x` (default `-1`, full width)
- `--marker-arm-y` (default `-1`, full map height)
- `--marker-color` marker color name (16 ANSI colors)
- `--animate-marker` animate marker output (TTY stdout only)
- `--animate-fps` animation refresh rate in frames per second (default `2`)
- `--animate-style` animation style: `pulse-color`, `blink` (default `pulse-color`)
- `--animate-duration` animation duration (for example `10s`, `2m`); `0` runs until interrupted

`--marker-lon` and `--marker-lat` must be provided together.

`--animate-marker` requires marker coordinates and cannot be used with `--output`.

## Color support

Color output targets ANSI 16-color terminals.

- `--color` controls color mode: `never`, `auto`, `always` (default `auto`)
- `--map-color` sets the map color
- `--frame-color` sets the frame color
- `--marker-color` sets the marker color

When `--color auto` is used, color output is enabled only on terminals that look color-capable. If `NO_COLOR` is set, color is disabled.

Supported color names:

- `black`
- `red`
- `green`
- `yellow`
- `blue`
- `magenta`
- `cyan`
- `white`
- `bright-black`
- `bright-red`
- `bright-green`
- `bright-yellow`
- `bright-blue`
- `bright-magenta`
- `bright-cyan`
- `bright-white`

## Flags

- `--size` map width in characters (default `60`)
- `--supersample` `N x N` supersampling per character cell (default `3`)
- `--char-aspect` character height/width ratio used to derive map height (default `2.0`)
- `--margin-y` empty rows above and below the map (outside the frame) (default `2`)
- `--frame` draw an ASCII frame around the output (default `false`)
- `--color` color output mode: `never`, `auto`, `always` (default `auto`)
- `--map-color` map color name (16 ANSI colors)
- `--frame-color` frame color name (16 ANSI colors)
- `--marker-color` marker color name (16 ANSI colors)
- `--animate-marker` animate marker output (TTY stdout only)
- `--animate-fps` animation refresh rate in frames per second (default `2`)
- `--animate-style` animation style: `pulse-color`, `blink` (default `pulse-color`)
- `--animate-duration` animation duration (`0` means until interrupted)
- `--mask` path to a PNG land mask (optional; defaults to local `data/landmask_3600x1800.png` with embedded fallback)
- `--continent` preset viewport: `africa`, `antarctica`, `asia`, `europe`, `north-america`, `south-america`, `oceania`
- `--bbox` viewport bounding box as `minLon,minLat,maxLon,maxLat`
- `--output` optional output text file

## Notes

- Expected mask alignment:
  - x-axis maps lon `-180..180`
  - y-axis maps lat `90..-90`
- `--continent` and `--bbox` are mutually exclusive.
- `--continent australia` is accepted as an alias of `oceania`.
- By default output has 2 empty rows of top/bottom margin; use `--margin-y` to change it.
- Use `--frame` to wrap the output in `+---+` and `|   |` style borders.
