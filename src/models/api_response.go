package models

type ApiResponse struct {
	Success    bool   `json:"success"`
	Output     string `json:"output"`
	Message    string `json:"message"`
	DurationMs int64  `json:"duration_ms"`
}
