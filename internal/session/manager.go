package session

import (
	"context"
)

// session  cold -> warming -> warmed -> in use -> delete
type Manager interface {
	// Lifecycle management
	Init(ctx context.Context, cfg *Config) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// Session pool management
	PoolStatus(ctx context.Context) (PoolStatus, error)

	// State transition methods (State Pattern)
	AcquireCold(ctx context.Context) (*Session, error)   // Get a cold session and change cold -> warming
	SetWarmed(ctx context.Context, id string) error      // Change warming -> warmed
	AcquireWarmed(ctx context.Context) (*Session, error) // Get a warmed session and change warmed -> in_use
	Release(ctx context.Context, id string) error        // Delete session completely

	// Session utilities
	GetSession(ctx context.Context, id string) (*Session, error)
	ListSessions(ctx context.Context) ([]*Session, error)
	Heartbeat(ctx context.Context, id string) error // Prevent session from being deleted due to timeout
}
