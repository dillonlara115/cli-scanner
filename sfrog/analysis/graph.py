"""Graph analysis utilities using :mod:`networkx`."""
from __future__ import annotations

from typing import Iterable, Tuple

import networkx as nx


def build_link_graph(edges: Iterable[Tuple[str, str]]) -> nx.DiGraph:
    """Build and return a directed graph from crawl edges."""

    graph = nx.DiGraph()
    for source, target in edges:
        graph.add_edge(source, target)
    return graph


__all__ = ["build_link_graph"]
