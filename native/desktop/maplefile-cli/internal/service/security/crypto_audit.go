// internal/service/security/crypto_audit.go
package security

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// CryptoAuditEvent represents a security-relevant crypto operation
type CryptoAuditEvent struct {
	Timestamp    time.Time              `json:"timestamp"`
	Operation    string                 `json:"operation"`
	UserID       string                 `json:"user_id"`
	CollectionID string                 `json:"collection_id,omitempty"`
	FileID       string                 `json:"file_id,omitempty"`
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	DurationMs   int64                  `json:"duration_ms"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// CryptoAuditService provides security audit logging for crypto operations
type CryptoAuditService interface {
	LogCryptoOperation(ctx context.Context, event *CryptoAuditEvent)
	LogSuspiciousActivity(ctx context.Context, event *CryptoAuditEvent)
	LogKeyOperation(ctx context.Context, event *CryptoAuditEvent)
}

type cryptoAuditService struct {
	logger *zap.Logger
}

func NewCryptoAuditService(logger *zap.Logger) CryptoAuditService {
	return &cryptoAuditService{
		logger: logger.Named("CryptoAudit"),
	}
}

// LogCryptoOperation logs normal crypto operations for audit trail
func (s *cryptoAuditService) LogCryptoOperation(ctx context.Context, event *CryptoAuditEvent) {
	event.Timestamp = time.Now()

	s.logger.Info("🔐 Crypto operation audit",
		zap.String("operation", event.Operation),
		zap.String("userID", event.UserID),
		zap.String("collectionID", event.CollectionID),
		zap.String("fileID", event.FileID),
		zap.Bool("success", event.Success),
		zap.String("errorMessage", event.ErrorMessage),
		zap.Int64("durationMs", event.DurationMs),
		zap.Any("metadata", event.Metadata))
}

// LogSuspiciousActivity logs potentially suspicious crypto operations
func (s *cryptoAuditService) LogSuspiciousActivity(ctx context.Context, event *CryptoAuditEvent) {
	event.Timestamp = time.Now()

	s.logger.Warn("🚨 Suspicious crypto activity detected",
		zap.String("operation", event.Operation),
		zap.String("userID", event.UserID),
		zap.String("collectionID", event.CollectionID),
		zap.String("fileID", event.FileID),
		zap.Bool("success", event.Success),
		zap.String("errorMessage", event.ErrorMessage),
		zap.String("ipAddress", event.IPAddress),
		zap.String("userAgent", event.UserAgent),
		zap.Int64("durationMs", event.DurationMs),
		zap.Any("metadata", event.Metadata))
}

// LogKeyOperation logs sensitive key operations (generation, rotation, sharing)
func (s *cryptoAuditService) LogKeyOperation(ctx context.Context, event *CryptoAuditEvent) {
	event.Timestamp = time.Now()

	s.logger.Info("🔑 Key operation audit",
		zap.String("operation", event.Operation),
		zap.String("userID", event.UserID),
		zap.String("collectionID", event.CollectionID),
		zap.Bool("success", event.Success),
		zap.String("errorMessage", event.ErrorMessage),
		zap.Int64("durationMs", event.DurationMs),
		zap.Any("metadata", event.Metadata))
}

// Helper function to detect suspicious patterns
func (s *cryptoAuditService) IsSuspiciousActivity(operation string, userID string, failures int, timeWindow time.Duration) bool {
	// Example: Multiple failed decryption attempts in short time
	if operation == "decrypt_collection_key" && failures > 5 {
		return true
	}

	// Example: Rapid key sharing operations
	if operation == "encrypt_for_sharing" && timeWindow < time.Minute {
		return true
	}

	return false
}
