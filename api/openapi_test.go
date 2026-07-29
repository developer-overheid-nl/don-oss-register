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

func TestListRepositoriesDocumentsSorting(t *testing.T) {
	data, err := os.ReadFile("openapi.json")
	require.NoError(t, err)

	var spec map[string]any
	require.NoError(t, json.Unmarshal(data, &spec))

	paths := spec["paths"].(map[string]any)
	listRepositories := paths["/repositories"].(map[string]any)["get"].(map[string]any)
	parameters := listRepositories["parameters"].([]any)
	refs := make([]string, 0, len(parameters))
	for _, parameter := range parameters {
		refs = append(refs, parameter.(map[string]any)["$ref"].(string))
	}
	require.Contains(t, refs, "#/components/parameters/SortBy")
	require.Contains(t, refs, "#/components/parameters/SortOrder")

	components := spec["components"].(map[string]any)
	parameterDefinitions := components["parameters"].(map[string]any)

	sortBy := parameterDefinitions["SortBy"].(map[string]any)
	assert.Equal(t, "sortBy", sortBy["name"])
	assert.Equal(t, "query", sortBy["in"])
	sortBySchema := sortBy["schema"].(map[string]any)
	assert.Equal(t, []any{"title", "lastActivity"}, sortBySchema["enum"])
	assert.Equal(t, "title", sortBySchema["default"])

	sortOrder := parameterDefinitions["SortOrder"].(map[string]any)
	assert.Equal(t, "sortOrder", sortOrder["name"])
	assert.Equal(t, "query", sortOrder["in"])
	sortOrderSchema := sortOrder["schema"].(map[string]any)
	assert.Equal(t, []any{"asc", "desc"}, sortOrderSchema["enum"])
	assert.Equal(t, "asc", sortOrderSchema["default"])
}
