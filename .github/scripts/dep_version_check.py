#!/usr/bin/env python3
"""Compare each pkgs/<pkg>/.deps.lock against the current Arch official-repo
versions reported by ``pacman -Si``. For every package whose locked
dependency set is stale, bump ``pkgrel`` by 1, rewrite the lock, then push a
branch and open a PR against ``master``.

Idempotency:
  - Branch name encodes the *new* pkgrel: ``deps/<pkg>-pkgrel-<N+1>``.
  - Before pushing, the script checks if a PR already exists for that branch
    name (``gh pr list --head ... --state all``). If one is found, the script
    skips push to avoid clobbering an in-flight review.
  - Commits are *unsigned*. By repository convention (CLAUDE.md) the merging
    human re-signs the head commit with the Nekono GPG key (YubiKey) before
    merging.

Designed to be invoked by ``.github/workflows/dep-version-pr.yml``.
"""
from __future__ import annotations

import json
import os
import re
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
PKGS = ROOT / "pkgs"

VERSION_RE = re.compile(r"^Version\s*:\s*(.+)$")
PKGREL_RE = re.compile(r"^pkgrel=(\d+)", re.MULTILINE)


def pacman_si_version(pkg: str) -> str | None:
    env = {**os.environ, "LC_ALL": "C"}
    try:
        result = subprocess.run(
            ["pacman", "-Si", pkg],
            capture_output=True,
            text=True,
            env=env,
            check=True,
        )
    except subprocess.CalledProcessError:
        return None
    for line in result.stdout.splitlines():
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


def write_lock(path: Path, pkg: str, versions: dict[str, str]) -> None:
    header = [
        f"# Snapshot of Arch official repo versions for depends + makedepends of {pkg}.",
        "# Auto-generated and updated by .github/workflows/dep-version-pr.yml.",
        f"# Source: pkgs/{pkg}/.SRCINFO (depends + makedepends entries).",
        "# Format: 1 line per package: <name>=<full-version>",
        "",
    ]
    body = [f"{name}={versions[name]}" for name in sorted(versions)]
    path.write_text("\n".join(header + body) + "\n")


def bump_pkgrel(pkgbuild: Path) -> tuple[int, int]:
    text = pkgbuild.read_text()
    m = PKGREL_RE.search(text)
    if not m:
        raise RuntimeError(f"pkgrel= not found in {pkgbuild}")
    old = int(m.group(1))
    new = old + 1
    text = PKGREL_RE.sub(f"pkgrel={new}", text, count=1)
    pkgbuild.write_text(text)
    return old, new


def run(*cmd: str, check: bool = True, capture: bool = False) -> subprocess.CompletedProcess:
    return subprocess.run(
        list(cmd),
        check=check,
        capture_output=capture,
        text=True,
    )


def existing_pr_for_branch(branch: str) -> bool:
    """Return True if any PR (open or closed) targets this branch as head."""
    res = run(
        "gh", "pr", "list",
        "--head", branch,
        "--state", "all",
        "--json", "number",
        capture=True,
    )
    try:
        return bool(json.loads(res.stdout or "[]"))
    except json.JSONDecodeError:
        return False


def remote_branch_exists(branch: str) -> bool:
    """Return True if the branch already exists on the origin remote.

    Needed because a prior workflow run can have pushed the branch but
    failed to open the PR (e.g. the repo's Actions permission was missing
    "create pull requests" at the time). Treating a leftover branch as
    "already in flight" prevents re-bumping pkgrel and producing an
    out-of-sync N+2 commit on top of master.
    """
    res = run(
        "git", "ls-remote", "--heads", "origin", branch,
        capture=True, check=False,
    )
    return bool(res.stdout.strip())


