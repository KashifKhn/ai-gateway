package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kashifkhan/ai-gateway/internal/adapters"
	"github.com/kashifkhan/ai-gateway/internal/models"
)

type Handler struct {
	registry  *adapters.Registry
	startTime time.Time
	version   string
}

func NewHandler(registry *adapters.Registry, version string) *Handler {
	return &Handler{
		registry:  registry,
		startTime: time.Now(),
		version:   version,
	}
}

func (h *Handler) Health(c *gin.Context) {
	backends := make(map[string]string)
	for _, adapter := range h.registry.List() {
		if adapter.IsHealthy() {
			backends[adapter.ID()] = "connected"
		} else {
			backends[adapter.ID()] = "disconnected"
		}
	}

	c.JSON(http.StatusOK, models.HealthResponse{
		Status:   "healthy",
		Version:  h.version,
		Backends: backends,
		Uptime:   int64(time.Since(h.startTime).Seconds()),
	})
}

func (h *Handler) ListModels(c *gin.Context) {
	var allModels []models.Model

	for _, adapter := range h.registry.List() {
		adapterModels, err := adapter.ListModels()
		if err != nil {
			continue
		}
		allModels = append(allModels, adapterModels...)
	}

	c.JSON(http.StatusOK, models.ModelsResponse{
		Object: "list",
		Data:   allModels,
	})
}

func (h *Handler) ListBackends(c *gin.Context) {
	var backends []models.Backend

	defaultAdapter, _ := h.registry.GetDefault()

	for _, adapter := range h.registry.List() {
		adapterModels, _ := adapter.ListModels()
		modelIDs := make([]string, 0, len(adapterModels))
		for _, m := range adapterModels {
			modelIDs = append(modelIDs, m.ID)
		}

		status := "inactive"
		if adapter.IsHealthy() {
			status = "active"
		}

		backends = append(backends, models.Backend{
			ID:      adapter.ID(),
			Name:    adapter.Name(),
			Status:  status,
			Models:  modelIDs,
			Default: defaultAdapter != nil && defaultAdapter.ID() == adapter.ID(),
		})
	}

	c.JSON(http.StatusOK, models.BackendsResponse{
		Object: "list",
		Data:   backends,
	})
}

func (h *Handler) ChatCompletions(c *gin.Context) {
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr := models.ErrInvalidMessages()
		c.JSON(apiErr.GetStatus(), apiErr)
		return
	}

	if len(req.Messages) == 0 {
		apiErr := models.ErrInvalidMessages()
		c.JSON(apiErr.GetStatus(), apiErr)
		return
	}

	model := req.Model
	if req.Backend != "" {
		model = req.Backend + "/" + req.Model
	}

	adapter, resolvedModel, err := h.registry.FindAdapterForModel(model)
	if err != nil {
		apiErr := models.ErrInvalidModel(req.Model)
		c.JSON(apiErr.GetStatus(), apiErr)
		return
	}

	if !adapter.IsHealthy() {
		apiErr := models.ErrBackendUnavailable(adapter.ID())
		c.JSON(apiErr.GetStatus(), apiErr)
		return
	}

	req.Model = resolvedModel

	if req.Stream {
		h.handleStreamingChat(c, adapter, &req)
	} else {
		h.handleNonStreamingChat(c, adapter, &req)
	}
}

func (h *Handler) handleNonStreamingChat(c *gin.Context, adapter adapters.Adapter, req *models.ChatRequest) {
	resp, err := adapter.Chat(c.Request.Context(), req)
	if err != nil {
		apiErr := models.NewAPIError(
			err.Error(),
			models.ErrorTypeBackend,
			models.ErrorCodeBackendUnavailable,
			500,
		)
		c.JSON(apiErr.GetStatus(), apiErr)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) handleStreamingChat(c *gin.Context, adapter adapters.Adapter, req *models.ChatRequest) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	chunks, errs := adapter.ChatStream(c.Request.Context(), req)

	c.Stream(func(w io.Writer) bool {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				fmt.Fprintf(w, "data: [DONE]\n\n")
				return false
			}

			data, _ := json.Marshal(chunk)
			fmt.Fprintf(w, "data: %s\n\n", string(data))
			c.Writer.Flush()
			return true

		case err, ok := <-errs:
			if ok && err != nil {
				errData := map[string]interface{}{
					"error": map[string]interface{}{
						"message": err.Error(),
						"type":    "backend_error",
					},
				}
				data, _ := json.Marshal(errData)
				fmt.Fprintf(w, "data: %s\n\n", string(data))
			}
			fmt.Fprintf(w, "data: [DONE]\n\n")
			return false

		case <-c.Request.Context().Done():
			return false
		}
	})
}
