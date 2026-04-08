package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

const (
	groqChatURL = "https://api.groq.com/openai/v1/chat/completions"
	groqSTTURL  = "https://api.groq.com/openai/v1/audio/transcriptions"
)

const parserSystemPromptTpl = `You are a financial transaction parser.
Extract transaction data from Russian speech transcript and return ONLY valid JSON.
If the transcript does NOT contain any financial transaction (e.g. general conversation, questions like "How are you"), return an empty JSON object: {}.

Current date for resolving relative days: %s

Allowed values (STRICT ENFORCEMENT):
- Types: %s
- Categories: %s
- Currencies: %s

Output schema:
{
  "value": <number, positive float>,
  "type": <string from allowed types>,
  "currency": <string from allowed currencies>,
  "category": <string from allowed categories>,
  "title": <string, merchant or item name in NOMINATIVE case, CAPITALIZED>,
  "description": <string, logical sentence in Russian, CAPITALIZED>,
  "date": <string, "YYYY-MM-DD">
}

Rules:
- title: Nominative case, starts with a CAPITAL letter.
- description: Logical Russian sentence, starts with a CAPITAL letter. Max 100 chars.
- If no date is mentioned, use the current date provided above.`

type groqChatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
	TopP        float64       `json:"top_p"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqChatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

type parsedDraft struct {
	Value       float64 `json:"value"`
	Type        string  `json:"type"`
	Currency    string  `json:"currency"`
	Category    string  `json:"category"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
}

type groqTranscriptResponse struct {
	Text string `json:"text"`
}

type groqErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

type GroqClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewGroqClient(apiKey, proxyStr string) *GroqClient {
	transport := &http.Transport{}
	if proxyStr != "" {
		if proxyURL, err := url.Parse(proxyStr); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	return &GroqClient{
		apiKey: strings.TrimSpace(apiKey),
		httpClient: &http.Client{
			Timeout:   60 * time.Second,
			Transport: transport,
		},
	}
}

func (c *GroqClient) Transcribe(ctx context.Context, audioData []byte, filename string) (string, error) {
	log := logger.GetLoggerWIthRequestId(ctx)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("model", "whisper-large-v3-turbo"); err != nil {
		log.Error("transcription: failed to write model field", zap.Error(err))
		return "", fmt.Errorf("write model field: %w", err)
	}

	if filename == "" {
		filename = "voice.webm"
	}

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		log.Error("transcription: failed to create form file", zap.Error(err))
		return "", fmt.Errorf("create form file: %w", err)
	}

	if _, err = io.Copy(part, bytes.NewReader(audioData)); err != nil {
		log.Error("transcription: failed to write audio data", zap.Error(err))
		return "", fmt.Errorf("write audio data: %w", err)
	}

	if err = writer.Close(); err != nil {
		log.Error("transcription: failed to close multipart writer", zap.Error(err))
		return "", fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, groqSTTURL, body)
	if err != nil {
		log.Error("transcription: failed to build request", zap.Error(err))
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error("transcription: request failed", zap.Error(err))
		return "", fmt.Errorf("groq stt request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		log.Error("transcription: failed to read response body", zap.Error(err))
		return "", fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var groqErr groqErrorResponse
		if jsonErr := json.Unmarshal(respBody, &groqErr); jsonErr != nil {
			log.Error("transcription: groq api error (non-json body)",
				zap.Int("status", resp.StatusCode),
				zap.ByteString("body", respBody))
			return "", ErrInternalClient
		}

		log.Error("transcription: groq api error",
			zap.Int("status", resp.StatusCode),
			zap.String("error_type", groqErr.Error.Type),
			zap.String("error_code", groqErr.Error.Code),
			zap.String("error_message", groqErr.Error.Message))

		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			return "", ErrClientRateLimit
		case http.StatusBadRequest:
			return "", ErrClientInvalidFile
		case http.StatusUnauthorized, http.StatusForbidden:
			return "", ErrClientUnauthorized
		default:
			return "", ErrInternalClient
		}
	}

	var result groqTranscriptResponse
	if err = json.Unmarshal(respBody, &result); err != nil {
		log.Error("transcription: failed to unmarshal response",
			zap.Error(err),
			zap.ByteString("body", respBody))
		return "", fmt.Errorf("parse groq response: %w", err)
	}

	if result.Text == "" {
		log.Warn("transcription: groq returned empty transcript",
			zap.Int("audio_bytes", len(audioData)))
		return "", ErrClientEmptyResult
	}

	log.Info("transcription: success",
		zap.Int("audio_bytes", len(audioData)),
		zap.String("text", result.Text))

	return result.Text, nil
}

func (c *GroqClient) ParseTransaction(ctx context.Context, transcript string, types, categories, currencies []string) (*models.TransactionDraft, error) {
	log := logger.GetLoggerWIthRequestId(ctx)

	if strings.TrimSpace(transcript) == "" {
		return nil, fmt.Errorf("empty transcript")
	}

	typesStr := strings.Join(types, ", ")
	categoriesStr := strings.Join(categories, ", ")
	currenciesStr := strings.Join(currencies, ", ")
	currentDate := time.Now().Format("2006-01-02")

	systemPrompt := fmt.Sprintf(parserSystemPromptTpl, currentDate, typesStr, categoriesStr, currenciesStr)

	payload := groqChatRequest{
		Model: "llama-3.3-70b-versatile",
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: transcript},
		},
		Temperature: 0.1,
		MaxTokens:   512,
		TopP:        1,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		log.Error("parser: failed to marshal payload", zap.Error(err))
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, groqChatURL, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Error("parser: failed to create request", zap.Error(err))
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error("parser: groq request failed", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("parser: failed to read response body", zap.Error(err))
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Error("parser: groq api error", zap.Int("status", resp.StatusCode))
		return nil, ErrInternalClient
	}

	var chatResp groqChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		log.Error("parser: failed to unmarshal response body", zap.Error(err))
		return nil, err
	}

	if len(chatResp.Choices) == 0 {
		return nil, ErrInternalClient
	}

	rawJSON := stripMarkdownFences(chatResp.Choices[0].Message.Content)

	if rawJSON == "{}" || rawJSON == "" {
		return nil, nil
	}

	var parsed parsedDraft
	if err = json.Unmarshal([]byte(rawJSON), &parsed); err != nil {
		log.Error("parser: failed to parse LLM json", zap.String("raw", rawJSON), zap.Error(err))
		return nil, err
	}

	transactionDate, err := time.Parse("2006-01-02", parsed.Date)
	if err != nil {
		transactionDate = time.Now()
	}

	return &models.TransactionDraft{
		RawText:     transcript,
		Value:       parsed.Value,
		Type:        parsed.Type,
		Category:    parsed.Category,
		Currency:    parsed.Currency,
		Title:       parsed.Title,
		Description: parsed.Description,
		Date:        transactionDate,
	}, nil
}

func stripMarkdownFences(s string) string {
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
