"""HTML parsing utilities based on :mod:`selectolax`."""
from __future__ import annotations

from dataclasses import dataclass
from typing import Iterable, List, Optional, Tuple

from selectolax.parser import HTMLParser

from sfrog.utils.url_tools import is_same_domain, normalize_url


Heading = Tuple[str, str]


@dataclass(slots=True)
class ParsedPage:
    url: str
    title: Optional[str]
    meta_description: Optional[str]
    canonical: Optional[str]
    headings: List[Heading]
    internal_links: List[str]
    external_links: List[str]
    images: List[str]


def parse_html(url: str, base_url: str, html: str) -> ParsedPage:
    """Parse HTML content extracting SEO signals."""

    parser = HTMLParser(html)
    title = _first_text(parser.css("title"))
    meta_description = None
    for node in parser.css("meta"):
        if node.attributes.get("name", "").lower() == "description":
            meta_description = node.attributes.get("content")
            break
    canonical = None
    for node in parser.css("link"):
        if node.attributes.get("rel", "").lower() == "canonical":
            canonical = normalize_url(url, node.attributes.get("href", ""))
            break

    headings: List[Heading] = []
    for node in parser.css("h1, h2, h3, h4, h5, h6"):
        text = node.text(strip=True)
        if text:
            headings.append((node.tag.lower(), text))

    internal_links: List[str] = []
    external_links: List[str] = []
    for node in parser.css("a"):
        href = node.attributes.get("href")
        if not href:
            continue
        normalized = normalize_url(url, href)
        if is_same_domain(normalized, base_url):
            internal_links.append(normalized)
        else:
            external_links.append(normalized)

    images = [node.attributes.get("alt", "") for node in parser.css("img") if node.attributes.get("alt")]

    return ParsedPage(
        url=url,
        title=title,
        meta_description=meta_description,
        canonical=canonical,
        headings=headings,
        internal_links=internal_links,
        external_links=external_links,
        images=images,
    )


def _first_text(nodes: Iterable[HTMLParser]) -> Optional[str]:
    for node in nodes:
        text = node.text(strip=True)
        if text:
            return text
    return None


__all__ = ["ParsedPage", "Heading", "parse_html"]
