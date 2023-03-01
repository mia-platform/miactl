// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package factory

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/mia-platform/miactl/old/renderer"
	"github.com/mia-platform/miactl/old/sdk"
	sdkErrors "github.com/mia-platform/miactl/old/sdk/errors"
)

// ContextKey key of the factory in context
type ContextKey struct{}

type miaClientCreator func(opts sdk.Options) (*sdk.MiaClient, error)

// Factory returns all the clients around the commands
type Factory struct {
	Renderer  renderer.IRenderer
	MiaClient *sdk.MiaClient

	miaClientCreator miaClientCreator
}

func (o *Factory) addMiaClientToFactory(opts sdk.Options) error {
	if o.miaClientCreator == nil {
		return fmt.Errorf("%w: newSdk not defined", sdkErrors.ErrCreateClient)
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
	return context.WithValue(ctx, ContextKey{}, Factory{
		Renderer:         renderer.New(writer),
		miaClientCreator: sdk.New,
	})
}

// FromContext returns factory from context
func FromContext(ctx context.Context, opts sdk.Options) (*Factory, error) {
	factory, ok := ctx.Value(ContextKey{}).(Factory)
	if !ok {
		return nil, errors.New("context error")
	}

	err := factory.addMiaClientToFactory(opts)
	if err != nil {
		return nil, err
	}

	return &factory, nil
}
