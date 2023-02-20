package factory

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/mia-platform/miactl/old/renderer"
	"github.com/mia-platform/miactl/old/sdk"
	sdkErrors "github.com/mia-platform/miactl/old/sdk/errors"

	"github.com/stretchr/testify/require"
)

func TestWithValue(t *testing.T) {
	t.Run("save factory to passed context", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithValue(ctx, &bytes.Buffer{})
		f := ctx.Value(ContextKey{})
		require.NotNil(t, f)
		if _, ok := f.(Factory); ok {
			return
		}
		t.Fail()
	})
}

func TestAddMiaClientToFactory(t *testing.T) {
	t.Run("throws if newSdk not defined", func(t *testing.T) {
		f := Factory{}
		require.NotNil(t, f)
		opts := sdk.Options{}
		err := f.addMiaClientToFactory(opts)
		require.EqualError(t, err, fmt.Sprintf("%s: newSdk not defined", sdkErrors.ErrCreateClient))
	})

	t.Run("throws if options are not passed", func(t *testing.T) {
		f := Factory{
			miaClientCreator: sdk.New,
		}
		require.NotNil(t, f)
		opts := sdk.Options{}
		err := f.addMiaClientToFactory(opts)
		require.EqualError(t, err, fmt.Sprintf("%s: client options are not correct", sdkErrors.ErrCreateClient))
	})

	t.Run("method add MiaClient to factory", func(t *testing.T) {
		opts := sdk.Options{
			APIKey:     "my-apiKey",
			APIBaseURL: "http://base-url/",
			APICookie:  "cookie",
		}
		miaClient, err := sdk.New(opts)
		require.NoError(t, err)
		miaClientCreator := func(optsArg sdk.Options) (*sdk.MiaClient, error) {
			require.Equal(t, opts, optsArg)
			return miaClient, nil
		}
		f := Factory{
			miaClientCreator: miaClientCreator,
		}
		require.NotNil(t, f)
		err = f.addMiaClientToFactory(opts)
		require.NoError(t, err)

		require.Equal(t, nil, f.Renderer)
		require.Equal(t, miaClient, f.MiaClient)
		require.Equal(t, reflect.ValueOf(miaClientCreator).Pointer(), reflect.ValueOf(f.miaClientCreator).Pointer())
	})
}

func TestFromContext(t *testing.T) {
	t.Run("throws if context error", func(t *testing.T) {
		ctx, cancFn := context.WithTimeout(context.Background(), 0)
		defer cancFn()
		f, err := FromContext(ctx, sdk.Options{})

		require.Nil(t, f)
		require.EqualError(t, err, "context error")
	})

	t.Run("throws if mia client error", func(t *testing.T) {
		ctx := context.Background()
		buf := &bytes.Buffer{}
		ctx = WithValue(ctx, buf)
		f, err := FromContext(ctx, sdk.Options{})

		require.Nil(t, f)
		require.Error(t, err)
		require.EqualError(t, err, fmt.Sprintf("%s: client options are not correct", sdkErrors.ErrCreateClient))
	})

	t.Run("returns factory", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithValue(ctx, &bytes.Buffer{})
		opts := sdk.Options{
			APIBaseURL: "http://base-url/",
			APICookie:  "cookie",
			APIKey:     "my-APIKey",
		}

		f, err := FromContext(ctx, opts)

		require.NoError(t, err)
		require.Equal(t, renderer.New(&bytes.Buffer{}), f.Renderer)
		require.NotNil(t, f.MiaClient)
		require.Equal(t, reflect.ValueOf(sdk.New).Pointer(), reflect.ValueOf(f.miaClientCreator).Pointer())
	})
}
