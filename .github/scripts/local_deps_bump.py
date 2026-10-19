#!/usr/bin/env python3
"""Local daily-ops counterpart of .github/scripts/dep_version_check.py.

Compares each pkgs/<pkg>/.deps.lock against the Arch official-repo versions
visible in a caller-provided (freshly synced, non-root) pacman db. For every
package whose locked dependency set is stale, creates a complete pkgrel-bump
pull request from the build host:

  - PKGBUILD pkgrel + 1
  - .deps.lock updated in place (comments / ``# MISSING`` lines preserved)
  - .SRCINFO pkgrel synced
  - REVIEW.md 更新履歴 gets the 5-column row (same format the manual
    Phase 1 "3-point fix" used for the old unsigned bot PRs)
  - commit signed with ``-S`` (Nekono GPG / YubiKey) on branch
    ``deps/<pkg>-pkgrel-<N+1>``, push (pre-push gate runs), gh pr create

Idempotency mirrors the workflow script: the branch name encodes the new
pkgrel; an open PR or a leftover remote branch for it -> skip. Merged/closed
PRs do not block, because a pkgver bump resets pkgrel and the branch name is
reused.

Invoked by bin/check-deps. Output is English (ASCII) per repo convention.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
from datetime import date
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
PKGS = ROOT / "pkgs"

VERSION_RE = re.compile(r"^Version\s*:\s*(.+)$")
PKGREL_RE = re.compile(r"^pkgrel=(\d+)", re.MULTILINE)
PKGVER_RE = re.compile(r"^pkgver=(.*)$", re.MULTILINE)
SRCINFO_PKGREL_RE = re.compile(r"^(\s*pkgrel = )(\d+)[ \t]*$", re.MULTILINE)
REVIEW_HEADING = "更新履歴"
REVIEW_HEADING_RE = re.compile(r"^#{1,6}\s*" + re.escape(REVIEW_HEADING))
ISO_DATE_RE = re.compile(r"^\d{4}-\d{2}-\d{2}$")


def run(*cmd: str, check: bool = True, capture: bool = False) -> subprocess.CompletedProcess:
    return subprocess.run(list(cmd), check=check, capture_output=capture, text=True)


def pacman_si_version(pkg: str, dbpath: str) -> str | None:
    env = {**os.environ, "LC_ALL": "C"}
    res = subprocess.run(
        ["pacman", "-Si", "--dbpath", dbpath, pkg],
        capture_output=True,
        text=True,
        env=env,
    )
    if res.returncode != 0:
        return None
    for line in res.stdout.splitlines():
        m = VERSION_RE.match(line)
        if m:
            return m.group(1).strip()
    return None


def parse_lock(path: Path) -> dict[str, str]:
    versions: dict[str, str] = {}
    for line in path.read_text().splitlines():
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        if "=" in line:
            name, ver = line.split("=", 1)
            versions[name.strip()] = ver.strip()
    return versions


def update_lock_in_place(path: Path, changed: dict[str, tuple[str, str]]) -> None:
    """Replace only the changed ``name=version`` lines; keep comments / order."""
    out: list[str] = []
    for line in path.read_text().splitlines():
        stripped = line.strip()
        if stripped and not stripped.startswith("#") and "=" in stripped:
            name = stripped.split("=", 1)[0].strip()
            if name in changed:
                line = f"{name}={changed[name][1]}"
        out.append(line)
    path.write_text("\n".join(out) + "\n")


def bump_pkgrel(path: Path, new_rel: int) -> None:
    text = path.read_text()
    new_text, n = PKGREL_RE.subn(f"pkgrel={new_rel}", text, count=1)
    if n != 1:
        raise RuntimeError(f"pkgrel= line not found in {path}")
    path.write_text(new_text)


def sync_srcinfo(path: Path, new_rel: int) -> None:
    text = path.read_text()
    new_text, n = SRCINFO_PKGREL_RE.subn(rf"\g<1>{new_rel}", text)
    if n != 1:
        raise RuntimeError(f"expected exactly one 'pkgrel = <N>' line in {path}, found {n}")
    path.write_text(new_text)


def add_review_row(path: Path, release: str, summary: str) -> None:
    """Append the Phase-1-style 5-column row under 更新履歴, matching the
    table's existing order (newest-last is the majority; newest-first exists
    for a couple of packages)."""
    lines = path.read_text().splitlines()
    # Match the markdown heading, not an earlier prose mention of the word
    # (e.g. "REVIEW.md「更新履歴」に 1 行追記" in the bump procedure section).
    heading = next((i for i, l in enumerate(lines) if REVIEW_HEADING_RE.match(l.strip())), None)
    if heading is None:
        raise RuntimeError(f"'{REVIEW_HEADING}' heading not found in {path}")
    end = len(lines)
    for j in range(heading + 1, len(lines)):
        if lines[j].startswith("#"):
            end = j
            break

    def row_date(j: int) -> str | None:
        cells = [c.strip() for c in lines[j].strip().strip("|").split("|")]
        if cells and ISO_DATE_RE.match(cells[0]):
            return cells[0]
        return None

    table_rows = [j for j in range(heading + 1, end) if lines[j].lstrip().startswith("|")]
    if not table_rows:
        raise RuntimeError(
            f"no markdown table under '{REVIEW_HEADING}' in {path}; update REVIEW.md manually"
        )
    dated = [j for j in table_rows if row_date(j)]
    if dated:
        if row_date(dated[0]) > row_date(dated[-1]):
            insert_at = dated[0]  # newest-first table
        else:
            insert_at = dated[-1] + 1  # newest-last table (majority)
    else:
        insert_at = table_rows[-1] + 1

    row = f"| {date.today().isoformat()} | {release} | (this PR) | — (pkgrel bump のみ) | {summary} |"
    lines[insert_at:insert_at] = [row]
    path.write_text("\n".join(lines) + "\n")


def existing_open_pr_for_branch(branch: str) -> bool:
    """True if an OPEN PR has this branch as head.

    Merged/closed PRs must not block regeneration: a pkgver bump resets
    pkgrel to 1, so the same branch name (deps/<pkg>-pkgrel-<N>) is
    legitimately reused later. A leftover remote branch is still treated as
    in-flight by remote_branch_exists().
    """
    res = run("gh", "pr", "list", "--head", branch, "--state", "open",
              "--json", "number", capture=True)
    try:
        return bool(json.loads(res.stdout or "[]"))
    except json.JSONDecodeError:
        return False


def remote_branch_exists(branch: str) -> bool:
    """True if the branch already exists on origin (leftover from a failed
    run). Treating it as in-flight prevents pkgrel double-bumps."""
    res = run("git", "ls-remote", "--heads", "origin", branch,
              capture=True, check=False)
    return bool(res.stdout.strip())


def detect_changes(dbpath: str) -> list[dict]:
    entries: list[dict] = []
    for pkg_dir in sorted(p for p in PKGS.iterdir() if p.is_dir()):
        pkg = pkg_dir.name
        lock = pkg_dir / ".deps.lock"
        pkgbuild = pkg_dir / "PKGBUILD"
        srcinfo = pkg_dir / ".SRCINFO"
        review = pkg_dir / "REVIEW.md"
        if not (lock.exists() and pkgbuild.exists() and srcinfo.exists() and review.exists()):
            continue
        old = parse_lock(lock)
        if not old:
            continue
        changed: dict[str, tuple[str, str]] = {}
        for dep, old_ver in old.items():
            cur = pacman_si_version(dep, dbpath)
            if cur is None:
                continue  # not in the official repos (virtual provider / AUR-only)
            if cur != old_ver:
                changed[dep] = (old_ver, cur)
        if not changed:
            continue

        text = pkgbuild.read_text()
        m = PKGREL_RE.search(text)
        if not m:
            print(f"[{pkg}] WARN: pkgrel= not found; skipping", file=sys.stderr)
            continue
        old_rel = int(m.group(1))
        new_rel = old_rel + 1

        vm = PKGVER_RE.search(text)
        pkgver = vm.group(1).strip().strip('"').strip("'") if vm else ""
        if " " in pkgver:
            pkgver = pkgver.split()[0]
        if not pkgver or "$" in pkgver:
            print(f"[{pkg}] WARN: pkgver not resolved from PKGBUILD; skipping", file=sys.stderr)
            continue

        entries.append({
            "pkg": pkg,
            "pkg_dir": pkg_dir,
            "changed": changed,
            "old_rel": old_rel,
            "new_rel": new_rel,
            "branch": f"deps/{pkg}-pkgrel-{new_rel}",
            "release": f"{pkgver}-{new_rel}",
        })
    return entries


def create_pr(entry: dict) -> None:
    pkg = entry["pkg"]
    changed = entry["changed"]
    old_rel, new_rel = entry["old_rel"], entry["new_rel"]
    branch, release = entry["branch"], entry["release"]
    pkg_dir: Path = entry["pkg_dir"]

    if existing_open_pr_for_branch(branch) or remote_branch_exists(branch):
        print(f"[{pkg}] branch or PR {branch} already exists, skipping")
        return

    run("git", "checkout", "-B", branch)
    bump_pkgrel(pkg_dir / "PKGBUILD", new_rel)
    update_lock_in_place(pkg_dir / ".deps.lock", changed)
    sync_srcinfo(pkg_dir / ".SRCINFO", new_rel)
    dep_summary = ", ".join(f"{d} {o} → {n}" for d, (o, n) in sorted(changed.items()))
    add_review_row(pkg_dir / "REVIEW.md", release, f"`pkgrel` +1 (deps changed): {dep_summary}")

    run("git", "add",
        str(pkg_dir / "PKGBUILD"),
        str(pkg_dir / "REVIEW.md"),
        str(pkg_dir / ".SRCINFO"),
        str(pkg_dir / ".deps.lock"))
    run("git", "commit", "-S",
        "-m", f"{pkg}: bump pkgrel to {new_rel} (deps changed)",
        "-m", "Generated by bin/check-deps (daily-ops local check). Signed and "
              "complete (REVIEW.md / .SRCINFO / .deps.lock updated); a judgment "
              "review is the only step left before bin/step-merge.")

    run("git", "push", "-u", "origin", branch)

    body_lines = [
        "## Summary",
        f"- pkg: `{pkg}`",
        f"- pkgrel: {old_rel} → {new_rel}",
        "- 依存パッケージの Arch 公式 repo 上 version が変化したため pkgrel bump (rebuild 用)",
        "",
        "## Dep changes",
    ]
    for dep, (o, n) in sorted(changed.items()):
        body_lines.append(f"- `{dep}`: `{o}` → `{n}`")
    body_lines += [
        "",
        "## 生成元",
        "`bin/check-deps` (daily-ops のローカル検知) が生成した **署名済み・完成 PR**",
        "(REVIEW.md / .SRCINFO / .deps.lock 反映済み)。merge 前に必要なのは",
        "judgment review (`gh pr diff`) のみ。",
        "",
        "Generated by bin/check-deps (daily-ops local check).",
    ]
    body = "\n".join(body_lines)

    try:
        run("gh", "pr", "create",
            "--title", f"{pkg}: pkgrel bump to {new_rel} (deps changed)",
            "--body", body,
            "--base", "master",
            "--head", branch)
    except subprocess.CalledProcessError:
        # Keep the remote state clean so the next run can retry this package.
        print(f"[{pkg}] PR creation failed, removing remote branch {branch}", file=sys.stderr)
        run("git", "push", "origin", "--delete", branch, check=False)
        run("git", "checkout", "master", check=False)
        raise
    print(f"[{pkg}] opened PR for branch {branch}")
    run("git", "checkout", "master")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--dbpath", required=True)
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()

    if not Path(args.dbpath).is_dir():
        print(f"FATAL: dbpath not found: {args.dbpath}", file=sys.stderr)
        return 2
    # Canary: a broken dbpath would make every lookup fail silently and look
    # like "no dep changes".
    if pacman_si_version("pacman", args.dbpath) is None:
        print("FATAL: cannot query the throwaway pacman db (canary 'pacman' failed)", file=sys.stderr)
        return 2

    entries = detect_changes(args.dbpath)
    if not entries:
        print("no dep changes")
        return 0

    print(f"detected {len(entries)} package(s) with dep changes")
    failed = 0
    for entry in entries:
        dep_summary = ", ".join(f"{d} {o} -> {n}" for d, (o, n) in sorted(entry["changed"].items()))
        if args.dry_run:
            print(f"[{entry['pkg']}] would bump pkgrel {entry['old_rel']} -> {entry['new_rel']} "
                  f"({entry['branch']}): {dep_summary}")
            continue
        try:
            create_pr(entry)
        except (subprocess.CalledProcessError, RuntimeError) as e:
            print(f"[{entry['pkg']}] FAILED: {e}", file=sys.stderr)
            run("git", "checkout", "--", ".", check=False)
            run("git", "checkout", "master", check=False)
            failed += 1

    if args.dry_run:
        print("dry-run: nothing changed")
        return 0
    print(f"done (failed: {failed})")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
