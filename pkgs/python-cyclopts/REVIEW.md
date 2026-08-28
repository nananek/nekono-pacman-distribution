# python-cyclopts review

## 状態

**review 済み、 approve** (最新: 2026-08-18 / 4.23.0)

AUR `python-cyclopts` 4.12.0-1 を fork。 **AUR の supply-chain 欠陥
(`sha256sums=('SKIP')`) を fix** して [nekono] に投入。

## 用途

`python-fastmcp` 3.2.4 の **直接依存** (= upstream pyproject.toml の
`cyclopts>=4.0.0`)。 fastmcp dep audit で漏れ発覚した 10 個のうちの 1 つ、
[nekono] 投入が必要だった分。

## Source

- Upstream: https://github.com/BrianPugh/cyclopts
  - tag `v4.12.0`、 license Apache-2.0
  - PyPI sdist (`cyclopts-4.12.0.tar.gz`) を vendor
- AUR: https://aur.archlinux.org/packages/python-cyclopts
  - AUR maintainer: Jesus Alvarez <jesusalv@rez.codes>

## 検証結果

- [x] `source` URL = `files.pythonhosted.org/packages/source/c/cyclopts/cyclopts-4.12.0.tar.gz`
  - PyPI 公式 CDN
- [x] `sha256sums` 独立検証
  - build host で curl + sha256sum で実測: `86bfb5b35cb078decc1cca6c1be41f9a0e6202dc43b4f6056d5cfc6d1f4a69d1`
  - **AUR は `SKIP` で hash 検証を skip** していた supply-chain 欠陥を [nekono]
    では実 hash で固定
- [x] `build()`: `python -m build --wheel --no-isolation` のみ
- [x] `package()`: `python -m installer --destdir` のみ
- [x] `depends`: python / python-attrs / python-docstring-parser / python-rich /
  python-rich-rst — Arch 公式 + [nekono] cross-dep 2 個 ([[python-docstring-parser]] / [[python-rich-rst]])
- [x] `makedepends`: python-build / python-installer / python-wheel / python-hatchling / python-hatch-vcs — Arch 公式
- [x] `secrets` 混入なし
- [x] `arch=('any')` — pure Python

## AUR との意図的差分

| 項目 | 差分 | 理由 |
|---|---|---|
| `# Maintainer:` | Jesus Alvarez → Nekono、 Jesus を Contributor に降格 | [nekono] fork 表示 |
| `sha256sums` | `'SKIP'` → `'86bfb5b35cb078decc1cca6c1be41f9a0e6202dc43b4f6056d5cfc6d1f4a69d1'` | **supply-chain audit 必須**、 SKIP は [nekono] 規約違反 (= AUR maintainer に修正 PR 送るのが筋だが [nekono] では先に正しい状態を反映) |

それ以外は **0 行差分** (= 純 fork)。

## 結論

