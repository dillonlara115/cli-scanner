"""Robots.txt handling utilities."""
from __future__ import annotations

from dataclasses import dataclass

try:  # pragma: no cover - optional dependency wrapper
    from robots_txt_parser import Robots  # type: ignore
except ImportError:  # pragma: no cover
    Robots = None  # type: ignore


@dataclass(slots=True)
class RobotsRules:
    user_agent: str
    robots: object | None

    def allows(self, url: str) -> bool:
        if Robots is None or self.robots is None:
            return True
        try:
            return bool(self.robots.allowed(url, self.user_agent))
        except Exception:  # pragma: no cover - defensive
            return True


def build_robots(user_agent: str, text: str | None) -> RobotsRules:
    """Create :class:`RobotsRules` from robots.txt content."""

    if text and Robots is not None:
        try:
            robots = Robots.parse(text)
        except Exception:  # pragma: no cover - robustness
            robots = None
    else:
        robots = None
    return RobotsRules(user_agent=user_agent, robots=robots)


__all__ = ["RobotsRules", "build_robots"]
