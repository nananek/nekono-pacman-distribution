# moonlight-qt-nekono review

## 状態

**review 済み、approve** (2026-09-15 / fork tag v6.1.0-nekono.3)

**本 repo で初めて「自分の fork を自分でビルドする」package。** 他の package が
upstream の release artifact を pin して再梱包するのに対し、これは信頼の起点が
Nekono 自身の commit になる。下記「信頼モデルの差」を読んでから運用すること。

## Source

- Fork: https://github.com/nananek/moonlight-qt (branch `multidisplay`)
  - Upstream: https://github.com/moonlight-stream/moonlight-qt (v6.1.0 + 578 commits)
  - GPL-3.0-or-later。ホストの複数ディスプレイを1ウィンドウずつ表示できるようにした改造版。
    配信元は本 repo の `sunshine-bin` (将来 `sunshine-nekono` に差し替え予定)
- PKGBUILD の下敷き: AUR `moonlight-qt`
  (maintainers: Konstantin Liberty / Cedric Girard)
  - 変更点: source を fork の release tarball に変更、`libplacebo` を depends に明示、
    `provides`/`conflicts` に `moonlight-qt` を追加、AUR 版の 2 つの sed patch
    (`string.h` / `cstring` の追記) を削除 — 現 Arch の gcc/glibc では不要で、
    実際に本 host で patch 無しにビルドが通ることを確認済み

## Tarball の素性

`source` は GitHub の自動生成 archive **ではない**。submodule を含める必要があるため、
下記 commit を `git clone --recurse-submodules` した作業ツリーから `.git` を除去して
生成し、fork の release asset として公開したもの。

| 対象 | commit |
|---|---|
| moonlight-qt (superproject) | `b87099a54d5b905db5f7afee55344fde2b8a7b4b` |
| moonlight-common-c | `6cb7e7aa1cb29dd70629685056cf8b57399292fa` (これも Nekono fork) |
| qmdnsengine | `920c097ffa742e2968290f15d4dde6693aec02e5` |
| app/SDL_GameControllerDB | `8d9fefd7b810f2541f78cc7a8ccbd185bc84c7a5` |

## 検証結果

- [x] `source` URL = `github.com/nananek/moonlight-qt/releases/download/v6.1.0-nekono.3/moonlight-qt-6.1.0.nekono3.tar.gz`
  - 自分の org/repo。typosquat の余地なし
- [x] `sha256sums` を **独立再計算で検証**
  - 公開済み asset を再 download して `sha256sum` =
    `3ba2f6cc33057b5a0ba74959c6aa358fcebfdfb5342e683003b0584b98129c3b`
  - PKGBUILD 値と一致。SKIP 不使用
  - 同じ作業ツリーから tar を作り直しても同一 hash (決定的生成を確認)
- [x] `prepare()` は `qmake6 PREFIX=/usr moonlight-qt.pro` のみ。patch 無し
- [x] `build()` は `make release` のみ。**ネットワーク取得なし**
  - submodule は tarball に同梱済みで `git submodule update` を呼ばない
    (AUR の `sunshine` PKGBUILD が build 中に `git submodule update` する形とは対照的)
  - curl / wget / pipe-to-shell / eval / pip install いずれも無し
- [x] `package()` は `make INSTALL_ROOT="$pkgdir" install` のみ
- [x] `depends` は AUR 版 + `libplacebo`。`libplacebo` は Vulkan レンダラが直接 link
  しており、ffmpeg 経由の推移依存に頼らず明示した
- [x] install scriptlet 無し、setcap 等の特権操作無し

## 信頼モデルの差 (要注意)

他 package は「upstream の署名/CI が作った artifact を pin」しているが、本 package は
**Nekono 自身が書いた改造コードを Nekono 自身がビルドする**。したがって:

- supply-chain 上の攻撃面は「fork repo への write 権限」そのもの。GitHub アカウントの
  2FA と push 権限の管理が実質的な防御線になる
- upstream の security fix は自動で入らない。`multidisplay` branch を定期的に
  upstream master に rebase/merge する運用が必要
- tarball は手作業生成。ただし下記の決定的な手順を踏んでおり、同じ commit からは
  **バイト単位で同じ tarball が再生成できる**:
  ```sh
  git clone --recurse-submodules --shallow-submodules -b multidisplay \
      https://github.com/nananek/moonlight-qt.git moonlight-qt-<pkgver>
  find moonlight-qt-<pkgver> -name .git -prune -exec rm -rf {} +
  tar --sort=name --owner=0 --group=0 --numeric-owner \
      --mtime="@$(git -C ... log -1 --format=%ct <commit>)" \
      -czf moonlight-qt-<pkgver>.tar.gz moonlight-qt-<pkgver>
  ```
- 差し替え時は必ず新しい tag を切って sha256 を取り直すこと

### mtime を未来にしてはいけない (実際に踏んだ)

`--mtime` に UTC 固定値 (`2026-09-15 00:00:00Z`) を使ったところ、JST では 09:00 となり
**ビルド時刻より未来**になった。qmake が生成した `Makefile` より `.pro` が新しいと
判定され続け、`Makefile: *.pro` の再生成ルールが毎回発火して qmake ↔ make が無限ループ
した (`make` が CPU を食い続けビルドが完了しない)。

対処: `--mtime="@<commit の author date>"` を使う。commit 日時なので定義上必ず過去になり、
同じ commit からは同じ値が得られる。**同種の自前 tarball を作る package では同じ罠を踏む**
ので、`--mtime` に固定日付を書かないこと。

## 既知の制限

- ホスト側は複数ディスプレイ配信に対応した `sunshine` fork が必要。本 repo の
  `sunshine-bin` (upstream 公式) では 1 ディスプレイのみ
- 追加ウィンドウでのキーボード/ゲームパッド入力、ペン/タッチのディスプレイ指定は未実装

## 更新履歴

- 6.1.0.nekono1 (2026-09-15): 初回追加。fork commit `517dd3be`
- 6.1.0.nekono2 (2026-09-15): fork commit `c70c5e92`。nekono1 は
  `MOONLIGHT_VIDEO_STREAMS` 無しの通常 (単一ディスプレイ) 接続で、stream 開始直後に
  `Session::computeStreamWindowLayout()` が null を参照して SIGSEGV していた。
  moonlight-common-c が video stream 数の正規化を自分の config コピーにしか
  掛けないため、Session 側の数が 0 のまま残っていたのが原因。superproject の
  この 1 commit (`session.cpp` のみ) 以外に source の差分なし、submodule も不変。
  PKGBUILD の差分は `pkgver` / `_reltag` / `sha256sums` のみ
- 6.1.0.nekono3 (2026-09-15): fork commit `b87099a5`。2 枚目以降のストリーム
  ウィンドウが `SDL_WINDOW_ALLOW_HIGHDPI` 無しで作られており、出力スケーリングの
  掛かったデスクトップ (ayaka: 4K @ 1.75x) では論理サイズで描画されてコンポジタに
  引き伸ばされ、文字が読めないほどぼやけていた。1 枚目と同じフラグで作るよう修正。
  superproject のこの 1 commit (`session.cpp` のみ) 以外に source の差分なし、
  submodule も不変。PKGBUILD の差分は `pkgver` / `_reltag` / `sha256sums` のみ
