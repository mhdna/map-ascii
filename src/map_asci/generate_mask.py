from __future__ import annotations

import importlib
from pathlib import Path

import numpy as np
from PIL import Image


def generate_regionmask_array(width: int, height: int) -> np.ndarray:
    if width <= 0:
        raise ValueError(f"width must be > 0, got {width}")
    if height <= 0:
        raise ValueError(f"height must be > 0, got {height}")

    regionmask = importlib.import_module("regionmask")

    lons = -180.0 + (np.arange(width, dtype=np.float64) + 0.5) * (360.0 / width)
    lats = 90.0 - (np.arange(height, dtype=np.float64) + 0.5) * (180.0 / height)
    lon_grid, lat_grid = np.meshgrid(lons, lats)

    land_110 = regionmask.defined_regions.natural_earth_v5_0_0.land_110
    region_idx = land_110.mask(lon_grid, lat_grid)
    land_mask = np.isfinite(region_idx.to_numpy())

    return land_mask.astype(np.float32)


def write_mask_png(mask: np.ndarray, output_path: Path) -> None:
    if output_path.suffix.lower() != ".png":
        raise ValueError(f"Output must be a .png file: {output_path}")
    if mask.ndim != 2:
        raise ValueError(f"mask must be 2D, got ndim={mask.ndim}")
    if not np.isfinite(mask).all():
        raise ValueError("mask contains non-finite values")
    if mask.min() < 0.0 or mask.max() > 1.0:
        raise ValueError("mask values must be in [0, 1]")

    output_path.parent.mkdir(parents=True, exist_ok=True)
    image = Image.fromarray(np.rint(mask * 255.0).astype(np.uint8), mode="L")
    image.save(output_path)
