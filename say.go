///usr/bin/env go run $0 $@ ; exit

package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Configuration
type Config struct {
	Source    string
	LocalURL  string
	SpeakerID string
	CloudURL  string
	APIKey    string
	ModelUUID string
	Speed     float64
	Pitch     float64
	Volume    float64
	CacheDir  string
	SleepTime time.Duration
	Debug     bool
}

// Audio data structure
type AudioData struct {
	Index  int
	Buffer []byte
	Format string
	Error  error
}

// Cloud TTS request
type CloudRequest struct {
	ModelUUID    string `json:"model_uuid"`
	Text         string `json:"text"`
	UseSSML      bool   `json:"use_ssml"`
	OutputFormat string `json:"output_format"`
}

func getConfig() *Config {
	cfg := &Config{
		Source:    getEnv("AIVIS_SOURCE", "local"),
		LocalURL:  getEnv("AIVISSPEECH_URL", "http://localhost:10101"),
		SpeakerID: getEnv("AIVISSPEECH_SPEAKER", "888753760"),
		CloudURL:  getEnv("AIVIS_CLOUD_URL", "https://api.aivis-project.com"),
		APIKey:    getEnv("AIVIS_CLOUD_API_KEY", ""),
		ModelUUID: getEnv("AIVIS_CLOUD_MODEL_UUID", "a59cb814-0083-4369-8542-f51a29e72af7"),
		Speed:     getEnvFloat("AIVISSPEECH_SPEED", 1.0),
		Pitch:     getEnvFloat("AIVISSPEECH_PITCH", 0.0),
		Volume:    getEnvFloat("AIVISSPEECH_VOLUME", 1.0),
		CacheDir:  getEnv("AIVIS_CACHE_DIR", filepath.Join(os.Getenv("HOME"), ".cache", "aivisay")),
		SleepTime: time.Duration(getEnvFloat("AIVIS_SLEEP", 0.3)*1000) * time.Millisecond,
		Debug:     getEnv("AIVIS_DEBUG", "") == "1",
	}
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func debugLog(cfg *Config, format string, args ...interface{}) {
	if cfg.Debug {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

func showHelp() {
	fmt.Printf(`Usage: say.go [OPTIONS] [TEXT]

Convert text to speech using AivisSpeech Engine or Aivis Cloud API.

Options:
  --help    Show this help message
  --cache   Enable caching (saves audio files for faster playback)

Examples:
  say.go "Hello, world"
  echo "Hello" | say.go
  say.go --cache "Cached message"

Environment:
  AIVIS_SOURCE         API mode: 'local' or 'cloud' (default: local)
  AIVIS_CACHE_DIR      Cache directory (default: ~/.cache/aivisay)

For more options, see: https://github.com/mohemohe/aivisay
`)
	os.Exit(0)
}

func checkConnections(cfg *Config) error {
	if cfg.Source == "local" {
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get(cfg.LocalURL + "/speakers")
		if err != nil {
			return fmt.Errorf("Cannot connect to AivisSpeech Engine at %s\nMake sure AivisSpeech Engine is running", cfg.LocalURL)
		}
		resp.Body.Close()
	} else if cfg.Source == "cloud" {
		if cfg.APIKey == "" {
			return fmt.Errorf("AIVIS_CLOUD_API_KEY is not set\nPlease set your API key with: export AIVIS_CLOUD_API_KEY=your_api_key")
		}
		if cfg.ModelUUID == "" {
			return fmt.Errorf("AIVIS_CLOUD_MODEL_UUID is not set\nPlease set your model UUID with: export AIVIS_CLOUD_MODEL_UUID=your_model_uuid")
		}
	} else {
		return fmt.Errorf("Invalid AIVIS_SOURCE. Use 'local' or 'cloud'")
	}
	return nil
}

func getCacheFilename(cfg *Config, text string) string {
	hasher := sha256.New()
	hasher.Write([]byte(text))
	hash := hex.EncodeToString(hasher.Sum(nil))

	var subdir, ext string
	if cfg.Source == "local" {
		subdir = cfg.SpeakerID
		ext = "wav"
	} else {
		subdir = cfg.ModelUUID
		ext = "mp3"
	}

	cacheDir := filepath.Join(cfg.CacheDir, subdir)
	return filepath.Join(cacheDir, hash+"."+ext)
}

func checkCache(cfg *Config, text string, cacheEnabled bool) ([]byte, error) {
	if !cacheEnabled {
		return nil, nil
	}

	filepath := getCacheFilename(cfg, text)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil, nil
	}

	return os.ReadFile(filepath)
}

func saveToCache(cfg *Config, text string, data []byte, cacheEnabled bool) error {
	if !cacheEnabled {
		return nil
	}

	filepath := getCacheFilename(cfg, text)
	dir := filepath[:strings.LastIndex(filepath, "/")]

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

func generateAudioLocal(cfg *Config, text string) ([]byte, error) {
	// Step 1: Create audio query (POST method with data in query params like bash --get --data-urlencode)
	queryURL := fmt.Sprintf("%s/audio_query?speaker=%s&text=%s",
		cfg.LocalURL, cfg.SpeakerID, url.QueryEscape(text))

	req, err := http.NewRequest("POST", queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio query request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio query: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to create audio query (HTTP %d)", resp.StatusCode)
	}

	var audioQuery map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&audioQuery); err != nil {
		return nil, fmt.Errorf("failed to decode audio query: %v", err)
	}

	// Step 2: Modify parameters (keep all other values from audio_query response)
	audioQuery["speedScale"] = cfg.Speed
	audioQuery["pitchScale"] = cfg.Pitch
	audioQuery["volumeScale"] = cfg.Volume
	audioQuery["prePhonemeLength"] = 0
	audioQuery["postPhonemeLength"] = 0

	// Step 3: Synthesize speech
	queryData, err := json.Marshal(audioQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %v", err)
	}

	synthURL := fmt.Sprintf("%s/synthesis?speaker=%s", cfg.LocalURL, cfg.SpeakerID)
	resp, err = http.Post(synthURL, "application/json", bytes.NewBuffer(queryData))
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize audio: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Read response body for error details
		if errorBody, readErr := io.ReadAll(resp.Body); readErr == nil {
			return nil, fmt.Errorf("failed to synthesize audio (HTTP %d): %s", resp.StatusCode, string(errorBody))
		}
		return nil, fmt.Errorf("failed to synthesize audio (HTTP %d)", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func generateAudioCloud(cfg *Config, text string) ([]byte, error) {
	request := CloudRequest{
		ModelUUID:    cfg.ModelUUID,
		Text:         text,
		UseSSML:      false,
		OutputFormat: "mp3",
	}

	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", cfg.CloudURL+"/v1/tts/synthesize", bytes.NewBuffer(requestData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call cloud API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to synthesize audio (HTTP %d)", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func generateAudio(cfg *Config, text string, cacheEnabled bool) ([]byte, error) {
	// Check cache first
	if cached, err := checkCache(cfg, text, cacheEnabled); err == nil && cached != nil {
		return cached, nil
	}

	var audioData []byte
	var err error

	if cfg.Source == "local" {
		audioData, err = generateAudioLocal(cfg, text)
	} else {
		audioData, err = generateAudioCloud(cfg, text)
	}

	if err != nil {
		return nil, err
	}

	// Save to cache
	if cacheEnabled {
		saveToCache(cfg, text, audioData, cacheEnabled)
	}

	return audioData, nil
}

func playAudio(audioData []byte, format string) error {
	cmd := exec.Command("play", "-q", "-t", format, "-")
	cmd.Stdin = bytes.NewReader(audioData)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func splitSentences(text string) []string {
	// Split by Japanese punctuation marks
	re := regexp.MustCompile(`([、。！？!?．])`)
	parts := re.Split(text, -1)
	matches := re.FindAllString(text, -1)

	var sentences []string
	for i, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}

		sentence := strings.TrimSpace(part)
		if i < len(matches) {
			sentence += matches[i]
		}

		if sentence != "" {
			sentences = append(sentences, sentence)
		}
	}

	return sentences
}

func streamingTTS(cfg *Config, text string, cacheEnabled bool) error {
	sentences := splitSentences(text)
	if len(sentences) == 0 {
		return nil
	}

	// For single sentence, just use regular TTS
	if len(sentences) == 1 {
		fmt.Println(sentences[0])
		audioData, err := generateAudio(cfg, sentences[0], cacheEnabled)
		if err != nil {
			return err
		}

		format := "wav"
		if cfg.Source == "cloud" {
			format = "mp3"
		}

		return playAudio(audioData, format)
	}

	// For multiple sentences, use background generation with goroutines
	audioChannel := make(chan AudioData, len(sentences))
	playChannel := make(chan AudioData, len(sentences))

	// Background audio generation goroutine
	go func() {
		defer close(audioChannel)

		// Generate first sentence immediately and send to play channel
		debugLog(cfg, "Generating audio for sentence 1: \"%s\"", sentences[0])
		audioData, err := generateAudio(cfg, sentences[0], cacheEnabled)
		format := "wav"
		if cfg.Source == "cloud" {
			format = "mp3"
		}

		audioChannel <- AudioData{
			Index:  0,
			Buffer: audioData,
			Format: format,
			Error:  err,
		}

		// Generate remaining sentences sequentially
		for i := 1; i < len(sentences); i++ {
			debugLog(cfg, "Generating audio for sentence %d: \"%s\"", i+1, sentences[i])
			audioData, err := generateAudio(cfg, sentences[i], cacheEnabled)
			debugLog(cfg, "Generated audio for sentence %d: \"%s\"", i+1, sentences[i])

			audioChannel <- AudioData{
				Index:  i,
				Buffer: audioData,
				Format: format,
				Error:  err,
			}
		}
	}()

	// Audio playback goroutine
	go func() {
		defer close(playChannel)

		audioBuffer := make(map[int]AudioData)
		nextPlayIndex := 0

		for audio := range audioChannel {
			audioBuffer[audio.Index] = audio

			// Play all consecutive available audio
			for {
				if nextAudio, exists := audioBuffer[nextPlayIndex]; exists {
					playChannel <- nextAudio
					delete(audioBuffer, nextPlayIndex)
					nextPlayIndex++
				} else {
					break
				}
			}
		}
	}()

	// Main playback loop
	for audio := range playChannel {
		sentence := sentences[audio.Index]
		fmt.Println(sentence)

		if audio.Error != nil {
			debugLog(cfg, "Error generating audio for sentence %d: %v", audio.Index+1, audio.Error)
			continue
		}

		if audio.Buffer != nil {
			if err := playAudio(audio.Buffer, audio.Format); err != nil {
				debugLog(cfg, "Error playing audio for sentence %d: %v", audio.Index+1, err)
			}
		}

		// Sleep between sentences (except for last one)
		if audio.Index < len(sentences)-1 {
			time.Sleep(cfg.SleepTime)
		}
	}

	return nil
}

func readStdin() (string, error) {
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(lines, "\n"), nil
}

func main() {
	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-sigChan
		cancel()
		os.Exit(1)
	}()

	cfg := getConfig()
	cacheEnabled := false
	var message string

	// Parse arguments
	args := os.Args[1:]
	for i, arg := range args {
		switch arg {
		case "--help":
			showHelp()
		case "--cache":
			cacheEnabled = true
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "Error: Unknown option: %s\n", arg)
				fmt.Fprintf(os.Stderr, "Use --help for usage information\n")
				os.Exit(1)
			}
			message = strings.Join(args[i:], " ")
			goto parseComplete
		}
	}

parseComplete:
	// Read from stdin if no message provided and stdin is not a terminal
	if message == "" {
		stat, err := os.Stdin.Stat()
		if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
			if stdinText, err := readStdin(); err == nil {
				message = strings.TrimSpace(stdinText)
			}
		}
	}

	if message == "" {
		fmt.Fprintf(os.Stderr, "Error: No text provided\n")
		fmt.Fprintf(os.Stderr, "Usage: say.go \"text to speak\"\n")
		fmt.Fprintf(os.Stderr, "   or: echo \"text\" | say.go\n")
		os.Exit(1)
	}

	// Check connections
	if err := checkConnections(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Process input with streaming TTS
	if err := streamingTTS(cfg, message, cacheEnabled); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
