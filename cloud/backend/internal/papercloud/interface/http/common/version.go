package unifiedhttp

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// curl http://localhost:8000/papercloud/api/v1/version
type PaperCloudVersionHTTPHandler struct {
	log *zap.Logger
}

func NewPaperCloudVersionHTTPHandler(
	log *zap.Logger,
) *PaperCloudVersionHTTPHandler {
	return &PaperCloudVersionHTTPHandler{log}
}

type PaperCloudVersionResponseIDO struct {
	Version string `json:"version"`
}

func (h *PaperCloudVersionHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	response := PaperCloudVersionResponseIDO{Version: "v1.0.0"}
	json.NewEncoder(w).Encode(response)
}

func (*PaperCloudVersionHTTPHandler) Pattern() string {
	return "/papercloud/api/v1/version"
}
