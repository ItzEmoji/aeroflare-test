package auth

import "aeroflare/src/secrets"

// WithSecretsManager is exported only for testing purposes so that test files
// (like those in auth_test package) can mock the secrets manager.
func (r *Resolver) WithSecretsManager(manager secrets.Manager) *Resolver {
	return r.withSecretsManager(manager)
}
