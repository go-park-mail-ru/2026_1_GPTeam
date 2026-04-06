package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

const groqSTTURL = "https://api.groq.com/openai/v1/audio/transcriptions"

type groqTranscriptResponse struct {
	Text string `json:"text"`
}

type TranscriptionService struct {
	apiKey     string
	httpClient *http.Client
}

func NewTranscriptionService(apiKey string) *TranscriptionService {
	return &TranscriptionService{
		apiKey: strings.TrimSpace(apiKey),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Transcribe отправляет аудио в Groq Whisper и возвращает текст транскрипции.
func (s *TranscriptionService) Transcribe(ctx context.Context, audioData []byte, filename string) (string, error) {
	log := logger.GetLoggerWIthRequestId(ctx)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".webm"
	}

	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="voice%s"`, ext))
	h.Set("Content-Type", audioMIMEByExt(ext))

	filePart, err := writer.CreatePart(h)
	if err != nil {
		log.Error("transcription: failed to create multipart part", zap.Error(err))
		return "", fmt.Errorf("create multipart part: %w", err)
	}
	if _, err = filePart.Write(audioData); err != nil {
		log.Error("transcription: failed to write audio data", zap.Error(err))
		return "", fmt.Errorf("write audio: %w", err)
	}

	_ = writer.WriteField("model", "whisper-large-v3")
	_ = writer.WriteField("language", "ru")
	_ = writer.WriteField("response_format", "json")
	_ = writer.WriteField("temperature", "0")

	if err = writer.Close(); err != nil {
		return "", fmt.Errorf("close writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, groqSTTURL, body)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "curl/8.5.0")
	req.Header.Set("Accept", "*/*")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Error("transcription: groq request failed", zap.Error(err))
		return "", fmt.Errorf("groq stt request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read groq response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Error("transcription: groq returned error status",
			zap.Int("status", resp.StatusCode),
			zap.ByteString("body", respBody))
		return "", fmt.Errorf("groq stt error %d: %s", resp.StatusCode, respBody)
	}

	var result groqTranscriptResponse
	if err = json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse groq response: %w", err)
	}
	if result.Text == "" {
		return "", fmt.Errorf("groq returned empty transcript")
	}

	log.Info("transcription: success", zap.String("text", result.Text))
	return result.Text, nil
}

func audioMIMEByExt(ext string) string {
	switch ext {
	case ".ogg":
		return "audio/ogg"
	case ".mp4", ".m4a":
		return "audio/mp4"
	case ".wav":
		return "audio/wav"
	case ".flac":
		return "audio/flac"
	default:
		return "audio/webm"
	}
}
