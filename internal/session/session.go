package session

import (
	"time"

	"github.com/letusgogo/playable-backend/internal/anbox"
)

type SessionStatus string

const (
	SessCold       SessionStatus = "cold"
	SessWarming    SessionStatus = "warming"
	SessWarm       SessionStatus = "warm"
	SessInUse      SessionStatus = "in_use"
	SessReclaiming SessionStatus = "reclaiming"
)

type Session struct {
	ID            string
	Game          string
	Status        SessionStatus
	Anbox         *anbox.SessionDetails
	ExpiresAt     time.Time // InUse 的业务 TTL
	LastHeartbeat time.Time
	CreatedAt     time.Time
}
