package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

const groqChatURL = "https://api.groq.com/openai/v1/chat/completions"

const parserSystemPrompt = `You are a financial transaction parser.
Extract transaction data from Russian speech transcript and return ONLY valid JSON.
No explanation, no markdown, no code blocks — raw JSON only.

Output schema:
{
  "value": <number, positive float>,
  "type": <EXACTLY one of: "expense", "income">,
  "currency": <EXACTLY one of: "RUB", "USD", "EUR">,
  "category": <EXACTLY one of: "groceries", "restaurant", "transport", "entertainment", "health", "utilities", "shopping", "salary", "transfer", "other">,
  "title": <string, 2-4 words in Russian, merchant name or short action>,
  "description": <string, full context in Russian, max 100 chars>
}

Rules:
- value is always positive regardless of expense/income
- If value cannot be determined, set 0
- Use ONLY the enum values listed above — never invent new ones
- Default type: "expense", default currency: "RUB"`

type groqChatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
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
}

// ParserService превращает транскрипцию в черновик транзакции
// через второй запрос в Groq LLaMA.
type ParserService struct {
	apiKey     string
	httpClient *http.Client
}

// NewParserService создаёт сервис парсинга транскрипций.
func NewParserService(apiKey string) *ParserService {
	return &ParserService{
		apiKey: strings.TrimSpace(apiKey),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// ParseTransaction отправляет транскрипцию в Groq LLaMA и возвращает
// структурированный черновик транзакции с полями, совместимыми с DB-енумами.
func (s *ParserService) ParseTransaction(ctx context.Context, transcript string) (*models.TransactionDraft, error) {
	log := logger.GetLoggerWIthRequestId(ctx)

	if strings.TrimSpace(transcript) == "" {
		return nil, fmt.Errorf("empty transcript")
	}

	payload := groqChatRequest{
		Model: "llama-3.3-70b-versatile",
		Messages: []chatMessage{
			{Role: "system", Content: parserSystemPrompt},
			{Role: "user", Content: transcript},
		},
		Temperature: 0,
		MaxTokens:   256,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, groqChatURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "curl/8.5.0")
	req.Header.Set("Accept", "*/*")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Error("parser: groq chat request failed", zap.Error(err))
		return nil, fmt.Errorf("groq chat request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 512<<10))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Error("parser: groq returned error status",
			zap.Int("status", resp.StatusCode),
			zap.ByteString("body", respBody))
		return nil, fmt.Errorf("groq chat error %d: %s", resp.StatusCode, respBody)
	}

	var chatResp groqChatResponse
	if err = json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("parse chat response: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in groq response")
	}

	raw := stripMarkdownFences(chatResp.Choices[0].Message.Content)

	var parsed parsedDraft
	if err = json.Unmarshal([]byte(raw), &parsed); err != nil {
		log.Error("parser: failed to unmarshal llm json",
			zap.String("raw", raw),
			zap.Error(err))
		return nil, fmt.Errorf("parse llm json: %w", err)
	}
	if parsed.Value <= 0 {
		return nil, fmt.Errorf("amount not found in: %q", transcript)
	}

	if parsed.Currency == "" {
		parsed.Currency = "RUB"
	}
	if parsed.Type == "" {
		parsed.Type = "expense"
	}

	log.Info("parser: success",
		zap.Float64("value", parsed.Value),
		zap.String("type", parsed.Type),
		zap.String("category", parsed.Category))

	return &models.TransactionDraft{
		RawText:     transcript,
		Value:       parsed.Value,
		Type:        parsed.Type,
		Category:    parsed.Category,
		Currency:    parsed.Currency,
		Title:       parsed.Title,
		Description: parsed.Description,
	}, nil
}

func stripMarkdownFences(s string) string {
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
