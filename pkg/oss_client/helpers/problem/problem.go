package problem

import "net/http"

type ErrorDetail struct {
	In       string `json:"in"`
	Location string `json:"location"`
	Code     string `json:"code"`
	Detail   string `json:"detail"`
}

// ProblemJSON implements the OSS-register error envelope.
type ProblemJSON struct {
	Status int           `json:"status"`
	Title  string        `json:"title"`
	Errors []ErrorDetail `json:"errors,omitempty"`
}

func (e ProblemJSON) Error() string { return e.Title }

func New(status int, title string, details ...ErrorDetail) ProblemJSON {
	return ProblemJSON{
		Status: status,
		Title:  title,
		Errors: details,
	}
}

func NewBadRequest(title string, details ...ErrorDetail) ProblemJSON {
	return New(http.StatusBadRequest, title, details...)
}

func NewNotFound(title string) ProblemJSON {
	return New(http.StatusNotFound, title)
}

func NewInternalServerError(title string) ProblemJSON {
	return New(http.StatusInternalServerError, title)
}

func NewForbidden(title string) ProblemJSON {
	return New(http.StatusForbidden, title)
}
