// Package validate checks that secrets across providers satisfy
// a set of rules (required keys, regex patterns, non-empty values).
package validate

import (
	"fmt"
	"regexp"

	"github.com/yourusername/vaultswap/internal/provider"
)

// Rule describes a single validation constraint for a secret key.
type Rule struct {
	Key      string // exact secret key to validate
	Required bool   // must exist in the provider
	Pattern  string // optional regex the value must match
	NonEmpty bool   // value must not be empty string
}

// Result holds the outcome of validating one rule against one provider.
type Result struct {
	Alias   string
	Key     string
	Passed  bool
	Message string
}

// Validator runs rules against one or more providers.
type Validator struct {
	providers map[string]provider.Provider
	rules     []Rule
}

// New creates a Validator for the given providers and rules.
func New(providers map[string]provider.Provider, rules []Rule) *Validator {
	return &Validator{providers: providers, rules: rules}
}

// Run executes all rules against all providers and returns results.
func (v *Validator) Run() ([]Result, error) {
	var results []Result

	for alias, p := range v.providers {
		for _, rule := range v.rules {
			result := Result{Alias: alias, Key: rule.Key, Passed: true}

			val, err := p.GetSecret(rule.Key)
			if err != nil {
				if rule.Required {
					result.Passed = false
					result.Message = fmt.Sprintf("required key %q not found", rule.Key)
				} else {
					result.Message = fmt.Sprintf("key %q absent (optional)", rule.Key)
				}
				results = append(results, result)
				continue
			}

			if rule.NonEmpty && val == "" {
				result.Passed = false
				result.Message = fmt.Sprintf("key %q is empty", rule.Key)
				results = append(results, result)
				continue
			}

			if rule.Pattern != "" {
				re, err := regexp.Compile(rule.Pattern)
				if err != nil {
					return nil, fmt.Errorf("invalid pattern %q: %w", rule.Pattern, err)
				}
				if !re.MatchString(val) {
					result.Passed = false
					result.Message = fmt.Sprintf("key %q value does not match pattern %q", rule.Key, rule.Pattern)
				}
			}

			results = append(results, result)
		}
	}
	return results, nil
}
