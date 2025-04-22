// Package log is based on uber zap
package log

type options struct {
	fileLogEnabled bool
	sentry         struct {
		dsn     string
		env     string
		release string
	}
}

// Option applies option
type Option interface{ apply(*options) }
type optionFunc func(*options)

func (f optionFunc) apply(o *options) { f(o) }

// WithSentry set sentryDSN
func WithSentry(dsn, env, version string) Option {
	return optionFunc(func(o *options) {
		o.sentry.dsn = dsn
		o.sentry.env = env
		o.sentry.release = version
	})
}

// WithFileLogEnabled enable logs write in file
func WithFileLogEnabled(enable bool) Option {
	return optionFunc(func(o *options) {
		o.fileLogEnabled = enable
	})
}
