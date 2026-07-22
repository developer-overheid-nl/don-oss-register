package api_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateOrganisationInputRequiresOnlyURI(t *testing.T) {
	data, err := os.ReadFile("openapi.json")
	require.NoError(t, err)

	var spec map[string]any
	require.NoError(t, json.Unmarshal(data, &spec))

	paths := spec["paths"].(map[string]any)
	organisations := paths["/organisations"].(map[string]any)
	post := organisations["post"].(map[string]any)
	requestBody := post["requestBody"].(map[string]any)
	content := requestBody["content"].(map[string]any)
	jsonContent := content["application/json"].(map[string]any)
	schema := jsonContent["schema"].(map[string]any)
	assert.Equal(t, "#/components/schemas/OrganisationInput", schema["$ref"])

	components := spec["components"].(map[string]any)
	schemas := components["schemas"].(map[string]any)
	input := schemas["OrganisationInput"].(map[string]any)
	assert.Equal(t, []any{"uri"}, input["required"])
	assert.Contains(t, input["properties"].(map[string]any), "label")
}
