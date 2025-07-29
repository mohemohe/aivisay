# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AiviSay is a command-line text-to-speech (TTS) tool that supports both local (AivisSpeech Engine) and cloud (Aivis Cloud API) modes. The tool is written in Bash and provides caching capabilities for improved performance.

## Common Commands

### Testing the TTS functionality

```bash
# Basic test (local mode)
./say "テストメッセージ"

# Test with cache
./say --cache "キャッシュテスト"

# Test cloud mode
export AIVIS_SOURCE=cloud
export AIVIS_CLOUD_API_KEY=your_api_key
export AIVIS_CLOUD_MODEL_UUID=your_model_uuid
./say "クラウドテスト"

# List available speakers (local mode only)
./list-speakers
```

### Development and Debugging

```bash
# Run with debug output (script has -x flag)
bash -x ./say "デバッグメッセージ"

# Check cache files
ls -la ~/.cache/aivisay/*/

# Clear cache
rm -rf ~/.cache/aivisay/
```

## Architecture

### Mode Switching
The tool operates in two modes controlled by `AIVIS_SOURCE`:
- `local` (default): Uses AivisSpeech Engine running locally
- `cloud`: Uses Aivis Cloud API

### API Workflows

**Local Mode**:
1. `/audio_query` endpoint with text parameter to generate audio query
2. Modify query parameters (speed, pitch, volume) using jq
3. `/synthesis` endpoint with modified query to generate WAV audio

**Cloud Mode**:
1. Direct POST to `/v1/tts/synthesize` with JSON body containing text and model UUID
2. Returns MP3 audio directly

### Caching System
- Cache directory structure: `~/.cache/aivisay/<speaker_id_or_model_uuid>/<sha256_hash>.<wav|mp3>`
- Text is hashed using SHA256 (sha256sum or shasum -a 256 on macOS)
- Each line of input is cached separately
- Cache is only used when `--cache` flag is provided

### Key Implementation Details

1. **Error Status Handling**: When checking cache, use separate variable declaration and assignment to preserve exit status:
   ```bash
   local cached_file
   cached_file=$(check_cache "$text")
   if [ $? -eq 0 ]; then
   ```

2. **Audio Streaming**: Uses `tee` with process substitution to simultaneously save to cache and play audio:
   ```bash
   curl ... | tee >(save_to_cache "$text") | play -q -t wav -
   ```

3. **Cross-platform Compatibility**: Handles both sha256sum (Linux) and shasum -a 256 (macOS)

## Environment Variables

Critical variables that affect behavior:
- `AIVIS_SOURCE`: Switches between local/cloud mode
- `AIVIS_CACHE_DIR`: Cache location (default: ~/.cache/aivisay)
- Local mode: `AIVISSPEECH_URL`, `AIVISSPEECH_SPEAKER`
- Cloud mode: `AIVIS_CLOUD_API_KEY`, `AIVIS_CLOUD_MODEL_UUID`

Note: The script includes `-x` flag for debugging. Remove it in production for cleaner output.