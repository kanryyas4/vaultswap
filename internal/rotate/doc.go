// Package rotate provides secret rotation logic for vaultswap.
//
// A Rotator is initialised with a map of named providers (aliases) and
// exposes a single Rotate method that:
//
//  1. Reads the current value of the target secret.
//  2. Optionally writes the old value to a backup key in the same provider.
//  3. Writes the new value to the target key.
//
// Example usage:
//
//	r := rotate.New(providers)
//	res := r.Rotate(ctx, rotate.Options{
//		Alias:     "vault-prod",
//		SecretKey: "db/password",
//		NewValue:  newPassword,
//		BackupKey: "db/password.prev",
//	})
//	if !res.Success {
//		log.Fatal(res.Err)
//	}
package rotate
