package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

const groqChatURL = "https://api.groq.com/openai/v1/chat/completions"

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

type ParserService struct {
	apiKey     string
	httpClient *http.Client
	enums      EnumsUseCase
}

func NewParserService(apiKey, proxyStr string, enums EnumsUseCase) *ParserService {
	var transport *http.Transport
	if proxyStr != "" {
		if proxyURL, err := url.Parse(proxyStr); err == nil {
			transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	}
	return &ParserService{
		apiKey:     strings.TrimSpace(apiKey),
		enums:      enums,
		httpClient: &http.Client{Timeout: 30 * time.Second, Transport: transport},
	}
}

func (s *ParserService) ParseTransaction(ctx context.Context, transcript string) (*models.TransactionDraft, error) {
	log := logger.GetLoggerWIthRequestId(ctx)

	if strings.TrimSpace(transcript) == "" {
		return nil, fmt.Errorf("empty transcript")
	}

	types := strings.Join(s.enums.GetTransactionTypes(), ", ")
	categories := strings.Join(s.enums.GetCategoryTypes(), ", ")
	currencies := strings.Join(s.enums.GetCurrencyCodes(), ", ")
	currentDate := time.Now().Format("2006-01-02")

	systemPrompt := fmt.Sprintf(parserSystemPromptTpl, currentDate, types, categories, currencies)

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

	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, groqChatURL, bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Error("parser: groq request failed", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Error("parser: groq api error", zap.Int("status", resp.StatusCode))
		return nil, InternalParserError
	}

	var chatResp groqChatResponse
	json.Unmarshal(respBody, &chatResp)

	if len(chatResp.Choices) == 0 {
		return nil, InternalParserError
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
