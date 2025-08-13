package api

var (
	ErrNot = 200
)

type CommonResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type CreateSessionRequest struct {
	Game string `json:"game"`
}

type SetWarmedRequest struct {
	SessionID string `json:"session_id"`
}

type ReleaseRequest struct {
	SessionID string `json:"session_id"`
}

type DetectStageRequest struct {
	CurrentStageNum int    `json:"current_stage_num"`
	Image           string `json:"image"`
}

type DetectStageResponse struct {
	Match    bool   `json:"match"`
	StageNum int    `json:"stage_num"`
	Evidence string `json:"evidence"`
}
