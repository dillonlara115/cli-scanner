"""Utilities for sitemap discovery and parsing."""
from __future__ import annotations

from typing import Iterable, List

import xmltodict

from sfrog.utils.url_tools import normalize_url


async def parse_sitemap(base_url: str, text: str) -> List[str]:
    """Parse sitemap XML content returning discovered URLs."""

    data = xmltodict.parse(text)
    urls: List[str] = []

    if "urlset" in data:
        entries = data["urlset"].get("url", [])
        if isinstance(entries, dict):
            entries = [entries]
        for entry in entries:
            loc = entry.get("loc")
            if loc:
                urls.append(normalize_url(base_url, loc))
    elif "sitemapindex" in data:
        sitemaps = data["sitemapindex"].get("sitemap", [])
        if isinstance(sitemaps, dict):
            sitemaps = [sitemaps]
        for item in sitemaps:
            loc = item.get("loc")
            if loc:
                urls.append(normalize_url(base_url, loc))
    return urls


__all__ = ["parse_sitemap"]
