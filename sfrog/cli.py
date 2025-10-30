"""Entry point for the sfrog command line interface."""
from __future__ import annotations

import asyncio
from pathlib import Path
from typing import Optional

import networkx as nx
import typer

from sfrog.analysis.duplicates import summarize_duplicates
from sfrog.analysis.graph import build_link_graph
from sfrog.analysis.seo_checks import analyze_seo
from sfrog.crawler.manager import CrawlManager, PageData
from sfrog.exporters.csv_exporter import export_to_csv
from sfrog.exporters.json_exporter import export_to_json
from sfrog.utils.config import RuntimeConfig
from sfrog.utils.logger import get_console, make_progress

app = typer.Typer(help="sfrog - a fast website crawler for SEO insights")


def _progress_callback_factory(task_id, progress, stats):
    def callback(page: PageData) -> None:
        stats["count"] += 1
        stats["total_time"] += page.response_time
        if page.status and page.status >= 400:
            stats["broken"] += 1
        avg_time = stats["total_time"] / max(stats["count"], 1)
        progress.update(
            task_id,
            advance=1,
            description=f"Crawled {stats['count']} pages | avg {avg_time:.2f}s | broken {stats['broken']}",
        )

    return callback


def _resolve_output(export_format: str, output: Optional[Path]) -> Path:
    if output:
        return output
    suffix = "csv" if export_format == "csv" else "json"
    return Path(f"results.{suffix}")


@app.command()
def crawl(
    url: str = typer.Argument(..., help="Base URL to crawl"),
    max_depth: Optional[int] = typer.Option(None, help="Maximum crawl depth"),
    threads: Optional[int] = typer.Option(None, help="Number of concurrent requests"),
    export: Optional[str] = typer.Option("csv", help="Export format (csv|json)"),
    output: Optional[Path] = typer.Option(None, help="Output file path"),
    config: Optional[Path] = typer.Option(None, help="Optional config file (JSON/YAML)"),
) -> None:
    """Crawl a website and export SEO data."""

    overrides = {"max_depth": max_depth, "threads": threads, "export": export}
    runtime = RuntimeConfig.from_sources(config_file=str(config) if config else None, overrides=overrides)
    export_format = runtime.export.lower()
    if export_format not in {"csv", "json"}:
        raise typer.BadParameter("Export format must be 'csv' or 'json'.")

    console = get_console()
    output_path = _resolve_output(export_format, output)

    async def runner() -> None:
        progress = make_progress()
        stats = {"count": 0, "total_time": 0.0, "broken": 0}
        task_id = progress.add_task("Preparing crawl", total=None)
        manager = CrawlManager(
            base_url=url,
            max_depth=runtime.max_depth,
            threads=runtime.threads,
            user_agent=runtime.user_agent,
            timeout=runtime.timeout,
            progress_callback=_progress_callback_factory(task_id, progress, stats),
        )
        with progress:
            result = await manager.crawl()

        graph = build_link_graph(result.edges)
        graph_path = Path("link_graph.gexf")
        nx.write_gexf(graph, graph_path)

        if export_format == "csv":
            export_to_csv(result.pages, output_path)
        else:
            export_to_json(result.pages, output_path)

        seo_issues = analyze_seo(result.pages)
        duplicates = summarize_duplicates(result.duplicate_map)

        console.rule("Crawl summary")
        console.print(f"Total pages: {len(result.pages)}")
        console.print(f"Broken pages: {len(result.broken_links)}")
        console.print(f"Duplicates: {len(duplicates)} groups")
        console.print(f"Exported data -> [bold]{output_path}[/bold]")
        console.print(f"Link graph -> [bold]{graph_path}[/bold]")
        console.print("SEO issues:")
        console.print(seo_issues.to_dict())

    asyncio.run(runner())


if __name__ == "__main__":
    app()
