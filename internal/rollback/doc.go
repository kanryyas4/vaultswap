// Package rollback provides point-in-time snapshot and restore capabilities
// for secrets stored in any registered provider.
//
// Typical usage:
//
//	rb := rollback.New(providers)
//
//	// Capture state before a destructive operation.
//	snap, err := rb.Capture(ctx, "prod")
//	if err != nil { ... }
//
//	// Perform the operation (rotate, sync, etc.).
//	// ...
//
//	// On failure, restore the previous state.
//	if err := rb.Restore(ctx, snap); err != nil { ... }
//
// Snapshots are held in memory only; they are not persisted to disk.
// For durable backups use the --backup flag on the rotate command.
package rollback
