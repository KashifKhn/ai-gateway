package adapters

import (
	"context"

	"github.com/kashifkhan/ai-gateway/internal/models"
)

type Adapter interface {
	ID() string
	Name() string

	Initialize(config map[string]interface{}) error
	Shutdown() error

	HealthCheck() error
	IsHealthy() bool

	ListModels() ([]models.Model, error)
	SupportsModel(modelID string) bool
	ResolveModel(modelID string) string

	Chat(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error)
	ChatStream(ctx context.Context, req *models.ChatRequest) (<-chan models.StreamChunk, <-chan error)

	SupportsStreaming() bool
	SupportsTools() bool
	SupportsSessions() bool
}

type BaseAdapter struct {
	id      string
	name    string
	healthy bool
}

func (a *BaseAdapter) ID() string {
	return a.id
}

func (a *BaseAdapter) Name() string {
	return a.name
}

func (a *BaseAdapter) IsHealthy() bool {
	return a.healthy
}

func (a *BaseAdapter) SetHealthy(healthy bool) {
	a.healthy = healthy
}

func (a *BaseAdapter) SupportsTools() bool {
	return false
}

func (a *BaseAdapter) SupportsSessions() bool {
	return false
}
