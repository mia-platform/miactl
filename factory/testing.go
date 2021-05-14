package factory

import (
	"context"
	"io"

	"github.com/mia-platform/miactl/renderer"
)

// WithTestValue add factory to passed context
func WithValueTest(ctx context.Context, writerMock io.Writer, miaClientMock miaClientCreator) context.Context {
	return context.WithValue(ctx, FactoryContextKey{}, Factory{
		Renderer:         renderer.New(writerMock),
		miaClientCreator: miaClientMock,
	})
}
