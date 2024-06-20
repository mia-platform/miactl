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

package printer

import (
	"io"
)

type NopPrinter struct{}

func (n *NopPrinter) SetWriter(_ io.Writer) IPrinter {
	return n
}

func (n *NopPrinter) Keys(_ ...string) IPrinter {
	return n
}

func (n *NopPrinter) Record(_ ...string) IPrinter {
	return n
}

func (n *NopPrinter) BulkRecords(_ ...[]string) IPrinter {
	return n
}

func (n *NopPrinter) Print() {
}
