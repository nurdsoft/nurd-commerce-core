package client

type options struct {
	external bool
}

// Option applies option
type Option interface{ apply(*options) }
type optionFunc func(*options)

func (f optionFunc) apply(o *options) { f(o) }

// WithExternalCall set for external third party calls
func WithExternalCall(external bool) Option {
	return optionFunc(func(o *options) {
		o.external = external
	})
}
