"""Logging utilities built on top of Rich."""
from __future__ import annotations

from rich.console import Console
from rich.progress import Progress, SpinnerColumn, TextColumn, TimeElapsedColumn


_console: Console | None = None


def get_console() -> Console:
    """Return a singleton :class:`~rich.console.Console`."""

    global _console
    if _console is None:
        _console = Console()
    return _console


def make_progress() -> Progress:
    """Create a progress instance tailored for crawl reporting."""

    return Progress(
        SpinnerColumn(spinner_name="dots"),
        TextColumn("{task.description}"),
        TextColumn("[progress.percentage]{task.percentage:>3.0f}%"),
        TimeElapsedColumn(),
        transient=True,
        console=get_console(),
    )


__all__ = ["get_console", "make_progress"]
