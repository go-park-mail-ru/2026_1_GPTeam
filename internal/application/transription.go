package application

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
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

func NewTranscriptionService(apiKey, proxyStr string) *TranscriptionService {
	var transport *http.Transport
	if proxyStr != "" {
		if proxyURL, err := url.Parse(proxyStr); err == nil {
			transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	}

	return &TranscriptionService{
		apiKey: strings.TrimSpace(apiKey),
		httpClient: &http.Client{
			Timeout:   60 * time.Second,
			Transport: transport,
		},
	}
}

func (s *TranscriptionService) Transcribe(ctx context.Context, audioData []byte, filename string) (string, error) {
	log := logger.GetLoggerWIthRequestId(ctx)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("model", "whisper-large-v3-turbo")

	if filename == "" {
		filename = "voice.webm"
	}
	part, _ := writer.CreateFormFile("file", filename)
	_, _ = io.Copy(part, bytes.NewReader(audioData))
	_ = writer.Close()

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, groqSTTURL, body)
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Error("transcription: request failed", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Error("transcription: groq api error", zap.Int("status", resp.StatusCode), zap.ByteString("body", respBody))
		return "", InternalTranscriptionError
	}

	var result groqTranscriptResponse
	if err = json.Unmarshal(respBody, &result); err != nil {
		log.Error("transcription: unmarshal failed", zap.Error(err))
		return "", err
	}

	log.Info("transcription: success", zap.Int("bytes", len(audioData)))
	return result.Text, nil
}
