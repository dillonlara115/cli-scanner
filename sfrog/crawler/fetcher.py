"""HTTP fetching utilities built on top of :mod:`aiohttp`."""
from __future__ import annotations

import asyncio
import time
from dataclasses import dataclass
from typing import Mapping, Optional

import aiohttp


@dataclass(slots=True)
class FetchResult:
    """Represents the outcome of fetching a URL."""

    url: str
    status: Optional[int]
    headers: Mapping[str, str]
    content: Optional[bytes]
    elapsed: float
    redirected_url: Optional[str] = None
    error: Optional[str] = None


class Fetcher:
    """A thin wrapper around :class:`aiohttp.ClientSession`."""

    def __init__(self, user_agent: str, timeout: int = 20) -> None:
        self._user_agent = user_agent
        self._timeout = aiohttp.ClientTimeout(total=timeout)
        self._session: Optional[aiohttp.ClientSession] = None
        self._lock = asyncio.Lock()

    async def _ensure_session(self) -> aiohttp.ClientSession:
        async with self._lock:
            if self._session is None or self._session.closed:
                headers = {"User-Agent": self._user_agent}
                self._session = aiohttp.ClientSession(timeout=self._timeout, headers=headers)
        return self._session

    async def close(self) -> None:
        if self._session and not self._session.closed:
            await self._session.close()

    async def fetch(self, url: str) -> FetchResult:
        session = await self._ensure_session()
        start = time.perf_counter()
        try:
            async with session.get(url, allow_redirects=True) as resp:
                content = await resp.read()
                elapsed = time.perf_counter() - start
                redirected_url = str(resp.url) if str(resp.url) != url else None
                headers = {k: v for k, v in resp.headers.items()}
                return FetchResult(
                    url=url,
                    status=resp.status,
                    headers=headers,
                    content=content,
                    elapsed=elapsed,
                    redirected_url=redirected_url,
                )
        except asyncio.TimeoutError as exc:
            return FetchResult(url=url, status=None, headers={}, content=None, elapsed=time.perf_counter() - start, error="timeout")
        except aiohttp.ClientError as exc:
            return FetchResult(
                url=url,
                status=None,
                headers={},
                content=None,
                elapsed=time.perf_counter() - start,
                error=str(exc),
            )


__all__ = ["Fetcher", "FetchResult"]
