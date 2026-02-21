from __future__ import annotations

import numpy as np

from map_asci.mask import sample_land_value


def char_for_land_fraction(fraction: float) -> str:
    if not np.isfinite(fraction):
        raise ValueError("Land fraction must be finite")
    if fraction < 0.0 or fraction > 1.0:
        raise ValueError(f"Land fraction must be in [0, 1], got {fraction}")

    if fraction < 0.12:
        return " "
    if fraction < 0.38:
        return "."
    if fraction < 0.62:
        return "*"
    if fraction < 0.86:
        return "@"
    return "#"


def render_world_ascii(
    mask: np.ndarray, size: int, supersample: int, char_aspect: float
) -> str:
    if mask.ndim != 2:
        raise ValueError(f"mask must be 2D, got ndim={mask.ndim}")
    if mask.shape[0] < 2 or mask.shape[1] < 2:
        raise ValueError(f"mask must be at least 2x2, got {mask.shape[1]}x{mask.shape[0]}")
    if not np.isfinite(mask).all():
        raise ValueError("mask contains non-finite values")
    if mask.min() < 0.0 or mask.max() > 1.0:
        raise ValueError("mask values must be in [0, 1]")

    if size <= 0:
        raise ValueError(f"size must be > 0, got {size}")
    if supersample <= 0:
        raise ValueError(f"supersample must be > 0, got {supersample}")
    if not np.isfinite(char_aspect) or char_aspect <= 0.0:
        raise ValueError(f"char_aspect must be > 0, got {char_aspect}")

    map_width = size
    map_height = int(round(map_width / (2.0 * char_aspect)))
    if map_height <= 0:
        raise ValueError(
            f"size={size} with char_aspect={char_aspect} produces zero map height"
        )

    subsamples_per_cell = supersample * supersample
    vertical_padding_rows = 2
    total_rows = map_height + (2 * vertical_padding_rows)
    lines: list[str] = []

    for row in range(total_rows):
        if row < vertical_padding_rows or row >= total_rows - vertical_padding_rows:
            lines.append(" " * map_width)
            continue

        row_chars: list[str] = []
        for col in range(map_width):
            land_sum = 0.0
            for sy in range(supersample):
                for sx in range(supersample):
                    x = col + (sx + 0.5) / supersample
                    y = row + (sy + 0.5) / supersample

                    lon = (x / map_width) * 360.0 - 180.0
                    y_active = y - vertical_padding_rows
                    t = y_active / map_height
                    lat = 90.0 - (180.0 * t)

                    land_sum += sample_land_value(mask, lon=lon, lat=lat)

            land_fraction = land_sum / subsamples_per_cell
            row_chars.append(char_for_land_fraction(land_fraction))

        lines.append("".join(row_chars))

    return "\n".join(lines)
