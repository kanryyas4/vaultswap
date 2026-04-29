package audit_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultswap/internal/audit"
)

func TestLog_Success(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf)
	l.Log(audit.OpSync, "aws-prod", "db/password", nil)

	var e audit.Entry
	if err := json.NewDecoder(&buf).Decode(&e); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if e.Operation != audit.OpSync {
		t.Errorf("op = %q, want %q", e.Operation, audit.OpSync)
	}
	if e.Provider != "aws-prod" {
		t.Errorf("provider = %q, want aws-prod", e.Provider)
	}
	if e.SecretKey != "db/password" {
		t.Errorf("key = %q, want db/password", e.SecretKey)
	}
	if !e.Success {
		t.Error("expected success=true")
	}
	if e.Error != "" {
		t.Errorf("unexpected error field: %q", e.Error)
	}
	if e.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}
}

func TestLog_Failure(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf)
	sentinel := errors.New("access denied")
	l.Log(audit.OpRotate, "vault-dev", "api/key", sentinel)

	var e audit.Entry
	if err := json.NewDecoder(&buf).Decode(&e); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if e.Success {
		t.Error("expected success=false")
	}
	if !strings.Contains(e.Error, "access denied") {
		t.Errorf("error field = %q, want 'access denied'", e.Error)
	}
}

func TestLog_MultipleEntries(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf)
	ops := []audit.Operation{audit.OpGet, audit.OpPut, audit.OpDelete}
	for _, op := range ops {
		l.Log(op, "gcp-staging", "secret/x", nil)
	}

	dec := json.NewDecoder(&buf)
	count := 0
	for dec.More() {
		var e audit.Entry
		if err := dec.Decode(&e); err != nil {
			t.Fatalf("decode entry %d: %v", count, err)
		}
		if e.Operation != ops[count] {
			t.Errorf("entry %d op = %q, want %q", count, e.Operation, ops[count])
		}
		count++
	}
	if count != 3 {
		t.Errorf("expected 3 entries, got %d", count)
	}
}

func TestLog_TimestampUTC(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf)
	before := time.Now().UTC()
	l.Log(audit.OpSync, "p", "k", nil)
	after := time.Now().UTC()

	var e audit.Entry
	_ = json.NewDecoder(&buf).Decode(&e)
	if e.Timestamp.Before(before) || e.Timestamp.After(after) {
		t.Errorf("timestamp %v not between %v and %v", e.Timestamp, before, after)
	}
}
