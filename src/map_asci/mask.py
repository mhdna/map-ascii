from __future__ import annotations

from pathlib import Path

import numpy as np
from PIL import Image


def load_land_mask(mask_path: Path) -> np.ndarray:
    if not mask_path.exists():
        raise FileNotFoundError(f"Mask file does not exist: {mask_path}")
    if mask_path.suffix.lower() != ".png":
        raise ValueError(f"Mask file must be a PNG: {mask_path}")

    with Image.open(mask_path) as image:
        mask = np.asarray(image.convert("L"), dtype=np.float32) / 255.0

    if mask.ndim != 2:
        raise ValueError(f"Mask must be a 2D grayscale image, got ndim={mask.ndim}")
    if mask.shape[0] < 2 or mask.shape[1] < 2:
        raise ValueError(
            f"Mask must be at least 2x2 pixels, got {mask.shape[1]}x{mask.shape[0]}"
        )
    if not np.isfinite(mask).all():
        raise ValueError("Mask contains non-finite values")

    return mask


def sample_land_value(mask: np.ndarray, lon: float, lat: float) -> float:
    if mask.ndim != 2:
        raise ValueError(f"Mask must be 2D, got ndim={mask.ndim}")
    if not np.isfinite(lon) or not np.isfinite(lat):
        raise ValueError("lon and lat must be finite")

    height, width = mask.shape
    u = ((lon + 180.0) / 360.0) % 1.0
    v = (90.0 - lat) / 180.0
    v = float(np.clip(v, 0.0, 1.0))

    x = min(int(u * width), width - 1)
    y = min(int(v * height), height - 1)
    return float(mask[y, x])
