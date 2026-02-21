# map-asci (Go)

Go port of the Python ASCII world map generator.

This implementation is self-contained under `go/` and uses pre-generated PNG masks from `../data/masks/` (or `data/masks/` when run from repo root).

## Quick start

Run from `go/`:

```bash
go run ./cmd/map-asci --size 60 --supersample 3
```

Render to a file:

```bash
go run ./cmd/map-asci --size 120 --supersample 3 --output ../out/world_120_go.txt
```

Use a specific mask:

```bash
go run ./cmd/map-asci --mask ../data/masks/landmask_1800x900.png --size 120
```

## Marker overlay

Add a crosshair marker centered on a coordinate:

```bash
go run ./cmd/map-asci --size 120 --marker-lon -73.9857 --marker-lat 40.7484
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
