# AiviSay

AivisSpeech Engine またはAivis Cloud APIを使用したコマンドラインTTS

## 必要なもの

### 共通
- curl
- jq (`brew install jq`)
- sox (`brew install sox`)

### ローカルモード（デフォルト）
- [AivisSpeech Engine](https://github.com/Aivis-Project/AivisSpeech-Engine) が起動していること

### クラウドモード
- Aivis Cloud APIキー（[Aivis Project Hub](https://hub.aivis-project.com/)から取得）
- モデルUUID

## インストール

```bash
chmod +x say
```

パスの通った場所に配置するか、エイリアスを設定してください：

```bash
# ~/.bashrc または ~/.zshrc に追加
alias say='/path/to/aivisay/say'
```

macOSでパスが通った場所に配置する場合は、標準の `say` コマンドよりも優先して読み込まれる場所に配置してください。

## 使い方

### 基本的な使い方

```bash
# テキストを音声に変換
say "爆音が銀世界の高原に広がる"

# パイプからの入力
echo "爆音が銀世界の高原に広がる" | say

# 複数行のテキスト
cat text.txt | say

# ヘルプを表示
say --help
```

### キャッシュ機能

キャッシュを有効にすると、同じテキストの音声合成結果を保存し、2回目以降は高速に再生できます。

```bash
# キャッシュを有効にして実行
say --cache "あらゆる現実をすべて自分のほうへねじ曲げたのだ"

# 2回目以降はキャッシュから高速再生
say --cache "あらゆる現実をすべて自分のほうへねじ曲げたのだ"
```

キャッシュは以下の場所に保存されます：
- デフォルト: `~/.cache/aivisay/<speaker_id_or_model_uuid>/<sha256_hash>.<wav|mp3>`
- 環境変数で変更可能: `export AIVIS_CACHE_DIR=/path/to/cache`

### ローカルモード（デフォルト）

```bash
# 基本的な使い方
say "一週間ばかりニューヨークを取材した"

# キャッシュを使用
say --cache "一週間ばかりニューヨークを取材した"
```

### クラウドモード

```bash
# 必要な環境変数を設定
export AIVIS_SOURCE=cloud
export AIVIS_CLOUD_API_KEY=your_api_key
export AIVIS_CLOUD_MODEL_UUID=your_model_uuid

# 使い方はローカルモードと同じ
say "小さな鰻屋に、熱気のようなものがみなぎる"

# キャッシュも同様に使用可能
say --cache "小さな鰻屋に、熱気のようなものがみなぎる"
```

## 設定

### APIモード選択

```bash
# ローカルモード（デフォルト）
export AIVIS_SOURCE=local

# クラウドモード
export AIVIS_SOURCE=cloud
```

### ローカルモード設定

```bash
# AivisSpeech EngineのURL（デフォルト: http://localhost:10101）
export AIVISSPEECH_URL=http://localhost:10101

# 話者ID（デフォルト: 888753760）
export AIVISSPEECH_SPEAKER=888753760

# 話速（デフォルト: 1.0、範囲: 0.5-2.0）
export AIVISSPEECH_SPEED=1.2

# ピッチ（デフォルト: 0.0、範囲: -0.15-0.15）
export AIVISSPEECH_PITCH=0.1

# 音量（デフォルト: 1.0、範囲: 0.0-2.0）
export AIVISSPEECH_VOLUME=1.5
```

### クラウドモード設定

```bash
# APIキー（必須）
export AIVIS_CLOUD_API_KEY=your_api_key_here

# モデルUUID（必須）
export AIVIS_CLOUD_MODEL_UUID=your_model_uuid_here

# APIエンドポイント（デフォルト: https://api.aivis-project.com）
export AIVIS_CLOUD_URL=https://api.aivis-project.com
```

### 共通設定（現在はローカルモードのみ）

```bash
# 話速（デフォルト: 1.0）
export AIVISSPEECH_SPEED=1.2

# ピッチ（デフォルト: 0.0）
export AIVISSPEECH_PITCH=0.1

# 音量（デフォルト: 1.0）
export AIVISSPEECH_VOLUME=1.5
```

### キャッシュ設定

```bash
# キャッシュディレクトリ（デフォルト: ~/.cache/aivisay）
export AIVIS_CACHE_DIR=/path/to/cache
```

## 話者の確認（ローカルモードのみ）

利用可能な話者を確認するには、`list-speakers`スクリプトを使用します：

```bash
./list-speakers
```

※クラウドモードでは、Aivis Project Hubでモデルを選択してUUIDを取得してください。

## トラブルシューティング

### "Cannot connect to AivisSpeech Engine" エラー

AivisSpeech Engineが起動していることを確認してください。

### "Required command 'jq' not found" エラー

```bash
brew install jq
```

### "Required command 'play' not found" エラー

```bash
brew install sox
```