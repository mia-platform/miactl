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

package marketplace

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
)

var (
	ErrServerDeleteItem     = errors.New("server error while deleting item")
	ErrUnexpectedDeleteItem = errors.New("unexpected response while deleting item")
)

func CheckDeleteResponseErrors(resp *client.Response) error {
	switch resp.StatusCode() {
	case http.StatusNoContent:
		fmt.Println("item deleted successfully")
		return nil
	case http.StatusNotFound:
		return marketplace.ErrItemNotFound
	default:
		if resp.StatusCode() >= http.StatusInternalServerError {
			return ErrServerDeleteItem
		}
		return fmt.Errorf("%w: %d", ErrUnexpectedDeleteItem, resp.StatusCode())
	}
}
