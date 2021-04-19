package webhook

// Option is a single configuration parameter used by this webhook.
type Option func(*hook)

// WithAllowedAccessID configures this webhook to accept requests only made by
// a producer with the provided access ID.
func WithAllowedAccessID(accessID string) Option {
	return func(h *hook) {
		h.accessID = accessID
	}
}

// WithAllowedItemName configures this webhook to accept requests only made by
// a producer with the provided name.
func WithAllowedItemName(name string) Option {
	return func(h *hook) {
		h.itemName = name
	}
}
