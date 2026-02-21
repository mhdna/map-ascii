---
title: "ASCII World Map Generator with Land/Ocean Mask (Python)"
depth: "detailed"
scope_type: "implementation-ready"
redaction: "not_needed"
source: "conversation transcript"
---

---

# 1) Snapshot (read this first)

- Outcome: Agreed approach to build a Python-based ASCII world map generator (X×X monospace grid) using a real-world land/ocean mask raster (Option 2A) with supersampling for high-quality coastlines.
- Purpose: Generate visually accurate ASCII maps by mapping real-world coordinates (lat/lon) into grid cells and deciding land vs ocean via a precomputed raster mask.
- Success:
  - Correct lat/lon → grid projection (Plate Carrée / equirectangular).
  - Land/ocean determined via real dataset (not procedural noise).
  - Coastlines look acceptable at low resolutions via supersampling and density characters.

- Current stage: Planning (implementation-ready direction defined, no production code yet).

---

# 2) Context & motivation (why this came up)

- Starting point: Desire to create an ASCII map generator of size X×X monospace characters.
- Core technical idea: Introduce a translation layer from real-world coordinates to character grid coordinates and decide per-cell if it is land or ocean.
- Key constraint: Must reflect real-world geography (not procedural generation).
- Final direction emerged after evaluating:
  - Polygon point-in-polygon coastline checks (accurate but heavier).
  - Precomputed raster land/sea mask (simpler sampling, preferred).
  - Procedural noise (rejected; not real-world).

- Chosen solution: Use a global equirectangular land/ocean mask raster and sample it per ASCII cell (with supersampling for quality).

---

# 3) Final agreements

1. Use Python as the implementation language.
2. Use Option 2A: a precomputed raster land/sea mask (global, equirectangular).
3. Use Plate Carrée (equirectangular) projection.
4. Longitude normalization formula:
   `u = (lon + 180) / 360`
5. For rendering:
   - For each ASCII cell, convert cell center (or sub-samples) to lat/lon.
   - Sample the land mask.
   - Compute land fraction via supersampling.

6. Use multiple ASCII characters to anti-alias coastlines.
7. Two acceptable mask sources:
   - Generate mask via `regionmask` (Natural Earth polygons).
   - Use NASA IMERG land/sea mask (0.1° raster).

---

# 4) Decisions & rationale (decision log)

## Decision 1: Use raster land/sea mask (Option 2A)

- Decision: Land/ocean classification will be done by sampling a precomputed raster mask.
- Rationale:
  - Simple array indexing.
  - Faster and easier than polygon point-in-polygon.
  - Clean integration with supersampling.

- Alternatives considered:
  - Polygon point-in-polygon (Natural Earth / GSHHG).
  - Procedural continent generation.

- Why not:
  - Polygon: heavier implementation complexity.
  - Procedural: not real-world.

- Implications / follow-ups:
  - Must obtain or generate a global equirectangular mask.
  - Must ensure mask orientation matches lat/lon conventions.

---

## Decision 2: Use equirectangular (Plate Carrée) projection

- Decision: Use linear lon/lat mapping:
  - `u = (lon + 180) / 360`
  - `v = (90 - lat) / 180`

- Rationale:
  - Simplest projection.
  - Matches most global raster land masks.
  - No need for Mercator distortion handling.

- Alternatives considered:
  - Web Mercator.

- Why not:
  - More distortion near poles.
  - Unnecessary complexity.

- Implications:
  - Map will distort areas near poles (accepted tradeoff).
  - ASCII map forced to X×X may appear vertically stretched.

---

## Decision 3: Supersampling per ASCII cell

- Decision: Use N×N supersampling (e.g., 3×3) per ASCII cell.
- Rationale:
  - Significantly improves coastline quality.
  - Reduces aliasing artifacts at low resolutions.

- Alternatives considered:
  - Single sample at cell center.

- Why not:
  - Produces jagged coastlines.

- Implications:
  - Slightly higher compute cost.
  - Land fraction becomes continuous (0..1), enabling density-based characters.

---

## Decision 4: Use multi-character land density

- Decision: Use thresholds on land fraction to select characters:
  - `" "` (ocean)
  - `"."`
  - `":"`
  - `"X"`
  - `"#"` (solid land)

- Rationale:
  - Anti-aliased coastlines look better.

- Alternatives considered:
  - Binary land/water.

- Why not:
  - Poor coastline quality.

- Implications:
  - Output becomes visually richer.
  - Thresholds are tunable.

---

## Decision 5: Mask generation options

Two valid implementation paths:

### Option A: Generate PNG mask using `regionmask`

- Uses Natural Earth land polygons.
- Outputs grayscale PNG (255=land, 0=water).
- One-time generation and reuse.

### Option B: Use NASA IMERG land/sea mask (NetCDF)

- Already a 0.1° global raster.
- Contains percent water per cell.
- Convert to land fraction and optionally export PNG.

Rationale:

- Both provide real-world classification.
- Regionmask easier for PNG-based workflow.
- NASA mask avoids polygon processing entirely.

---

# 5) Scope boundaries

## In scope

- Global world map (lon -180..180, lat -90..90).
- X×X ASCII output.
- Real-world geography.
- Supersampling.
- Density-based coastline rendering.

## Out of scope / explicitly not doing

- Procedural continent generation.
- Advanced projections (Robinson, Winkel Tripel).
- Interactive UI.
- Map labeling or graticules.
- Performance optimization beyond reasonable supersampling.

---

# 6) Open questions / TBDs

## TBDs

1. Which mask source will be used?
   - `regionmask`-generated PNG?
   - NASA IMERG NetCDF?

2. Desired default ASCII resolution (X)?
3. Desired supersampling factor (default 3×3?).
4. Should aspect ratio distortion be compensated?

## What’s needed to resolve

- Confirm mask source.
- Confirm expected map size range (e.g., 40×40 vs 200×200).
- Confirm performance constraints.

---

# 7) Next steps (actionable)

1. Choose mask source (STOP if not decided).
2. If regionmask:
   - Install dependencies.
   - Generate `landmask_0_255.png`.

3. Implement ASCII renderer:
   - Lat/lon inversion per cell.
   - Supersampling loop.
   - Land fraction computation.
   - Character threshold mapping.

4. Validate:
   - Render at low resolution (e.g., 60×60).
   - Render at medium resolution (120×120).
   - Inspect coast quality.

5. Tune:
   - Adjust thresholds.
   - Adjust supersampling factor.

6. Optional:
   - Add bounding box support.
   - Add configurable character palette.

---

# Constraints & assumptions

- Mask raster must be equirectangular and aligned:
  - x-axis: lon -180..180
  - y-axis: lat 90..-90

- Mask values must clearly distinguish land vs water.
- Longitude wrapping must be handled (u % 1.0).
- Latitude must be clamped.

---

# Rejected ideas recap

- Polygon-based coastline checks (accurate but heavier).
- Procedural noise continents (not real-world).
- Web Mercator projection (unnecessary distortion for this use case).
