package search

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEntityJSON(t *testing.T) {
	type SDN struct {
		EntityID string `json:"entityID"`
	}
	bs, err := json.MarshalIndent(Entity[SDN]{
		SourceData: SDN{
			EntityID: "12345",
		},
	}, "", "  ")
	require.NoError(t, err)

	expected := strings.TrimSpace(`{
  "name": "",
  "entityType": "",
  "sourceList": "",
  "sourceID": "",
  "person": null,
  "business": null,
  "organization": null,
  "aircraft": null,
  "vessel": null,
  "cryptoAddresses": null,
  "addresses": null,
  "affiliations": null,
  "sanctionsInfo": null,
  "historicalInfo": null,
  "titles": null,
  "sourceData": {
    "entityID": "12345"
  }
}`)
	require.Equal(t, expected, string(bs))
}
