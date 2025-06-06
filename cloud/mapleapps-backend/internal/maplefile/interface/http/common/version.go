package unifiedhttp

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// curl http://localhost:8000/maplefile/api/v1/version
type MapleFileVersionHTTPHandler struct {
	log *zap.Logger
}

func NewMapleFileVersionHTTPHandler(
	log *zap.Logger,
) *MapleFileVersionHTTPHandler {
	log = log.Named("MapleFileVersionHTTPHandler")
	return &MapleFileVersionHTTPHandler{log}
}

type MapleFileVersionResponseIDO struct {
	Version string `json:"version"`
}

func (h *MapleFileVersionHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	response := MapleFileVersionResponseIDO{Version: "v1.0.0"}
	json.NewEncoder(w).Encode(response)
}

func (*MapleFileVersionHTTPHandler) Pattern() string {
	return "/maplefile/api/v1/version"
}
