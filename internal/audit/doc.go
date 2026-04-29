// Package audit implements structured, append-only audit logging for
// vaultswap operations.
//
// Each secret operation (sync, rotate, get, put, delete) is recorded as a
// newline-delimited JSON entry containing:
//
//   - timestamp  – UTC time of the operation
//   - operation  – one of sync | rotate | get | put | delete
//   - provider   – alias of the secret-manager backend involved
//   - secret_key – path / name of the secret
//   - success    – whether the operation completed without error
//   - error      – human-readable error message (omitted on success)
//
// # Usage
//
//	l := audit.New(os.Stdout)
//	l.Log(audit.OpSync, "aws-prod", "db/password", err)
//
// For persistent logs, use NewFile which opens the target path in
// append-only mode and returns the underlying *os.File so the caller
// can close it when done:
//
//	l, f, err := audit.NewFile("/var/log/vaultswap/audit.jsonl")
//	if err != nil { ... }
//	defer f.Close()
package audit