def process_pkg(pkg_dir: Path) -> None:
    pkg = pkg_dir.name
    lock = pkg_dir / ".deps.lock"
    pkgbuild = pkg_dir / "PKGBUILD"
    if not lock.exists() or not pkgbuild.exists():
        return

    old = parse_lock(lock)
    if not old:
        return

    new = dict(old)
    changed: dict[str, tuple[str, str]] = {}
    for dep, old_ver in old.items():
        cur = pacman_si_version(dep)
        if cur is None:
            # Not in official repo (virtual / AUR-only). Skip silently; keeping
            # the lock entry as-is avoids spurious bumps when pacman cannot
            # resolve a virtual provider.
            continue
        if cur != old_ver:
            new[dep] = cur
            changed[dep] = (old_ver, cur)

    if not changed:
        print(f"[{pkg}] no dep changes")
        return

    # Compute new_rel without mutating the file yet. The previous design
    # called bump_pkgrel() here and then did `git checkout master`, but git
    # checkout preserves dirty working-tree changes, so a second bump_pkgrel()
    # after the checkout would land on the *already bumped* file and produce
    # pkgrel=N+2 (observed once in PR #21 going 1→3).
    text = pkgbuild.read_text()
    m = PKGREL_RE.search(text)
    if not m:
        raise RuntimeError(f"pkgrel= not found in {pkgbuild}")
    old_rel = int(m.group(1))
    new_rel = old_rel + 1

    branch = f"deps/{pkg}-pkgrel-{new_rel}"

    # Fetch to ensure remote refs are visible (for the duplicate check below).
    run("git", "fetch", "origin", check=False)

    if existing_pr_for_branch(branch) or remote_branch_exists(branch):
        print(f"[{pkg}] branch or PR {branch} already exists, skipping")
        return

    # Branch off the current ref (master is assumed) before any working-tree
    # mutation, so the bump is applied exactly once on top of master.
    run("git", "checkout", "-B", branch)
    bump_pkgrel(pkgbuild)
    write_lock(lock, pkg, new)
    run("git", "add", str(pkgbuild), str(lock))

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
        "## 注意",
        "このコミットは **unsigned** (GitHub Actions に Nekono GPG 秘密鍵は",
        "持ち込まない方針)。merge 前に build host で取り込んで sign し直し",
        "てから merge してください。",
    ]
    body = "\n".join(body_lines)

    run(
        "git", "commit",
        "-m", f"{pkg}: bump pkgrel to {new_rel} (deps changed)",
        "-m", "Auto-generated by .github/workflows/dep-version-pr.yml",
    )
    run("git", "push", "-u", "origin", branch)

    try:
        run(
            "gh", "pr", "create",
            "--title", f"{pkg}: pkgrel bump to {new_rel} (deps changed)",
            "--body", body,
            "--base", "master",
            "--head", branch,
        )
    except subprocess.CalledProcessError:
        # PR creation failed (typically a repo permission issue, e.g. "Allow
        # GitHub Actions to create and approve pull requests" not enabled).
        # Roll back the remote branch so the next workflow run can re-attempt
        # cleanly. Without this cleanup, remote_branch_exists() would short-
        # circuit on the leftover branch and the PR would never be created.
        print(
            f"[{pkg}] PR creation failed, removing remote branch {branch}",
            file=sys.stderr,
        )
        run("git", "push", "origin", "--delete", branch, check=False)
        run("git", "checkout", "master", check=False)
        raise
    print(f"[{pkg}] opened PR for branch {branch}")

    # Return to master for the next package iteration.
    run("git", "checkout", "master")


def main() -> int:
    # Refresh pacman db once.
    run("pacman", "-Sy", "--noconfirm")
    for pkg_dir in sorted(p for p in PKGS.iterdir() if p.is_dir()):
        try:
            process_pkg(pkg_dir)
        except subprocess.CalledProcessError as e:
            # Surface the failure but continue with the next package so a
            # single misconfigured PKGBUILD doesn't block everything.
            print(f"[{pkg_dir.name}] FAILED: {e}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
