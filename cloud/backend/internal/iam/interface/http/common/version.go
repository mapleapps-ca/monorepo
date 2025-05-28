package unifiedhttp

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// curl http://localhost:8000/iam/api/v1/version
type GetMapleSendVersionHTTPHandler struct {
	log *zap.Logger
}

func NewGetMapleSendVersionHTTPHandler(
	log *zap.Logger,
) *GetMapleSendVersionHTTPHandler {
	log = log.Named("MapleSendVersionHTTPHandler")
	return &GetMapleSendVersionHTTPHandler{log}
}

type MapleSendVersionResponseIDO struct {
	Version string `json:"version"`
}

func (h *GetMapleSendVersionHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	response := MapleSendVersionResponseIDO{Version: "v1.0.0"}
	json.NewEncoder(w).Encode(response)
}

func (*GetMapleSendVersionHTTPHandler) Pattern() string {
	return "/iam/api/v1/version"
}
