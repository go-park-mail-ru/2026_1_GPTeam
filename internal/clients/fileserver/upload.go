package fileserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

type Uploader struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

func NewUploader(baseURL, token string) *Uploader {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	return &Uploader{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		token: strings.TrimSpace(token),
	}
}

type uploadResponse struct {
	Filename string `json:"filename"`
}

func (c *Uploader) Upload(ctx context.Context, reader io.Reader, extension string) (string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", "upload"+extension)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, reader); err != nil {
		return "", err
	}
	if err := w.WriteField("extension", extension); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/upload", &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fileserver: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var out uploadResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if out.Filename == "" {
		return "", fmt.Errorf("fileserver: empty filename")
	}
	return out.Filename, nil
}
