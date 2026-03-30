# TmpLink アップローダー

[TmpLink](https://tmp.link/) 向けのシンプルで強力なファイルアップロードツールです。大容量ファイル・レジューム・一括アップロードに対応しています。

[English](README.en.md) | [简体中文](README.zh-CN.md) | **日本語**

## 概要

### CLI でファイルをアップロード

![CLI アップローダー](docs/images/tmplink-cli.webp)

### TUI でファイルをアップロード

![TUI アップローダー](docs/images/tmplink.webp)

## 機能

- **ワンクリックアップロード** — 最大 50GB まで対応
- **高速転送** — チャンクアップロード + マルチスレッドで帯域幅を最大活用
- **レジューム機能** — ネットワーク障害からの自動復旧
- **デュアルインターフェース** — 日常利用向けの TUI と、スクリプト自動化向けの CLI
- **メンバー機能** — スポンサーユーザーは追加の高度な設定を利用可能

## クイックスタート

### インストール

#### ワンクリックインストール（推奨）

**方法 1: オンラインインストールスクリプト**

Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/tmplink/tmplink_uploader/main/install-linux.sh | bash
```

macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/tmplink/tmplink_uploader/main/install-macos.sh | bash
```

Windows（管理者として PowerShell を起動）:

```powershell
iex ((New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/tmplink/tmplink_uploader/main/install-windows.ps1'))
```

**方法 2: クローンしてインストール**

```bash
git clone https://github.com/tmplink/tmplink_uploader.git
cd tmplink_uploader
```

Linux:

```bash
./install-linux.sh
```

macOS:

```bash
./install-macos.sh
```

Windows:

```powershell
.\install-windows.ps1
```

インストール後、`tmplink` と `tmplink-cli` がシステム全体で使用可能になります。

#### 手動ダウンロード

ビルド済みバイナリは [build](build/) ディレクトリからダウンロードできます:

- **Windows**: `windows-64bit` または `windows-32bit`
- **macOS**: `macos-arm64` (M1/M2) または `macos-intel`
- **Linux**: `linux-64bit`、`linux-32bit`、または `linux-arm64`

### アクセストークンの取得

1. [TmpLink](https://tmp.link/) を開いてログイン
2. 「ファイルをアップロード」をクリックし、「リセット」を選択
3. 「CLI アップロード」セクションからトークンをコピー

### 使い方

#### TUI（初心者に推奨）

```bash
# スクリプトでインストールした場合:
tmplink

# 手動ダウンロードの場合:
# Windows
tmplink.exe

# macOS/Linux
./tmplink
```

初回起動時にトークンの入力を求められます。以降はインターフェースでファイルを選択してアップロードできます。

#### CLI（上級者向け）

```bash
# トークンを保存（初回のみ）
tmplink-cli -set-token あなたのTOKEN

# ファイルをアップロード
tmplink-cli -file /path/to/file
```

## 使用例

### TUI の操作

起動後、矢印キーで操作します:

- **ファイルを選択** — アップロードするファイルを選ぶ
- **アップロード開始** — 進捗と速度をリアルタイムで確認
- **結果を確認** — ダウンロードリンクをコピー

### CLI の主な使い方

```bash
# 単一ファイルのアップロード
tmplink-cli -file ~/Documents/report.pdf

# 大容量ファイルのアップロード（チャンクサイズを大きく）
tmplink-cli -file ~/Videos/movie.mp4 -chunk-size 10

# ファイルを永久保存
tmplink-cli -file ~/backup.zip -model 99

# 一時的に別のトークンを使用
tmplink-cli -file test.txt -token 一時TOKEN
```

## パラメーター

### 基本パラメーター

| フラグ | 説明 |
| --- | --- |
| `-file` | ファイルパス（必須） |
| `-token` | API トークン（設定ファイルに保存可能） |
| `-chunk-size` | チャンクサイズ（MB）、1〜99（デフォルト: 3） |
| `-model` | 保存期間: 0=24時間、1=3日、2=7日、99=永久 |

### 設定管理

| フラグ | 説明 |
| --- | --- |
| `-set-token` | トークンを設定ファイルに保存 |
| `-set-model` | デフォルト保存期間を設定 |
| `-set-mr-id` | デフォルトアップロードディレクトリを設定 |

## トラブルシューティング

### macOS: 「開発元を確認できません」

```bash
xattr -d com.apple.quarantine tmplink tmplink-cli
```

### Linux/macOS: 実行権限がない

```bash
chmod +x tmplink tmplink-cli
```

### Windows Defender の警告

「詳細情報」→「実行」をクリックするか、バイナリを信頼リストに追加してください。

### アップロード失敗時の確認

1. ネットワーク接続を確認
2. トークンが有効かどうかを確認
3. `-debug` フラグで詳細なエラーを確認

```bash
tmplink-cli -debug -file test.txt
```

## ヘルプ

```bash
# 利用可能なフラグを一覧表示
tmplink-cli -h

# 詳細ログを有効化
tmplink-cli -debug -file yourfile.txt
```

問題が解決しない場合は [Issue を作成](https://github.com/tmplink/tmplink_uploader/issues) してください。

## ライセンス

このプロジェクトは [Apache 2.0](LICENSE) ライセンスで公開されています。
