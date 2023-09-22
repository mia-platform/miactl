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

	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/olekukonko/tablewriter"
)

func buildTable(headers []string, items []marketplace.ApplyResponseItem, columnTransform func(item marketplace.ApplyResponseItem) []string) string {
	strBuilder := &strings.Builder{}
	table := tablewriter.NewWriter(strBuilder)
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	table.SetHeader(headers)

	for _, item := range items {
		table.Append(columnTransform(item))
	}

	table.Render()
	return strBuilder.String()
}

func buildSuccessTable(items []marketplace.ApplyResponseItem) string {
	headers := []string{"Item ID", "Name", "Status"}
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
		return []string{item.ItemID, item.Name, status}
	}

	return buildTable(headers, items, columnTransform)
}

func buildFailureTable(items []marketplace.ApplyResponseItem) string {
	headers := []string{"Item ID", "Name", "Validation Errors"}
	columnTransform := func(item marketplace.ApplyResponseItem) []string {
		var validationErrorsStr string
		validationErrors := item.ValidationErrors
		for i, valErr := range validationErrors {
			validationErrorsStr += valErr.Message
			if len(validationErrors)-1 > i {
				validationErrorsStr += "\n"
			}
		}
		if validationErrorsStr == "" {
			validationErrorsStr = "-"
		}
		return []string{item.ItemID, item.Name, validationErrorsStr}
	}

	return buildTable(headers, items, columnTransform)
}

func buildOutcomeSummaryAsTables(outcome *marketplace.ApplyResponse) string {
	successfulItems, failedItems := separateSuccessAndFailures(outcome.Items)
	successfulCount := len(successfulItems)
	failedCount := len(failedItems)

	var outcomeStr string

	if successfulCount > 0 {
		outcomeStr += fmt.Sprintf("%d of %d items have been successfully applied:\n\n", successfulCount, len(outcome.Items))
		outcomeStr += buildSuccessTable(successfulItems)
	}

	if failedCount > 0 {
		if successfulCount > 0 {
			outcomeStr += fmt.Sprintln()
		}
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
