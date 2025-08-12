package game

import "context"

type Manager struct {
	games map[string]*Game
}

func NewManager(games map[string]*Game) *Manager {
	return &Manager{
		games: games,
	}
}

func (m *Manager) Get(ctx context.Context, game string) (*Game, bool) {
	g, ok := m.games[game]
	return g, ok
}
