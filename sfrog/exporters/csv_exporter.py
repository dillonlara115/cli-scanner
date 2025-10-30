"""CSV export helpers built on top of :mod:`pandas`."""
from __future__ import annotations

from pathlib import Path
from typing import Iterable

import pandas as pd

from sfrog.crawler.manager import PageData


EXPORT_COLUMNS = [
    "url",
    "status",
    "title",
    "meta_description",
    "h1_count",
    "internal_links",
    "external_links",
    "canonical",
    "content_hash",
    "response_time",
    "redirect_target",
    "error",
]


def _page_to_record(page: PageData) -> dict:
    return {
        "url": page.url,
        "status": page.status,
        "title": page.title,
        "meta_description": page.meta_description,
        "h1_count": page.h1_count,
        "internal_links": list(dict.fromkeys(page.internal_links)),
        "external_links": list(dict.fromkeys(page.external_links)),
        "canonical": page.canonical,
        "content_hash": page.content_hash,
        "response_time": page.response_time,
        "redirect_target": page.redirect_target,
        "error": page.error,
    }


def export_to_csv(pages: Iterable[PageData], output_path: str | Path) -> Path:
    """Export crawl pages to a CSV file."""

    output = Path(output_path)
    records = [_page_to_record(page) for page in pages]
    df = pd.DataFrame(records, columns=EXPORT_COLUMNS)
    df.to_csv(output, index=False)
    return output


def export_to_excel(pages: Iterable[PageData], output_path: str | Path) -> Path:
    output = Path(output_path)
    records = [_page_to_record(page) for page in pages]
    df = pd.DataFrame(records, columns=EXPORT_COLUMNS)
    df.to_excel(output, index=False)
    return output


__all__ = ["export_to_csv", "export_to_excel", "EXPORT_COLUMNS"]
