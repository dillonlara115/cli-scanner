"""Helpers for working with duplicate content detection."""
from __future__ import annotations

from typing import Dict, List

from sfrog.crawler.manager import PageData


def summarize_duplicates(duplicate_map: Dict[str, List[str]]) -> Dict[str, List[str]]:
    """Filter duplicate hashes to only those affecting multiple URLs."""

    return {hash_: urls for hash_, urls in duplicate_map.items() if len(urls) > 1}


def pages_by_hash(pages: List[PageData]) -> Dict[str, List[str]]:
    """Build a hash -> URLs mapping from page data."""

    mapping: Dict[str, List[str]] = {}
    for page in pages:
        if not page.content_hash:
            continue
        mapping.setdefault(page.content_hash, []).append(page.url)
    return summarize_duplicates(mapping)


__all__ = ["summarize_duplicates", "pages_by_hash"]
