from pathlib import Path
import sys
from importlib import import_module

sys.path.insert(0, str(Path(__file__).parent / "src"))


def main() -> None:
    cli = import_module("map_asci.cli")
    cli.main()


if __name__ == "__main__":
    main()
