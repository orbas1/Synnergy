#!/usr/bin/env python3
"""Inventory audit script for Synnergy project.

Parses AGENTS.md for documented files and compares against actual files under
`synnergy-network/core`, `synnergy-network/GUI` and `synnergy-network/cmd`.

The generated report lists missing and undocumented files with assigned owners
and priority tags and summarizes module dependencies from Go and Node manifests.
"""
from __future__ import annotations

import argparse
import json
import os
import re
from collections import defaultdict
from pathlib import Path
from typing import Dict, Iterable, List, Set, Tuple

REPO_ROOT = Path(__file__).resolve().parent.parent
AGENTS_FILE = REPO_ROOT / "AGENTS.md"

# Match repository paths without trailing punctuation or spaces
FILE_PATTERN = re.compile(r"synnergy-network/[\w./-]+")
TARGET_DIRS = [
    REPO_ROOT / "synnergy-network" / "core",
    REPO_ROOT / "synnergy-network" / "GUI",
    REPO_ROOT / "synnergy-network" / "cmd",
]

OWNER_MAP = {
    "core": "Core Team",
    "GUI": "GUI Team",
    "cmd": "CLI Team",
}


def parse_documented_files() -> Set[str]:
    """Return set of file paths documented in AGENTS.md."""
    files: Set[str] = set()
    with AGENTS_FILE.open("r", encoding="utf-8") as fh:
        for line in fh:
            for match in FILE_PATTERN.findall(line):
                if match in {
                    "synnergy-network/core",
                    "synnergy-network/GUI",
                    "synnergy-network/cmd",
                }:
                    continue
                files.add(match)
    return files


def list_actual_files() -> Set[str]:
    """Return set of actual file paths under target directories."""
    files: Set[str] = set()
    for base in TARGET_DIRS:
        if not base.exists():
            continue
        for path in base.rglob("*"):
            if path.is_file():
                rel = path.relative_to(REPO_ROOT).as_posix()
                if Path(rel).name == ".DS_Store":
                    continue
                files.add(rel)
    return files


def parse_go_mod(path: Path) -> List[str]:
    deps: List[str] = []
    in_require = False
    for line in path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if line.startswith("require ("):
            in_require = True
            continue
        if in_require and line == ")":
            in_require = False
            continue
        if in_require and line:
            deps.append(line)
        elif line.startswith("require ") and not line.endswith("("):
            deps.append(line.split("require ", 1)[1])
    return deps


def parse_package_json(path: Path) -> List[str]:
    deps: List[str] = []
    data = json.loads(path.read_text(encoding="utf-8"))
    for section in ("dependencies", "devDependencies"):
        for name, version in data.get(section, {}).items():
            deps.append(f"{name} {version}")
    return deps


def collect_dependencies() -> Dict[str, List[str]]:
    deps: Dict[str, List[str]] = {}
    for go_mod in REPO_ROOT.rglob("go.mod"):
        deps[go_mod.relative_to(REPO_ROOT).as_posix()] = parse_go_mod(go_mod)
    for pkg in REPO_ROOT.rglob("package.json"):
        deps[pkg.relative_to(REPO_ROOT).as_posix()] = parse_package_json(pkg)
    return deps


def generate_report(documented: Set[str], actual: Set[str]) -> str:
    """Generate a markdown report summarizing audit results."""
    missing = sorted(documented - actual)
    extra = sorted(actual - documented)

    lines = ["# Stage 1 Inventory Audit Report", ""]
    lines.append(f"Documented files: {len(documented)}")
    lines.append(f"Actual files: {len(actual)}")
    lines.append("")

    lines.append(f"## Missing Files ({len(missing)})")
    if missing:
        lines.append("| File | Owner | Priority |")
        lines.append("| --- | --- | --- |")
        for path in missing:
            owner = OWNER_MAP.get(path.split("/")[1], "Unassigned")
            lines.append(f"| {path} | {owner} | High |")
    else:
        lines.append("(none)")
    lines.append("")

    lines.append(f"## Undocumented Files ({len(extra)})")
    if extra:
        lines.append("| File | Owner | Priority |")
        lines.append("| --- | --- | --- |")
        for path in extra:
            owner = OWNER_MAP.get(path.split("/")[1], "Unassigned")
            lines.append(f"| {path} | {owner} | Review |")
    else:
        lines.append("(none)")
    lines.append("")

    deps = collect_dependencies()
    lines.append("## Dependencies")
    for file, dep_list in deps.items():
        lines.append(f"### {file}")
        for dep in dep_list:
            lines.append(f"- {dep}")
        lines.append("")

    return "\n".join(lines)


def main() -> None:
    parser = argparse.ArgumentParser(description="Audit file inventory against AGENTS.md")
    parser.add_argument("-o", "--output", type=Path, help="Path to write markdown report")
    args = parser.parse_args()

    documented = parse_documented_files()
    actual = list_actual_files()
    report = generate_report(documented, actual)

    if args.output:
        args.output.write_text(report, encoding="utf-8")
    else:
        print(report)


if __name__ == "__main__":
    main()
