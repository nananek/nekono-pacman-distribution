# nekono-btkeycast review

## 状態

**review 済み、 approve** (2026-06-12、 v0.1.0 initial release)

自家 source PKGBUILD (= AUR fork ではない、 nananek 自身の upstream を
GitHub release tarball から直接 vendor)。 own-package フローは
nekono-pipewire-mcp / nekono-voicevox-mcp と同一。

## 用途

ayaka のキーボードを **BLE HID-over-GATT (HOGP) キーボード** として iPad
(or 任意の BLE central) へ転送する常駐デーモン + waybar モジュール。
bluetoothd の D-Bus API (= BlueZ GATT server registration) 経由で HID
service を advertise し、 Gtk3 + gtk-layer-shell の popup overlay で
状態表示する。

dotfiles 側の waybar 配線 (`custom/btkeycast`) は配線済みで、
`exec-if: which btkeycast` のため本パッケージ導入で自動的に有効化される。
user session 内で waybar から起動する常駐プロセスであり、 system daemon
ではないので systemd unit / sysusers / tmpfiles は同梱しない。

## Source

- Upstream: https://github.com/nananek/nekono-btkeycast
  - 主開発者 nananek (= 自家 upstream)
  - tag `v0.1.0` (= prefix "v")、 commit `54beee2171a0e4f683286a295b0cdd70b1ccc1cd`
  - tag GPG 署名: **verified** (= GitHub API `verification.verified=true` を review 時に確認済み)
  - license: MIT
- distfile: GitHub Release archive (= `archive/refs/tags/v0.1.0.tar.gz`)

## 検証結果

- [x] `source` URL = `github.com/nananek/nekono-btkeycast/archive/refs/tags/v0.1.0.tar.gz`
  - 自家 upstream、 typosquat 検討対象外
- [x] `sha256sums` 独立検証
  - 実測 (review 時に curl 直取得 + `sha256sum`): `21cd9bcb7dd098b5a885ab786c2836d86fbd2ce6fd9f65497ab1ba6ed439fe1d`
- [x] tag `v0.1.0` GPG 署名状態: verified (上記 Source 参照)
- [x] `build()`: `python -m build --wheel --no-isolation` のみ、 ネットワーク取得なし
  - upstream `pyproject.toml` の `dependencies = []` (= pip 依存ゼロ、 vendor 不要)
- [x] `package()`: `python -m installer --destdir` + LICENSE / README.md 配置のみ
  - console_script `btkeycast` (= `btkeycast.cli:main`) は wheel dist-info 経由で `/usr/bin/` に自動配置
  - shell injection / exec / curl / wget なし
- [x] `depends`: upstream source の import 実査と対応を確認
  - `python-dbus` — `import dbus` / `dbus.service` / `dbus.mainloop.glib` (hog.py)
  - `python-gobject` — `import gi` + `gi.repository.GLib/Gtk/Gdk` (daemon.py / ui.py)
  - `gtk3` — `gi.require_version('Gtk', '3.0')` / `('Gdk', '3.0')`
  - `gtk-layer-shell` — `gi.require_version('GtkLayerShell', '0.1')`
  - `bluez` — bluetoothd D-Bus API (BLE GATT server 登録) の runtime 前提
- [x] `optdepends`: `libnotify` — cli.py は `shutil.which('notify-send')` で存在
  チェックして無ければ skip する (= hard dep 不要を code で確認)
- [x] `makedepends`: python-build / python-installer / python-wheel / python-hatchling — 標準的な hatchling build pattern
- [x] `license`: MIT — SPDX、 LICENSE を `/usr/share/licenses/` に配置
- [x] `secrets` 混入なし
- [x] `arch=('any')` — pure Python、 C extension 無し

## 設計判断 (= AUR fork でないので「AUR との意図的差分」 ではなく §設計判断)

| 判断 | 理由 |
|---|---|
| **systemd unit を同梱しない** | waybar (`custom/btkeycast`) が user session 内で起動・常駐させる。 system daemon 化しない |
| **bluez を depends に置く (optdepends ではなく)** | bluetoothd の D-Bus API が無いとデーモンが起動即 fail。 BLE 転送が本体機能なので hard dep |
| **libnotify は optdepends** | `notify-send` 不在時は graceful skip するコードを確認済み |
| **python-evdev は depends (optdepends ではなく)** | code は ImportError を guard して EXCLUSIVE capture に fallback するが、 0.2.0 の主要機能 (キーボード個別転送) が黙って無効化されるのは UX 事故のもと。 box から出してすぐ全機能が動く状態を取る |
| **コマンド名 `btkeycast` ≠ pkgname** | upstream プロジェクト名は btkeycast (wheel 名 `btkeycast-0.1.0-*.whl`)。 nekono-*-mcp と同じ流儀で pkgname のみ nekono- prefix |
| **arch=('any')** | pure Python、 C extension 無し |
| **`# Contributor:` 等の引用なし** | AUR fork ではないので不要 |

## 結論

**approve** — build host で `bin/build-all nekono-btkeycast` で build + sign + repo db 追加可。

publish 後、 ayaka 側で:
1. `sudo pacman -Sy nekono-btkeycast` で install (= /usr/bin/btkeycast 配置)
2. `pkill -SIGUSR2 -x waybar` (config reload → `custom/btkeycast` が有効化)

## 更新方針

upstream で新 release (= v0.2.0 等) が出たら:
1. user が `git tag -s v0.2.0 -m "Release 0.2.0"` + `git push origin v0.2.0` + GitHub Release publish
2. nvchecker (GitHub Releases、 `nananek/nekono-btkeycast`、 prefix "v") が新 release を検知 → Issue が立つ
3. PKGBUILD の pkgver を更新、 sha256 を独立再計算 (= `curl + sha256sum`)
4. depends 変化があれば反映 (= import の増減を source diff で確認)
5. `.SRCINFO` + `.deps.lock` を再生成
6. REVIEW.md「更新履歴」 に 1 行追記

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-06-12 | 0.1.0 | (this PR) | `54beee2171a0e4f683286a295b0cdd70b1ccc1cd` | 初回 add、 自家 source v0.1.0 (Issue #210) |
| 2026-06-12 | 0.2.0 | (this PR) | `2742671feec268a5b06b58f2390acf58b469625f` | 0.2.0 bump (flock 単一インスタンス / evdev キーボード個別転送 / 接続先 pin)。 depends += python-evdev (= `import evdev` を source 実査)。 sha256 独立再計算、 tag GPG 署名 verified (GitHub API) |
