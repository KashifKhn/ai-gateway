package adapters

import (
	"fmt"
	"sync"
)

type Registry struct {
	adapters       map[string]Adapter
	mu             sync.RWMutex
	defaultBackend string
}

func NewRegistry(defaultBackend string) *Registry {
	return &Registry{
		adapters:       make(map[string]Adapter),
		defaultBackend: defaultBackend,
	}
}

func (r *Registry) Register(adapter Adapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[adapter.ID()] = adapter
}

func (r *Registry) Get(id string) (Adapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	adapter, ok := r.adapters[id]
	return adapter, ok
}

func (r *Registry) GetDefault() (Adapter, bool) {
	return r.Get(r.defaultBackend)
}

func (r *Registry) List() []Adapter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapters := make([]Adapter, 0, len(r.adapters))
	for _, a := range r.adapters {
		adapters = append(adapters, a)
	}
	return adapters
}

func (r *Registry) FindAdapterForModel(model string) (Adapter, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	backendID, modelID := parseModel(model)

	if backendID != "" {
		adapter, ok := r.adapters[backendID]
		if !ok {
			return nil, "", fmt.Errorf("backend '%s' not found", backendID)
		}
		if !adapter.SupportsModel(modelID) {
			return nil, "", fmt.Errorf("model '%s' not found in backend '%s'", modelID, backendID)
		}
		return adapter, adapter.ResolveModel(modelID), nil
	}

	if defaultAdapter, ok := r.adapters[r.defaultBackend]; ok {
		if defaultAdapter.SupportsModel(model) {
			return defaultAdapter, defaultAdapter.ResolveModel(model), nil
		}
	}

	for _, adapter := range r.adapters {
		if adapter.SupportsModel(model) {
			return adapter, adapter.ResolveModel(model), nil
		}
	}

	return nil, "", fmt.Errorf("model '%s' not found in any backend", model)
}

func parseModel(model string) (backend, modelID string) {
	for i, c := range model {
		if c == '/' {
			return model[:i], model[i+1:]
		}
	}
	return "", model
}

func (r *Registry) Shutdown() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, adapter := range r.adapters {
		adapter.Shutdown()
	}
}
