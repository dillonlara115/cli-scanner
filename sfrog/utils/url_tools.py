"""Utilities for working with and validating URLs."""
from __future__ import annotations

import re
from urllib.parse import urljoin, urlparse, urlunparse

from yarl import URL

CANONICAL_SCHEME_RE = re.compile(r"^https?$")


def normalize_url(base: str, link: str) -> str:
    """Resolve and normalize a link against ``base``."""

    joined = urljoin(base, link)
    parsed = URL(joined)
    normalized = parsed.with_fragment("")
    return str(normalized)


def is_same_domain(url: str, base: str) -> bool:
    """Return ``True`` if ``url`` shares the host of ``base``."""

    parsed_url = urlparse(url)
    parsed_base = urlparse(base)
    return parsed_url.netloc == parsed_base.netloc


def canonicalize(url: str) -> str:
    """Return a canonical representation of the URL."""

    parsed = urlparse(url)
    scheme = parsed.scheme.lower() if CANONICAL_SCHEME_RE.match(parsed.scheme.lower()) else "http"
    netloc = parsed.netloc.lower()
    path = parsed.path or "/"
    canonical = urlunparse((scheme, netloc, path, "", parsed.query, ""))
    return canonical


def is_allowed_scheme(url: str) -> bool:
    """Return whether the URL uses an allowed HTTP(S) scheme."""

    scheme = urlparse(url).scheme
    return scheme in {"http", "https", ""}


__all__ = [
    "normalize_url",
    "is_same_domain",
    "canonicalize",
    "is_allowed_scheme",
]
