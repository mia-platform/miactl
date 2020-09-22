package factory

import (
	"context"
	"errors"
	"fmt"

	"github.com/mia-platform/miactl/iostreams"
	"github.com/mia-platform/miactl/renderer"
	"github.com/mia-platform/miactl/sdk"
)

// FactoryContextKey key of the factory in context
type FactoryContextKey struct{}

type miaClientCreator func(opts sdk.Options) (*sdk.MiaClient, error)

// Factory returns all the clients around the commands
type Factory struct {
	Renderer  renderer.IRenderer
	MiaClient *sdk.MiaClient

	miaClientCreator miaClientCreator
}

func (o *Factory) addMiaClientToFactory(opts sdk.Options) error {
	if o.miaClientCreator == nil {
		return fmt.Errorf("%w: newSdk not defined", sdk.ErrCreateClient)
	}
	miaSdk, err := o.miaClientCreator(opts)
	if err != nil {
		return err
	}
	o.MiaClient = miaSdk
	return nil
}

// WithFactoryValue add factory to passed context
func WithFactoryValue(ctx context.Context, iostream *iostreams.IOStreams) context.Context {
	return context.WithValue(ctx, FactoryContextKey{}, Factory{
		Renderer:         renderer.New(iostream.Out),
		miaClientCreator: sdk.New,
	})
}

// GetFactoryFromContext returns factory from context
func GetFactoryFromContext(ctx context.Context, opts sdk.Options) (*Factory, error) {
	factory, ok := ctx.Value(FactoryContextKey{}).(Factory)
	if !ok {
		return nil, errors.New("context error")
	}

	err := factory.addMiaClientToFactory(opts)
	if err != nil {
		return nil, err
	}

	return &factory, nil
}
