package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	problem "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/problem"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvalidParamsFromBindingReturnsGenericDetailForNonValidationErrors(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/repositories", nil)

	details := invalidParamsFromBinding(ctx, errors.New("bad body"))

	require.Len(t, details, 1)
	assert.Equal(t, "body", details[0].In)
	assert.Equal(t, "#/", details[0].Location)
	assert.Equal(t, "invalid", details[0].Code)
	assert.Equal(t, "bad body", details[0].Detail)
}

func TestInvalidParamsFromBindingMapsValidationErrors(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/repositories", nil)
	validate := validator.New()

	err := validate.Struct(struct {
		Query string `validate:"required"`
	}{})
	require.Error(t, err)

	details := invalidParamsFromBinding(ctx, err)

	require.Len(t, details, 1)
	assert.Equal(t, problem.ErrorDetail{
		In:       "query",
		Location: "#/query",
		Code:     "required",
		Detail:   "is required",
	}, details[0])
}

func TestHumanReasonAndFieldHelpers(t *testing.T) {
	validate := validator.New()
	err := validate.Var("not-a-url", "url")
	require.Error(t, err)
	var validationErrors validator.ValidationErrors
	require.ErrorAs(t, err, &validationErrors)

	assert.Equal(t, "must be a valid URL", humanReason(validationErrors[0]))
	assert.Equal(t, "body", normalizeFieldName(""))
	assert.Equal(t, "repositoryUrl", normalizeFieldName("RepositoryUrl"))

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPut, "/v1/repositories/id", nil)
	assert.Equal(t, "path", inferLocation(ctx, "id"))
	assert.Equal(t, "body", inferLocation(ctx, "url"))
	assert.True(t, isValidationErr(err))
	assert.False(t, isValidationErr(errors.New("plain")))
}
