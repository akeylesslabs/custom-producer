package producer

// Option is a single configuration parameter used by this producer.
type Option func(*producer)

// WithDryRunEmail configures this webhook to use the provided email during
// dry-run requests to Let's Encrypt service. Regular, "production" calls use
// the email of an end user that initiated the operation (called
// `get-dynamic-secret-value` command).
func WithDryRunEmail(email string) Option {
	return func(p *producer) {
		p.dryRunEmail = email
	}
}

// WithDryRunDomain configures this webhook to use the provided domain during
// dry-run requests to Let's Encrypt service. Regular, "production" calls use
// the domain provided alongside `get-dynamic-secret-value` operation.
func WithDryRunDomain(domain string) Option {
	return func(p *producer) {
		p.dryRunDomain = domain
	}
}
