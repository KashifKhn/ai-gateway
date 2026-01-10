package opencode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kashifkhan/ai-gateway/internal/adapters"
	"github.com/kashifkhan/ai-gateway/internal/config"
	"github.com/kashifkhan/ai-gateway/internal/models"
)

type Adapter struct {
	adapters.BaseAdapter
	config     config.BackendConfig
	httpClient *http.Client
	baseURL    string
	models     map[string]config.ModelConfig
	aliases    map[string]string
}

func New(cfg config.BackendConfig) *Adapter {
	return &Adapter{
		BaseAdapter: adapters.BaseAdapter{},
		config:      cfg,
		models:      make(map[string]config.ModelConfig),
		aliases:     make(map[string]string),
	}
}

func (a *Adapter) ID() string {
	return "opencode"
}

func (a *Adapter) Name() string {
	return "OpenCode"
}

func (a *Adapter) Initialize(cfg map[string]interface{}) error {
	a.baseURL = fmt.Sprintf("http://%s:%d", a.config.Host, a.config.Port)

	a.httpClient = &http.Client{
		Timeout: a.config.Timeout,
	}

	for _, m := range a.config.Models {
		a.models[m.ID] = m
		for _, alias := range m.Aliases {
			a.aliases[alias] = m.ID
		}
	}

	if err := a.HealthCheck(); err != nil {
		a.SetHealthy(false)
		return nil
	}

	a.SetHealthy(true)
	return nil
}

func (a *Adapter) Shutdown() error {
	a.httpClient.CloseIdleConnections()
	return nil
}

func (a *Adapter) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/session", nil)
	if err != nil {
		return err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		a.SetHealthy(false)
		return err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if resp.StatusCode != http.StatusOK || !strings.Contains(contentType, "application/json") {
		a.SetHealthy(false)
		return fmt.Errorf("health check failed: status %d, content-type %s", resp.StatusCode, contentType)
	}

	a.SetHealthy(true)
	return nil
}

func (a *Adapter) ListModels() ([]models.Model, error) {
	result := make([]models.Model, 0, len(a.config.Models))
	for _, m := range a.config.Models {
		result = append(result, models.Model{
			ID:      m.ID,
			Object:  "model",
			Created: time.Now().Unix(),
			OwnedBy: "opencode",
			Backend: "opencode",
			Free:    m.Free,
		})
	}
	return result, nil
}

func (a *Adapter) SupportsModel(modelID string) bool {
	if _, ok := a.models[modelID]; ok {
		return true
	}
	if _, ok := a.aliases[modelID]; ok {
		return true
	}
	return false
}

func (a *Adapter) ResolveModel(modelID string) string {
	if actual, ok := a.aliases[modelID]; ok {
		return actual
	}
	return modelID
}

func (a *Adapter) SupportsStreaming() bool {
	return true
}

func (a *Adapter) Chat(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	sessionID, err := a.createSession(ctx, req.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer a.deleteSession(sessionID)

	var userMessage string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			userMessage = req.Messages[i].Content
			break
		}
	}

	if userMessage == "" {
		return nil, fmt.Errorf("no user message found")
	}

	fullContent, err := a.sendMessageNonStreaming(ctx, sessionID, userMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return &models.ChatResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", sessionID),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []models.Choice{
			{
				Index: 0,
				Message: models.Message{
					Role:    "assistant",
					Content: fullContent,
				},
				FinishReason: "stop",
			},
		},
	}, nil
}

func (a *Adapter) ChatStream(ctx context.Context, req *models.ChatRequest) (<-chan models.StreamChunk, <-chan error) {
	chunks := make(chan models.StreamChunk, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)

		sessionID, err := a.createSession(ctx, req.Model)
		if err != nil {
			errs <- fmt.Errorf("failed to create session: %w", err)
			return
		}
		defer a.deleteSession(sessionID)

		var userMessage string
		for i := len(req.Messages) - 1; i >= 0; i-- {
			if req.Messages[i].Role == "user" {
				userMessage = req.Messages[i].Content
				break
			}
		}

		if userMessage == "" {
			errs <- fmt.Errorf("no user message found")
			return
		}

		err = a.sendMessageStreaming(ctx, sessionID, userMessage, chunks)
		if err != nil {
			errs <- err
		}
	}()

	return chunks, errs
}

func (a *Adapter) createSession(ctx context.Context, model string) (string, error) {
	reqBody := map[string]interface{}{}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/session", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create session: %s", string(bodyBytes))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.ID, nil
}

func (a *Adapter) deleteSession(sessionID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "DELETE", a.baseURL+"/session/"+sessionID, nil)
	resp, err := a.httpClient.Do(req)
	if err == nil {
		resp.Body.Close()
	}
}

type OpenCodeResponse struct {
	Info struct {
		ID string `json:"id"`
	} `json:"info"`
	Parts []struct {
		Type string `json:"type"`
		Text string `json:"text,omitempty"`
	} `json:"parts"`
	Error   interface{} `json:"error,omitempty"`
	Success bool        `json:"success"`
}

func (a *Adapter) sendMessageNonStreaming(ctx context.Context, sessionID, message string) (string, error) {
	reqBody := map[string]interface{}{
		"parts": []map[string]interface{}{
			{
				"type": "text",
				"text": message,
			},
		},
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/session/"+sessionID+"/message", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var ocResp OpenCodeResponse
	if err := json.Unmarshal(bodyBytes, &ocResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w (body: %s)", err, string(bodyBytes))
	}

	if ocResp.Error != nil && !ocResp.Success {
		return "", fmt.Errorf("opencode error: %v", ocResp.Error)
	}

	var fullContent strings.Builder
	for _, part := range ocResp.Parts {
		if part.Type == "text" && part.Text != "" {
			fullContent.WriteString(part.Text)
		}
	}

	return fullContent.String(), nil
}

func (a *Adapter) sendMessageStreaming(ctx context.Context, sessionID, message string, chunks chan<- models.StreamChunk) error {
	reqBody := map[string]interface{}{
		"parts": []map[string]interface{}{
			{
				"type": "text",
				"text": message,
			},
		},
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/session/"+sessionID+"/message", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	chunkID := fmt.Sprintf("chatcmpl-%s", sessionID)
	created := time.Now().Unix()

	chunks <- models.StreamChunk{
		ID:      chunkID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   "opencode",
		Choices: []models.ChunkChoice{
			{
				Index: 0,
				Delta: models.Delta{
					Role: "assistant",
				},
			},
		},
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var ocResp OpenCodeResponse
	if err := json.Unmarshal(bodyBytes, &ocResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if ocResp.Error != nil && !ocResp.Success {
		return fmt.Errorf("opencode error: %v", ocResp.Error)
	}

	for _, part := range ocResp.Parts {
		if part.Type == "text" && part.Text != "" {
			chunks <- models.StreamChunk{
				ID:      chunkID,
				Object:  "chat.completion.chunk",
				Created: created,
				Model:   "opencode",
				Choices: []models.ChunkChoice{
					{
						Index: 0,
						Delta: models.Delta{
							Content: part.Text,
						},
					},
				},
			}
		}
	}

	chunks <- models.StreamChunk{
		ID:      chunkID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   "opencode",
		Choices: []models.ChunkChoice{
			{
				Index:        0,
				Delta:        models.Delta{},
				FinishReason: "stop",
			},
		},
	}

	return nil
}
