package problem_test

import (
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
		details []problem.ErrorDetail
	}{
		{
			name:    "generic",
			got:     problem.New(http.StatusTeapot, "short and stout", detail),
			status:  http.StatusTeapot,
			title:   "short and stout",
			details: []problem.ErrorDetail{detail},
		},
		{
			name:    "bad request",
			got:     problem.NewBadRequest("invalid input", detail),
			status:  http.StatusBadRequest,
			title:   "invalid input",
			details: []problem.ErrorDetail{detail},
		},
		{
			name:   "not found",
			got:    problem.NewNotFound("missing"),
			status: http.StatusNotFound,
			title:  "missing",
		},
		{
			name:   "internal server error",
			got:    problem.NewInternalServerError("failed"),
			status: http.StatusInternalServerError,
			title:  "failed",
		},
		{
			name:   "forbidden",
			got:    problem.NewForbidden("denied"),
			status: http.StatusForbidden,
			title:  "denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.status, tt.got.Status)
			assert.Equal(t, tt.title, tt.got.Title)
			assert.Equal(t, tt.details, tt.got.Errors)
		})
	}
}
