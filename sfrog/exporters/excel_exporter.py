"""Excel export helpers."""
from __future__ import annotations

from pathlib import Path
from typing import Iterable

from sfrog.crawler.manager import PageData
from sfrog.exporters.csv_exporter import export_to_excel


def export(pages: Iterable[PageData], output_path: str | Path) -> Path:
    return export_to_excel(pages, output_path)


__all__ = ["export"]
