#!/usr/bin/env python3
"""Compare nvchecker newver.json against each PKGBUILD's pkgver value.

Reads ``newver.json`` (nvchecker state-file format) from the repository root
and walks ``pkgs/*/PKGBUILD`` to compute the set of packages whose upstream
``version`` differs from the in-repo ``pkgver``.

The result is emitted as a JSON array of ``{"pkg", "old_pkgver",
"new_pkgver"}`` objects to stdout, suitable for consumption by GitHub Actions
``${{ steps.<id>.outputs.updates }}`` and matrix expansion in a downstream
job.

Designed to be invoked by ``.github/workflows/upstream-version-issue.yml``.
"""
from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
NEWVER = ROOT / "newver.json"
PKGS_DIR = ROOT / "pkgs"

PKGVER_RE = re.compile(r"^pkgver=(.*)$")


def pkgver_of(pkgbuild: Path) -> str:
    for line in pkgbuild.read_text().splitlines():
        m = PKGVER_RE.match(line)
        if m:
            return m.group(1).strip().strip('"').strip("'")
    return ""


def main() -> int:
    if not NEWVER.exists():
        print("[]")
        return 0
    try:
        state = json.loads(NEWVER.read_text() or "{}")
    except json.JSONDecodeError:
        # Treat malformed state file as "no updates" rather than crashing
        # the workflow; nvchecker itself will have surfaced the error.
        print("[]")
        return 0
    data = state.get("data", {}) if isinstance(state, dict) else {}
    updates = []
    for pkg, info in sorted(data.items()):
        new_ver = (info or {}).get("version", "")
        if not new_ver:
            continue
        pkgbuild = PKGS_DIR / pkg / "PKGBUILD"
        if not pkgbuild.exists():
            continue
        old_ver = pkgver_of(pkgbuild)
        if old_ver and old_ver != new_ver:
            updates.append(
                {"pkg": pkg, "old_pkgver": old_ver, "new_pkgver": new_ver}
            )
    json.dump(updates, sys.stdout)
    return 0


if __name__ == "__main__":
    sys.exit(main())
