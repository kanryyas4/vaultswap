// Package audit provides structured logging of secret operations
// performed by vaultswap (sync, rotate, delete, etc.).
package audit

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

// Operation represents the type of secret operation performed.
type Operation string

const (
	OpSync   Operation = "sync"
	OpRotate Operation = "rotate"
	OpGet    Operation = "get"
	OpPut    Operation = "put"
	OpDelete Operation = "delete"
)

// Entry is a single audit log record.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Operation Operation `json:"operation"`
	Provider  string    `json:"provider"`
	SecretKey string    `json:"secret_key"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

// Logger writes audit entries to an io.Writer as newline-delimited JSON.
type Logger struct {
	w       io.Writer
	encoder *json.Encoder
}

// New creates a Logger writing to w. Pass os.Stdout for console output
// or an *os.File for a persistent audit trail.
func New(w io.Writer) *Logger {
	enc := json.NewEncoder(w)
	return &Logger{w: w, encoder: enc}
}

// NewFile opens (or creates) the file at path for append-only audit logging.
func NewFile(path string) (*Logger, *os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, nil, err
	}
	return New(f), f, nil
}

// Log records an audit entry. err may be nil on success.
func (l *Logger) Log(op Operation, provider, key string, err error) {
	e := Entry{
		Timestamp: time.Now().UTC(),
		Operation: op,
		Provider:  provider,
		SecretKey: key,
		Success:   err == nil,
	}
	if err != nil {
		e.Error = err.Error()
	}
	// Best-effort write; audit failures must not block the caller.
	_ = l.encoder.Encode(e)
}
