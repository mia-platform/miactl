package factory

import (
	"context"
	"errors"
	"fmt"
	"io"

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

// WithValue add factory to passed context
func WithValue(ctx context.Context, writer io.Writer) context.Context {
	return context.WithValue(ctx, FactoryContextKey{}, Factory{
		Renderer:         renderer.New(writer),
		miaClientCreator: sdk.New,
	})
}

// FromContext returns factory from context
func FromContext(ctx context.Context, opts sdk.Options) (*Factory, error) {
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
