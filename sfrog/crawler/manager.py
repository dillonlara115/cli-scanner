"""High level crawling coordination."""
from __future__ import annotations

import asyncio
import hashlib
from dataclasses import dataclass
from typing import Callable, Dict, List, Optional, Set, Tuple

from yarl import URL

from sfrog.crawler.fetcher import FetchResult, Fetcher
from sfrog.crawler.parser import ParsedPage, parse_html
from sfrog.crawler.robots import RobotsRules, build_robots
from sfrog.crawler.sitemap import parse_sitemap
from sfrog.utils.url_tools import canonicalize, is_allowed_scheme, is_same_domain


@dataclass(slots=True)
class PageData:
    url: str
    status: Optional[int]
    title: Optional[str]
    meta_description: Optional[str]
    canonical: Optional[str]
    headings: List[Tuple[str, str]]
    internal_links: List[str]
    external_links: List[str]
    h1_count: int
    content_hash: Optional[str]
    response_time: float
    redirect_target: Optional[str]
    error: Optional[str]


@dataclass
class CrawlResult:
    pages: List[PageData]
    edges: List[Tuple[str, str]]
    duplicate_map: Dict[str, List[str]]
    broken_links: List[PageData]


class CrawlManager:
    """Coordinate the crawling pipeline."""

    def __init__(
        self,
        base_url: str,
        max_depth: int,
        threads: int,
        user_agent: str,
        timeout: int,
        progress_callback: Callable[[PageData], None] | None = None,
    ) -> None:
        self.base_url = canonicalize(base_url)
        self.max_depth = max_depth
        self.threads = threads
        self.fetcher = Fetcher(user_agent=user_agent, timeout=timeout)
        self.user_agent = user_agent
        self.queue: asyncio.Queue[Tuple[str, int, Optional[str]]] = asyncio.Queue()
        self.visited: Set[str] = set()
        self.enqueued: Set[str] = set()
        self.edges: List[Tuple[str, str]] = []
        self.pages: List[PageData] = []
        self.duplicate_map: Dict[str, List[str]] = {}
        self.broken_links: List[PageData] = []
        self.robots: RobotsRules | None = None
        self.sem = asyncio.Semaphore(threads)
        self._progress_callback = progress_callback

    async def crawl(self) -> CrawlResult:
        await self._prepare_seed_urls()

        workers = [asyncio.create_task(self._worker(i)) for i in range(self.threads)]
        await self.queue.join()
        for worker in workers:
            worker.cancel()
        await asyncio.gather(*workers, return_exceptions=True)
        await self.fetcher.close()
        return CrawlResult(
            pages=self.pages,
            edges=self.edges,
            duplicate_map={k: v for k, v in self.duplicate_map.items() if len(v) > 1},
            broken_links=self.broken_links,
        )

    async def _prepare_seed_urls(self) -> None:
        await self.queue.put((self.base_url, 0, None))
        self.enqueued.add(self.base_url)

        robots_url = str(URL(self.base_url).with_path("/robots.txt"))
        robots_result = await self.fetcher.fetch(robots_url)
        text = robots_result.content.decode("utf-8", errors="ignore") if robots_result.content else None
        self.robots = build_robots(self.user_agent, text)

        sitemap_url = str(URL(self.base_url).with_path("/sitemap.xml"))
        sitemap_result = await self.fetcher.fetch(sitemap_url)
        if sitemap_result.status and sitemap_result.status < 400 and sitemap_result.content:
            try:
                urls = await parse_sitemap(self.base_url, sitemap_result.content.decode("utf-8", errors="ignore"))
            except Exception:
                urls = []
            for url in urls:
                if is_same_domain(url, self.base_url) and url not in self.enqueued:
                    await self.queue.put((url, 0, None))
                    self.enqueued.add(url)

    async def _worker(self, worker_id: int) -> None:
        while True:
            try:
                url, depth, source = await self.queue.get()
            except asyncio.CancelledError:
                return
            try:
                await self._process_url(url, depth, source)
            finally:
                self.queue.task_done()

    async def _process_url(self, url: str, depth: int, source: Optional[str]) -> None:
        if url in self.visited:
            return
        self.visited.add(url)

        if self.robots and not self.robots.allows(url):
            return

        if depth > self.max_depth:
            return

        if source:
            self.edges.append((source, url))

        async with self.sem:
            result = await self.fetcher.fetch(url)

        content_hash: Optional[str] = None
        parsed: Optional[ParsedPage] = None

        if result.content:
            content_hash = hashlib.md5(result.content).hexdigest()
            parsed = self._maybe_parse_html(url, result)
            if content_hash:
                bucket = self.duplicate_map.setdefault(content_hash, [])
                bucket.append(url)

        page = self._build_page(result, parsed, content_hash)
        self.pages.append(page)
        if page.status and page.status >= 400:
            self.broken_links.append(page)

        if self._progress_callback:
            self._progress_callback(page)

        if not parsed:
            return

        for link in parsed.internal_links:
            if not is_allowed_scheme(link):
                continue
            if link not in self.visited and link not in self.enqueued and is_same_domain(link, self.base_url):
                await self.queue.put((link, depth + 1, url))
                self.enqueued.add(link)

    def _maybe_parse_html(self, url: str, result: FetchResult) -> Optional[ParsedPage]:
        content_type = result.headers.get("Content-Type", "").split(";")[0].strip().lower()
        if content_type and "html" not in content_type:
            return None
        if not result.content:
            return None
        try:
            html = result.content.decode("utf-8", errors="ignore")
            return parse_html(url, self.base_url, html)
        except Exception:
            return None

    def _build_page(
        self,
        result: FetchResult,
        parsed: Optional[ParsedPage],
        content_hash: Optional[str],
    ) -> PageData:
        headings = parsed.headings if parsed else []
        h1_count = sum(1 for tag, _ in headings if tag == "h1")
        return PageData(
            url=result.url,
            status=result.status,
            title=parsed.title if parsed else None,
            meta_description=parsed.meta_description if parsed else None,
            canonical=parsed.canonical if parsed else None,
            headings=headings,
            internal_links=parsed.internal_links if parsed else [],
            external_links=parsed.external_links if parsed else [],
            h1_count=h1_count,
            content_hash=content_hash,
            response_time=result.elapsed,
            redirect_target=result.redirected_url,
            error=result.error,
        )


__all__ = ["CrawlManager", "CrawlResult", "PageData"]