**approve** — build host で `bin/build-all python-cyclopts` で build + sign + repo db 追加可。 ただし **python-docstring-parser (= PR #70) + python-rich-rst (= PR #71) を先に publish してから** でないと depends が解決しない。

## 更新方針

upstream で新 release が出たら nvchecker (= `[python-cyclopts]` section) が検知 → Issue 経由で人間が手作業更新。 sha256 は **毎 bump 必ず独立計算** (= AUR が SKIP のままなら [nekono] でも入れ続ける必要)。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-21 | 4.14.1-1 | (本 commit 後に確定) | upstream tag `v4.14.1` | python-rich-rst 上限緩和 (<2.0.0→<3.0.0)。PKGBUILD 改変は pkgver + sha256 の 2 値のみ。safe-to-bump。 |
| 2026-06-03 | 4.16.1-1 | (this PR) | upstream tag `v4.16.1` | safe-to-bump。 機能追加 3 件 (App.synonym / __cyclopts_returncode__ / PEP 692) + バグ修正 2 件。 deps / build 無変化。 sha256 独立検証済み |
| 2026-06-10 | 4.17.0-1 | (this PR) | upstream tag `v4.17.0` | safe-to-bump。 機能追加 (async コマンド対応 / Parameter.show_default の文字列対応) + PyPI publish attestation 導入 (供給チェーン強化)。 deps / build 無変化。 sha256 独立検証済み |
| 2026-06-12 | 4.18.0-1 | (this PR) | upstream tag `v4.18.0` | safe-to-bump (Issue #199)。 slice 型 native サポート + Slice validator / NonEmptySlice 追加、 ネスト Annotated/Optional/NewType の Parameter メタデータ脱落修正。 deps / build 無変化。 sha256 独立検証済み (PyPI API + 実測一致) |
| 2026-06-15 | 4.18.0-2 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-hatchling 1.29.0-1 → 1.30.1-1 |
| 2026-06-21 | 4.18.0-3 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python 3.14.5-1 → 3.14.6-1 |
| 2026-05-20 | 4.12.0-1 | `0cced1785b136180286916e6e0225cc709cb9de7` | upstream tag `v4.12.0` | 初回 add、 AUR fork + sha256 SKIP fix、 fastmcp 直接依存 |
| 2026-07-03 | 4.20.0-1 | (this PR) | upstream tag `v4.20.0` (`db654ce65d2ae47759b283154c4fa000aa60bcc2`) | safe-to-bump (Issue #329)。 v4.19.0 (Parameter.short_alias / filter_by) + v4.20.0 (`cyclopts tree` サブコマンド、 network I/O 無し) の 2 release 分。 deps / build 無変化 (pyproject.toml diff 0)。 sha256 独立検証済み (PyPI CDN 実測 + PyPI JSON API 一致)。 pkgrel reset 3→1 |
| 2026-07-06 | 4.20.0-2 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-rich-rst 2.0.2-1 → 2.1.0-1 rebuild に追随 (build-all の cascade-warn 由来) |
| 2026-07-13 | 4.21.0-1 | (this PR) | upstream tag `v4.21.0` | safe-to-bump (Issue #385)。 `Parameter.negative_alias` 追加 (negative form flag に短縮 alias 付与) + `default_factory` dataclass コマンドで pydantic import 時に誤バリデーションエラーが出る不具合修正。 deps / build 無変化 (pyproject.toml diff 0、sdist ファイル一覧完全一致)。 sha256 独立検証済み (curl 実測 + PyPI JSON API 一致)。 pkgrel reset 2→1 |
| 2026-07-17 | 4.21.1-1 | (this PR) | upstream tag `v4.21.1` | safe-to-bump (Issue #409)。 zsh 補完のサブコマンド dispatch バグ修正 (meta positional 後のサブコマンドがスロット固定される問題)、bash/fish は部分対応。 deps / build 無変化 (pyproject.toml diff 0)。 sha256 独立検証済み (curl 実測 + PyPI JSON API 一致)。 |
| 2026-07-22 | 4.22.1-1 | (this PR) | upstream tag `v4.22.1` | safe-to-bump (Issue #425)。 v4.21.2 (`--` delimiter がエラーメッセージに混入するバグ修正) + v4.22.0 (config search_parents=False の親ディレクトリ walk バグ / 不正 timedelta 文字列 reject / 符号付き base-prefix 整数パース修正) + v4.22.1 (env_var iterable 分割 / boolean flag エラーメッセージ改善) の 3 release 分。 deps / build 無変化 (pyproject.toml diff 0)。 sha256 独立検証済み (curl 実測 + PyPI JSON API 一致)。 |
| 2026-07-27 | 4.22.2-1 | (this PR) | upstream tag `v4.22.2` | safe-to-bump (Issue #439)。 空 mapping bind バグ修正 (#871) + `Parameter.count=True` flag への `=value` 明示指定サポート (#872) + config root_keys が table でない場合の CycloptsError 化 (#873) の 3 PR 分、いずれもバグ修正のみ。 deps / build 無変化 (pyproject.toml diff 0、sdist ファイル一覧完全一致)。 sha256 独立検証済み (curl 実測 + PyPI JSON API 一致)。 |
| 2026-07-31 | 4.22.3-1 | (this PR) | upstream tag `v4.22.3` | safe-to-bump (Issue #462)。`__init__` が定義された `Enum` 型のパース失敗を修正 (PR #877)。バグ修正のみ、deps / build 無変化 (pyproject.toml diff 0、sdist ファイル一覧完全一致)。sha256 独立検証済み (curl 実測 + PyPI JSON API 一致)。 |
| 2026-08-04 | 4.22.4-1 | (this PR) | upstream tag `v4.22.4` | safe-to-bump (Issue #478)。`ConsumeMultipleError` のエラーメッセージで要求数 1 の場合の名詞単数形化のみ (PR #879)。deps / build 無変化 (pyproject.toml byte 一致、sdist diff は PKG-INFO/_version.py/exceptions.py の 3 file のみ)。sha256 独立検証済み (curl 実測 + PyPI JSON API 一致)。 |
| 2026-08-06 | 4.22.5-1 | (this PR) | upstream tag `v4.22.5` (`954e72e1c0fd25d75ac67955a89e80f8e36e864b`) | safe-to-bump (Issue #485)。抽象コレクション型 (Collection/Container/Reversible) 周りのバグ修正 3 件 (PR #881/#882/#883): 値なし渡し時の TypeError 修正、型ヒントサポート追加、`--empty-*` flag 対応拡張。deps / build 無変化 (pyproject.toml byte 一致、sdist ファイル一覧完全一致)。sha256 独立検証済み (curl 実測 + Issue 記載値一致)。 |
| 2026-08-18 | 4.23.0-1 | (this PR) | upstream tag `v4.23.0` (`b393026ab6cbba88b3e7563447051bdc5c8dfb87`) | safe-to-bump (Issue #547)。`cyclopts docs` の markdown formatter で colon 直後の余分な空白を修正 (PR #888) のみの thin bugfix release、CLI パース挙動には影響なし。deps / build 無変化 (pyproject.toml byte 一致、sdist ファイル一覧完全一致)。sha256 独立検証済み (curl 実測 + Issue 記載値一致)。 |
| 2026-08-24 | 4.23.2-1 | (this PR) | upstream tag `v4.23.2` (`22cfd43249ee740e8815626a29d7da3d095501b0`) | safe-to-bump (Issue #560)。v4.23.1 分含む bugfix 2 release 分 (upstream changelog に大きな挙動変化の記載なし)。pyproject.toml 完全一致 (byte diff 0) につき deps / build 無変化。sha256 独立検証済み (curl 実測 + PyPI JSON API 一致)。 |
