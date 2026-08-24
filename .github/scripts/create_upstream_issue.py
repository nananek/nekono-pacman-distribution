#!/usr/bin/env python3
"""Create a deterministic, human-review issue for an upstream version bump."""
from __future__ import annotations

import argparse
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]


def source_lines(pkgbuild: Path) -> list[str]:
    lines: list[str] = []
    in_source = False
    for line in pkgbuild.read_text().splitlines():
        if line.startswith("source="):
            in_source = True
        if in_source:
            lines.append(line)
            if ")" in line and not line.rstrip().endswith("\\"):
                break
    return lines or ["(source entry not found)"]


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--pkg", required=True)
    parser.add_argument("--old-version", required=True)
    parser.add_argument("--new-version", required=True)
    args = parser.parse_args()
    pkgbuild = ROOT / "pkgs" / args.pkg / "PKGBUILD"
    body = "\n".join([
        "## Automated detection",
        "この Issue は nvchecker の version 差分から機械的に生成されました。",
        "AI による安全性判定や PKGBUILD の自動変更は行っていません。",
        "",
        f"- package: `{args.pkg}`",
        f"- current pkgver: `{args.old_version}`",
        f"- detected version: `{args.new_version}`",
        "",
        "## PKGBUILD source (current)",
        "```bash",
        *source_lines(pkgbuild),
        "```",
        "",
        "## Human review checklist",
        "- [ ] upstream の公式 release/tag と source URL を確認",
        "- [ ] 新しい source の sha256 を実測して pin",
        "- [ ] build()/package()/prepare() と依存関係の差分を確認",
        "- [ ] release notes と install script の変更を確認",
        "- [ ] REVIEW.md と .SRCINFO を更新して署名付き commit を作成",
        "",
        "## Suggested action",
        "人間が上記を確認した後、必要な PKGBUILD 更新を行ってください。",
    ])
    subprocess.run(
        ["gh", "issue", "create", "--title",
         f"[{args.pkg}] upstream version: {args.new_version}", "--body", body],
        check=True,
        cwd=ROOT,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
