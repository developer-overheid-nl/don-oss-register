package problem

import (
	"net/http"

	commonproblem "github.com/developer-overheid-nl/don-register-common/problem"
)

type ErrorDetail = commonproblem.ErrorDetail

// ProblemJSON implements the OSS-register error envelope.
type ProblemJSON = commonproblem.Problem

func New(status int, title string, details ...ErrorDetail) ProblemJSON {
	return commonproblem.New(status, title, details...)
}

func NewBadRequest(title string, details ...ErrorDetail) ProblemJSON {
	return New(http.StatusBadRequest, title, details...)
}

func NewNotFound(title string) ProblemJSON {
	return commonproblem.NewNotFound(title)
}

func NewMethodNotAllowed(title string) ProblemJSON {
	return commonproblem.NewMethodNotAllowed(title)
}

func NewInternalServerError(title string) ProblemJSON {
	return commonproblem.NewInternalServerError(title)
}

func NewForbidden(title string) ProblemJSON {
	return commonproblem.NewForbidden(title)
}
