package scraper

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog"
)

// Manager handles registration and access to manga providers.
type Manager struct {
	providers map[string]Provider
	mu        sync.RWMutex
	log       zerolog.Logger
}

// NewManager creates a new scraper manager.
func NewManager(log zerolog.Logger) *Manager {
	return &Manager{
		providers: make(map[string]Provider),
		log:       log.With().Str("component", "scraper").Logger(),
	}
}

// Register adds a provider to the manager.
func (m *Manager) Register(provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()

	info := provider.Info()
	m.providers[info.ID] = provider
	m.log.Info().Str("provider", info.ID).Str("name", info.Name).Msg("registered provider")
}

// Get returns a provider by ID.
func (m *Manager) Get(id string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, ok := m.providers[id]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", id)
	}
	return provider, nil
}

// List returns info for all registered providers.
func (m *Manager) List() []ProviderInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]ProviderInfo, 0, len(m.providers))
	for _, p := range m.providers {
		infos = append(infos, p.Info())
	}
	return infos
}

// Search searches for manga using the specified provider.
func (m *Manager) Search(ctx context.Context, providerID, query string) ([]MangaResult, error) {
	provider, err := m.Get(providerID)
	if err != nil {
		return nil, err
	}
	return provider.Search(ctx, query)
}

// GetManga fetches manga details from the specified provider.
func (m *Manager) GetManga(ctx context.Context, providerID, mangaID string) (*Manga, error) {
	provider, err := m.Get(providerID)
	if err != nil {
		return nil, err
	}
	return provider.GetManga(ctx, mangaID)
}

// GetChapters fetches chapters from the specified provider.
func (m *Manager) GetChapters(ctx context.Context, providerID, mangaID string) ([]Chapter, error) {
	provider, err := m.Get(providerID)
	if err != nil {
		return nil, err
	}
	return provider.GetChapters(ctx, mangaID)
}

// GetPages fetches pages from the specified provider.
func (m *Manager) GetPages(ctx context.Context, providerID, chapterID string) ([]Page, error) {
	provider, err := m.Get(providerID)
	if err != nil {
		return nil, err
	}
	return provider.GetPages(ctx, chapterID)
}
