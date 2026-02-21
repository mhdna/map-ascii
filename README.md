# map-asci (POC)

Prototype ASCII world map generator using a real-world land/ocean mask in an equirectangular projection.

This POC uses:

- Python + `uv`
- Natural Earth land polygons via `regionmask`
- In-memory land/ocean mask by default
- `N x N` supersampling per ASCII cell

## Quick start

1. Install dependencies:

```bash
uv sync
```

2. Render an ASCII map:

```bash
uv run map-asci --size 60 --supersample 3
```

3. Render to a file:

```bash
uv run map-asci --size 120 --supersample 3 --output out/world_120.txt
```

## Commands

Render to stdout:

```bash
uv run map-asci --size 120 --supersample 3
```

Render to file:

```bash
uv run map-asci --size 120 --supersample 3 --output out/world_120.txt
```

Render with explicit aspect setting:

```bash
uv run map-asci --size 120 --supersample 3 --char-aspect 2.0 --output out/world_120.txt
```

## Notes

- Fail-fast behavior: invalid inputs, missing files, and unexpected command states raise errors directly.
- Mask generation is internal and in-memory by default.
- Internal mask resolution defaults to `3600x1800` for better coastline detail.
- No fallback data source is implemented in this POC.
- Expected mask alignment:
  - x-axis maps lon `-180..180`
  - y-axis maps lat `90..-90`
- Output keeps fixed vertical padding of 2 empty rows at top and bottom.
- `--size` is map width in characters.
- Map height is auto-derived from world aspect and `--char-aspect` (default `2.0`).
