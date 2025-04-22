// Package module contains set of Uber fx modules
package module

import (
	"context"

	"go.uber.org/fx"
)

// Run the UberFX app
func Run(opts ...fx.Option) error {
	app := fx.New(opts...)

	if err := app.Err(); err != nil {
		return err
	}

	if err := app.Start(context.Background()); err != nil {
		return err
	}

	<-app.Done()

	return app.Stop(context.Background())
}
