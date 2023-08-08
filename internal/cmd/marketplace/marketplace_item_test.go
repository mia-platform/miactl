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
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// Marketplace item
	MarketplaceItemJSON = `{
		"_id": "1234567890abcdefg",
		"name": "RocketScience 101: Hello Universe Example",
		"description": "A simple Hello Universe example based on Rocket-Launcher's Interstellar Template.",
		"type": "example",
		"imageUrl": "/v2/files/download/rocket-launch-image.png",
		"supportedByImageUrl": "/v2/files/download/rocket-science-logo.png",
		"supportedBy": "NASA's Humor Department",
		"documentation": {
			"type": "markdown",
			"url": "https://raw.githubusercontent.com/rocket-launcher/Interstellar-Hello-Universe-Example/master/README.md"
		},
		"categoryId": "rocketScience",
		"resources": {
			"services": {
				"rocket-science-hello-universe-example": {
					"archiveUrl": "https://github.com/rocket-launcher/Interstellar-Hello-Universe-Example/archive/master.tar.gz",
					"containerPorts": [
						{
							"name": "spaceport",
							"from": 80,
							"to": 3000,
							"protocol": "TCP"
						}
					],
					"type": "template",
					"name": "rocket-science-hello-universe-example",
					"pipelines": {
						"space-station-ci": {
							"path": "/projects/space-station%2Fpipelines-templates/repository/files/console-pipeline%2Frocket-template.gitlab-ci.yml/raw"
						}
					}
				}
			}
		}
	}`
	MarketplaceItemYaml = `---
_id: 1234567890abcdefg
name: 'RocketScience 101: Hello Universe Example'
description: A simple Hello Universe example based on Rocket-Launcher's Interstellar
  Template.
type: example
imageUrl: "/v2/files/download/rocket-launch-image.png"
supportedByImageUrl: "/v2/files/download/rocket-science-logo.png"
supportedBy: NASA's Humor Department
documentation:
  type: markdown
  url: https://raw.githubusercontent.com/rocket-launcher/Interstellar-Hello-Universe-Example/master/README.md
categoryId: rocketScience
resources:
  services:
    rocket-science-hello-universe-example:
      archiveUrl: https://github.com/rocket-launcher/Interstellar-Hello-Universe-Example/archive/master.tar.gz
      containerPorts:
      - name: spaceport
        from: 80
        to: 3000
        protocol: TCP
      type: template
      name: rocket-science-hello-universe-example
      pipelines:
        space-station-ci:
          path: "/projects/space-station%2Fpipelines-templates/repository/files/console-pipeline%2Frocket-template.gitlab-ci.yml/raw"
`
)

func TestJSONParsing(t *testing.T) {
	marketplaceItem, err := UnmarshalMarketplaceItem([]byte(MarketplaceItemJSON))
	require.NoError(t, err)
	assert.NotEmpty(t, marketplaceItem)
	snaps.MatchSnapshot(t, marketplaceItem)
}

func TestMarketplaceItemToJSON(t *testing.T) {
	marketplaceItem, err := UnmarshalMarketplaceItem([]byte(MarketplaceItemJSON))
	require.NoError(t, err)
	assert.NotEmpty(t, marketplaceItem)
	json, err := marketplaceItem.MarshalMarketplaceItem()
	require.NoError(t, err)
	assert.NotEmpty(t, json)
	snaps.MatchJSON(t, json)
}

func TestMarketplaceItemToJSONIndent(t *testing.T) {
	marketplaceItem, err := UnmarshalMarketplaceItem([]byte(MarketplaceItemJSON))
	require.NoError(t, err)
	assert.NotEmpty(t, marketplaceItem)
	json, err := marketplaceItem.MarshalMarketplaceItemIndent()
	require.NoError(t, err)
	assert.NotEmpty(t, json)
	snaps.MatchSnapshot(t, string(json))
}

func TestYAMLParsing(t *testing.T) {
	marketplaceItem, err := UnmarshalMarketplaceItemYaml([]byte(MarketplaceItemYaml))
	require.NoError(t, err)
	assert.NotEmpty(t, marketplaceItem)
	snaps.MatchSnapshot(t, marketplaceItem)
}

func TestMarketplaceItemToYAML(t *testing.T) {
	marketplaceItem, err := UnmarshalMarketplaceItem([]byte(MarketplaceItemJSON))
	require.NoError(t, err)
	assert.NotEmpty(t, marketplaceItem)
	yaml, err := marketplaceItem.MarshalMarketplaceItemYaml()
	require.NoError(t, err)
	assert.NotEmpty(t, yaml)
	snaps.MatchSnapshot(t, string(yaml))
}
