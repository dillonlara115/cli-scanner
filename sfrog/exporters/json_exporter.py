"""JSON export helpers."""
from __future__ import annotations

import json
from pathlib import Path
from typing import Iterable

from sfrog.crawler.manager import PageData
from sfrog.exporters.csv_exporter import _page_to_record


def export_to_json(pages: Iterable[PageData], output_path: str | Path) -> Path:
    output = Path(output_path)
    records = [_page_to_record(page) for page in pages]
    output.write_text(json.dumps(records, indent=2), encoding="utf-8")
    return output


__all__ = ["export_to_json"]
