from __future__ import annotations

import argparse
import math
from pathlib import Path

from map_asci.generate_mask import generate_regionmask_array
from map_asci.render import render_world_ascii


DEFAULT_MASK_WIDTH = 3600
DEFAULT_MASK_HEIGHT = 1800


def _positive_int(value: str) -> int:
    parsed = int(value)
    if parsed <= 0:
        raise argparse.ArgumentTypeError(f"Expected positive integer, got {value}")
    return parsed


def _positive_float(value: str) -> float:
    parsed = float(value)
    if not math.isfinite(parsed) or parsed <= 0.0:
        raise argparse.ArgumentTypeError(f"Expected positive float, got {value}")
    return parsed


def _build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="ASCII world map generator")
    parser.add_argument("--size", type=_positive_int, default=60)
    parser.add_argument("--supersample", type=_positive_int, default=3)
    parser.add_argument(
        "--char-aspect",
        type=_positive_float,
        default=2.0,
        help="Character height/width ratio used for output height",
    )
    parser.add_argument("--output", type=Path, help="Optional output text file")

    return parser


def _run(args: argparse.Namespace) -> None:
    mask = generate_regionmask_array(width=DEFAULT_MASK_WIDTH, height=DEFAULT_MASK_HEIGHT)

    ascii_map = render_world_ascii(
        mask=mask,
        size=args.size,
        supersample=args.supersample,
        char_aspect=args.char_aspect,
    )

    if args.output is None:
        print(ascii_map)
        return

    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(ascii_map + "\n", encoding="utf-8")
    print(f"Wrote ASCII map to {args.output}")


def main() -> None:
    parser = _build_parser()
    args = parser.parse_args()
    _run(args)
