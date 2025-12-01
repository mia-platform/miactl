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
	"fmt"
	"strings"

	"github.com/mia-platform/miactl/internal/printer"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
)

func buildTable(headers []string, items []marketplace.ApplyResponseItem, columnTransform func(item marketplace.ApplyResponseItem) []string) string {
	// FIXME: should use the printer from clioptions!
	str := &strings.Builder{}
	p := printer.NewTablePrinter(printer.TablePrinterOptions{WrapLinesDisabled: true}, str)
	p.Keys(headers...)

	for _, item := range items {
		p.Record(columnTransform(item)...)
	}

	p.Print()
	return str.String()
}

func buildSuccessTable(items []marketplace.ApplyResponseItem) string {
	headers := []string{"Object ID", "Item ID", "Status"}
	columnTransform := func(item marketplace.ApplyResponseItem) []string {
		var status string
		switch {
		case item.Inserted:
			status = "Inserted"
		case item.Updated:
			status = "Updated"
		default:
			// should never happen, but just in case:
			status = "UNKNOWN"
		}
		return []string{item.ID, item.ItemID, status}
	}

	return buildTable(headers, items, columnTransform)
}

func buildFailureTable(items []marketplace.ApplyResponseItem) string {
	headers := []string{"Object ID", "Item ID", "Validation Errors"}
	columnTransform := func(item marketplace.ApplyResponseItem) []string {
		var stringBuilder strings.Builder
		var validationErrors []marketplace.ApplyResponseItemValidationError
		if len(item.Errors) > 0 {
			validationErrors = item.Errors
		} else {
			validationErrors = item.ValidationErrors
		}
		for i, valErr := range validationErrors {
			stringBuilder.WriteString(valErr.Message)
			if len(validationErrors)-1 > i {
				stringBuilder.WriteString("\n")
			}
		}
		validationErrString := stringBuilder.String()
		if validationErrString == "" {
			validationErrString = "-"
		}
		id := "N/A"
		if item.ID != "" {
			id = item.ID
		}
		return []string{id, item.ItemID, validationErrString}
	}

	return buildTable(headers, items, columnTransform)
}

func buildOutcomeSummaryAsTables(outcome *marketplace.ApplyResponse) string {
	successfulItems, failedItems := separateSuccessAndFailures(outcome.Items)
	successfulCount := len(successfulItems)
	failedCount := len(failedItems)

	outcomeStr := ""

	if successfulCount > 0 {
		outcomeStr += fmt.Sprintf("%d of %d items have been successfully applied:\n\n", successfulCount, len(outcome.Items))
		outcomeStr += buildSuccessTable(successfulItems)
	}

	if failedCount > 0 && successfulCount > 0 {
		outcomeStr += "\n"
	}

	if failedCount > 0 {
		outcomeStr += fmt.Sprintf("%d of %d items have not been applied due to validation errors:\n\n", failedCount, len(outcome.Items))
		outcomeStr += buildFailureTable(failedItems)
	}

	return outcomeStr
}

func separateSuccessAndFailures(items []marketplace.ApplyResponseItem) ([]marketplace.ApplyResponseItem, []marketplace.ApplyResponseItem) {
	var successfulItems, failedItems []marketplace.ApplyResponseItem
	for _, item := range items {
		if item.Done {
			successfulItems = append(successfulItems, item)
		} else {
			failedItems = append(failedItems, item)
		}
	}
	return successfulItems, failedItems
}
