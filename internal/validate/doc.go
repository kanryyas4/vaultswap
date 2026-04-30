// Package validate provides secret validation for vaultswap.
//
// It allows operators to define rules — required keys, non-empty
// constraints, and regex patterns — and check them against one or
// more configured providers in a single pass.
//
// Example usage:
//
//	rules := []validate.Rule{
//		{Key: "DB_PASSWORD", Required: true, NonEmpty: true},
//		{Key: "PORT",        Pattern: `^\d{4,5}$`},
//	}
//	v := validate.New(providers, rules)
//	results, err := v.Run()
package validate
