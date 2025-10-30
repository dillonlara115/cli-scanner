"""Configuration loading utilities for sfrog.

This module provides helpers to merge CLI arguments with values defined in an
optional configuration file. Both JSON and YAML formats are supported.
"""
from __future__ import annotations

import json
import pathlib
from dataclasses import dataclass, field
from typing import Any, Dict, Optional

try:
    import yaml  # type: ignore
except ImportError:  # pragma: no cover - optional dependency
    yaml = None


DEFAULT_CONFIG = {
    "max_depth": 2,
    "threads": 10,
    "user_agent": "sfrog/1.0 (+https://github.com/openai)",
    "export": "csv",
    "timeout": 20,
}


@dataclass
class RuntimeConfig:
    """Holds runtime configuration for the crawler."""

    max_depth: int = DEFAULT_CONFIG["max_depth"]
    threads: int = DEFAULT_CONFIG["threads"]
    user_agent: str = DEFAULT_CONFIG["user_agent"]
    export: str = DEFAULT_CONFIG["export"]
    timeout: int = DEFAULT_CONFIG["timeout"]
    config_path: Optional[pathlib.Path] = None
    extra: Dict[str, Any] = field(default_factory=dict)

    @classmethod
    def from_sources(
        cls,
        config_file: Optional[str] = None,
        overrides: Optional[Dict[str, Any]] = None,
    ) -> "RuntimeConfig":
        """Build a :class:`RuntimeConfig` from file and CLI overrides."""

        data: Dict[str, Any] = DEFAULT_CONFIG.copy()
        path: Optional[pathlib.Path] = None

        if config_file:
            path = pathlib.Path(config_file)
            file_data = load_config_file(path)
            data.update(file_data)

        if overrides:
            data.update({k: v for k, v in overrides.items() if v is not None})

        config = cls(
            max_depth=int(data.get("max_depth", DEFAULT_CONFIG["max_depth"])),
            threads=int(data.get("threads", DEFAULT_CONFIG["threads"])),
            user_agent=str(data.get("user_agent", DEFAULT_CONFIG["user_agent"])),
            export=str(data.get("export", DEFAULT_CONFIG["export"])),
            timeout=int(data.get("timeout", DEFAULT_CONFIG["timeout"])),
            config_path=path,
            extra={k: v for k, v in data.items() if k not in DEFAULT_CONFIG},
        )
        return config


def load_config_file(path: pathlib.Path) -> Dict[str, Any]:
    """Load configuration data from JSON or YAML files."""

    if not path.exists():
        raise FileNotFoundError(f"Config file not found: {path}")

    text = path.read_text(encoding="utf-8")
    if path.suffix.lower() in {".yaml", ".yml"}:
        if yaml is None:
            raise RuntimeError("PyYAML is required to read YAML configuration files.")
        data = yaml.safe_load(text) or {}
    else:
        data = json.loads(text or "{}")

    if not isinstance(data, dict):
        raise ValueError("Configuration file must contain a JSON/YAML object.")

    return data


__all__ = ["RuntimeConfig", "load_config_file", "DEFAULT_CONFIG"]
