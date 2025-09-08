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
例えば、[Claude Code Hooks](https://docs.anthropic.com/en/docs/claude-code/hooks)のように、同じセリフの通知ボイスを何回も流す場合に生成ラグがなくなります。

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

## モデルアップロード（ローカルモードのみ）

合法的なAIVMXファイルをアップロードするwrapperスクリプトを使用します：

```bash
./upload-model --file /path/to/file
```

例: [あけさとさん](https://akesato-goods.booth.pm/items/6636929)

```bash
 ❯ ./upload-model --file ~/Downloads/AivisSpeech-あけさとさんα-/あけさとさんα（ほがらか）.aivmx
Uploading model: /Users/mohemohe/Downloads/AivisSpeech-あけさとさんα-/あけさとさんα（ほがらか）.aivmx
Target URL: http://172.16.34.2:10101/aivm_models/install
✓ Model uploaded successfully!

Note: You may need to load the model using the appropriate UUID.
Use './list-speakers' to see available models and their UUIDs.


>>> elapsed time 7s251ms
```

## 話者の確認（ローカルモードのみ）

利用可能な話者を確認するには、`list-speakers`スクリプトを使用します：

```bash
./list-speakers
```

```bash
 ❯ ./list-speakers
Fetching available speakers from http://172.16.34.2:10101...

Available speakers and styles:
==============================
■ Anneli (Model UUID: e756b8e4-b606-4e15-99b1-3f9c6a1b2317)
  └─ ノーマル (Speaker ID: 888753760, Type: talk)
  └─ 通常 (Speaker ID: 888753761, Type: talk)
  └─ テンション高め (Speaker ID: 888753762, Type: talk)
  └─ 上機嫌 (Speaker ID: 888753764, Type: talk)
  └─ 落ち着き (Speaker ID: 888753763, Type: talk)
  └─ 怒り・悲しみ (Speaker ID: 888753765, Type: talk)
■ あけさとさんα（ひそひそ） (Model UUID: 13632fe5-58e5-41d2-a4e3-d36c9ec4a269)
  └─ Whisper (Speaker ID: 1063129024, Type: talk)
■ あけさとさんα（ほがらか） (Model UUID: 74b5afe2-af67-4140-8e1e-64e49759a20c)
  └─ Bright (Speaker ID: 97713888, Type: talk)
■ あけさとさんα（やわらか） (Model UUID: 24d6399a-0a3f-4560-93f6-68c7967e890e)
  └─ Soft (Speaker ID: 1420750528, Type: talk)

 ❯ ./list-speakers 74b5afe2-af67-4140-8e1e-64e49759a20c
Fetching available speakers from http://172.16.34.2:10101...

Speaker details for UUID: 74b5afe2-af67-4140-8e1e-64e49759a20c
==============================
Name: あけさとさんα（ほがらか）
Model UUID: 74b5afe2-af67-4140-8e1e-64e49759a20c

Available styles:
  • Bright (Speaker ID: 97713888, Type: talk)
```

※上記は出力例であり、Anneliの利用を推奨するものではありません。

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