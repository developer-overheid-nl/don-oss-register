package problem_test

import (
	"encoding/json"
	"net/http"
	"testing"

	problem "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/problem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProblemConstructors(t *testing.T) {
	detail := problem.ErrorDetail{
		In:       "body",
		Location: "#/name",
		Code:     "required",
		Detail:   "is required",
	}

	tests := []struct {
		name    string
		got     problem.ProblemJSON
		status  int
		title   string
		detail  string
		details []problem.ErrorDetail
	}{
		{
			name:    "generic",
			got:     problem.New(http.StatusTeapot, "short and stout", detail),
			status:  http.StatusTeapot,
			title:   "short and stout",
			detail:  "name: is required",
			details: []problem.ErrorDetail{detail},
		},
		{
			name:    "bad request",
			got:     problem.NewBadRequest("invalid input", detail),
			status:  http.StatusBadRequest,
			title:   "invalid input",
			detail:  "name: is required",
			details: []problem.ErrorDetail{detail},
		},
		{
			name:   "not found",
			got:    problem.NewNotFound("missing"),
			status: http.StatusNotFound,
			title:  "missing",
			detail: "The requested resource does not exist.",
		},
		{
			name:   "internal server error",
			got:    problem.NewInternalServerError("failed"),
			status: http.StatusInternalServerError,
			title:  "failed",
			detail: "An unexpected error occurred while processing the request.",
		},
		{
			name:   "forbidden",
			got:    problem.NewForbidden("denied"),
			status: http.StatusForbidden,
			title:  "denied",
			detail: "Access to the requested resource is forbidden.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.status, tt.got.Status)
			assert.Equal(t, tt.title, tt.got.Title)
			assert.Equal(t, tt.detail, tt.got.Detail)
			assert.Equal(t, tt.details, tt.got.Errors)
		})
	}
}

func TestProblemDetailCombinesMultipleErrors(t *testing.T) {
	got := problem.NewBadRequest("Request validation failed",
		problem.ErrorDetail{In: "body", Location: "#/url", Code: "url", Detail: "must be a valid URL"},
		problem.ErrorDetail{In: "query", Location: "sortBy", Code: "sortBy", Detail: "unsupported value"},
		problem.ErrorDetail{In: "body", Location: "#/", Code: "invalid", Detail: "unexpected EOF"},
	)

	assert.Equal(t, "url: must be a valid URL; sortBy: unsupported value; unexpected EOF", got.Detail)
}

func TestProblemJSONSerialisesRequiredMembers(t *testing.T) {
	data, err := json.Marshal(problem.NewNotFound("Resource does not exist"))
	require.NoError(t, err)

	var body map[string]any
	require.NoError(t, json.Unmarshal(data, &body))
	assert.Equal(t, float64(http.StatusNotFound), body["status"])
	assert.Equal(t, "Resource does not exist", body["title"])
	assert.Equal(t, "The requested resource does not exist.", body["detail"])
	assert.NotContains(t, body, "errors")
}
