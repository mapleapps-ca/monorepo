package http

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// curl http://localhost:8000/healthcheck
type GetHealthCheckHTTPHandler struct {
	log *zap.Logger
}

func NewGetHealthCheckHTTPHandler(
	log *zap.Logger,
) *GetHealthCheckHTTPHandler {
	return &GetHealthCheckHTTPHandler{log}
}

type HealthCheckResponseIDO struct {
	Status string `json:"status"`
}

func (h *GetHealthCheckHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	response := HealthCheckResponseIDO{Status: "running"}
	json.NewEncoder(w).Encode(response)
}

func (*GetHealthCheckHTTPHandler) Pattern() string {
	return "/healthcheck"
}
