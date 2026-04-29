package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultswap/internal/audit"
)

func writeTempAuditLog(t *testing.T, entries []audit.Entry) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "audit-*.jsonl")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}
	return f.Name()
}

func TestAuditCmd_PrintsEntries(t *testing.T) {
	entries := []audit.Entry{
		{
			Timestamp: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
			Operation: audit.OpSync,
			Provider:  "aws-prod",
			SecretKey: "db/password",
			Success:   true,
		},
		{
			Timestamp: time.Date(2024, 6, 1, 12, 1, 0, 0, time.UTC),
			Operation: audit.OpRotate,
			Provider:  "vault-dev",
			SecretKey: "api/key",
			Success:   false,
			Error:     "permission denied",
		},
	}
	path := writeTempAuditLog(t, entries)

	var out bytes.Buffer
	auditCmd.SetOut(&out)
	auditCmd.SetArgs([]string{"--file", path})
	if err := auditCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	result := out.String()
	for _, want := range []string{"aws-prod", "db/password", "vault-dev", "permission denied"} {
		if !strings.Contains(result, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, result)
		}
	}
}

func TestAuditCmd_MissingFile(t *testing.T) {
	auditCmd.SetArgs([]string{"--file", "/nonexistent/audit.jsonl"})
	if err := auditCmd.Execute(); err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
