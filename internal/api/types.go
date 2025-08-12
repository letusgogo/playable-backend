package api

var (
	ErrNot = 200
)

type CreateSessionRequest struct {
	Game string `json:"game"`
}

type CommonResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}
