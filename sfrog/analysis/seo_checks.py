"""SEO validation helpers."""
from __future__ import annotations

from collections import Counter
from typing import Dict, Iterable, List

from sfrog.crawler.manager import PageData


class SEOIssues:
    """Container for common SEO issues discovered during crawl."""

    def __init__(self) -> None:
        self.missing_titles: List[str] = []
        self.missing_meta: List[str] = []
        self.duplicate_titles: Dict[str, List[str]] = {}
        self.duplicate_meta: Dict[str, List[str]] = {}
        self.multiple_h1: List[str] = []
        self.invalid_canonicals: List[str] = []

    def to_dict(self) -> Dict[str, object]:
        return {
            "missing_titles": self.missing_titles,
            "missing_meta": self.missing_meta,
            "duplicate_titles": self.duplicate_titles,
            "duplicate_meta": self.duplicate_meta,
            "multiple_h1": self.multiple_h1,
            "invalid_canonicals": self.invalid_canonicals,
        }


def analyze_seo(pages: Iterable[PageData]) -> SEOIssues:
    pages_list = list(pages)
    issues = SEOIssues()
    title_counter: Counter[str] = Counter()
    meta_counter: Counter[str] = Counter()

    canonical_seen: set[str] = set()

    for page in pages_list:
        if not page.title:
            issues.missing_titles.append(page.url)
        else:
            title_counter[page.title] += 1

        if not page.meta_description:
            issues.missing_meta.append(page.url)
        else:
            meta_counter[page.meta_description] += 1

        if page.h1_count > 1:
            issues.multiple_h1.append(page.url)

        if page.canonical:
            if page.canonical in canonical_seen and page.canonical != page.url:
                issues.invalid_canonicals.append(page.url)
            canonical_seen.add(page.canonical)

    issues.duplicate_titles = {
        title: [page.url for page in pages_list if page.title == title]
        for title, count in title_counter.items()
        if count > 1
    }

    issues.duplicate_meta = {
        meta: [page.url for page in pages_list if page.meta_description == meta]
        for meta, count in meta_counter.items()
        if count > 1
    }

    return issues


__all__ = ["analyze_seo", "SEOIssues"]
