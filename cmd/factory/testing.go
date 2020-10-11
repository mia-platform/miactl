package factory

import (
	"context"

	"github.com/mia-platform/miactl/iostreams"
	"github.com/mia-platform/miactl/renderer"
)

// WithFactoryTestValue add factory to passed context
func WithFactoryValueTest(ctx context.Context, iostreamMock *iostreams.IOStreams, miaClientMock miaClientCreator) context.Context {
	return context.WithValue(ctx, FactoryContextKey{}, Factory{
		Renderer:         renderer.New(iostreamMock.Out),
		miaClientCreator: miaClientMock,
	})
}
