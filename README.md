# map-ascii

Go CLI for generating an ASCII world map from a pre-generated PNG land mask.

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
go run ./cmd/map-ascii --mask data/landmask_1800x900.png --size 120
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

`--marker-lon` and `--marker-lat` must be provided together.

## Flags

- `--size` map width in characters (default `60`)
- `--supersample` `N x N` supersampling per character cell (default `3`)
- `--char-aspect` character height/width ratio used to derive map height (default `2.0`)
- `--mask` path to a PNG land mask (default auto-detected `3600x1800` mask)
- `--output` optional output text file

## Notes

- Expected mask alignment:
  - x-axis maps lon `-180..180`
  - y-axis maps lat `90..-90`
- Output keeps fixed vertical padding of 2 empty rows at top and bottom.
