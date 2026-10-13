package problem

import (
	"net/http"
	"strings"

	commonproblem "github.com/developer-overheid-nl/don-register-common/problem"
)

type ErrorDetail = commonproblem.ErrorDetail

// ProblemJSON implements the OSS-register error envelope (RFC 9457) with the
// members required by the NL GOV API Design Rules: status, title and detail.
type ProblemJSON struct {
	Status int           `json:"status"`
	Title  string        `json:"title"`
	Detail string        `json:"detail"`
	Errors []ErrorDetail `json:"errors,omitempty"`
}

func (e ProblemJSON) Error() string { return e.Title }

func New(status int, title string, details ...ErrorDetail) ProblemJSON {
	var errs []ErrorDetail
	if len(details) > 0 {
		errs = details
	}
	return ProblemJSON{
		Status: status,
		Title:  title,
		Detail: buildDetail(status, title, details),
		Errors: errs,
	}
}

func NewBadRequest(title string, details ...ErrorDetail) ProblemJSON {
	return New(http.StatusBadRequest, title, details...)
}

func NewNotFound(title string) ProblemJSON {
	return New(http.StatusNotFound, title)
}

func NewMethodNotAllowed(title string) ProblemJSON {
	return New(http.StatusMethodNotAllowed, title)
}

func NewInternalServerError(title string) ProblemJSON {
	return New(http.StatusInternalServerError, title)
}

func NewForbidden(title string) ProblemJSON {
	return New(http.StatusForbidden, title)
}

// buildDetail derives a human-readable explanation specific to this occurrence
// of the problem. When error details are present they are summarised; otherwise
// a default explanation for the status code is used.
func buildDetail(status int, title string, details []ErrorDetail) string {
	if len(details) > 0 {
		parts := make([]string, 0, len(details))
		for _, d := range details {
			msg := strings.TrimSpace(d.Detail)
			if msg == "" {
				msg = strings.TrimSpace(d.Code)
			}
			field := strings.TrimPrefix(strings.TrimSpace(d.Location), "#/")
			switch {
			case field != "" && msg != "":
				parts = append(parts, field+": "+msg)
			case msg != "":
				parts = append(parts, msg)
			case field != "":
				parts = append(parts, field+": invalid")
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, "; ")
		}
	}

	switch status {
	case http.StatusBadRequest:
		return "The request is invalid."
	case http.StatusForbidden:
		return "Access to the requested resource is forbidden."
	case http.StatusNotFound:
		return "The requested resource does not exist."
	case http.StatusMethodNotAllowed:
		return "The request method is not supported for the requested resource."
	case http.StatusConflict:
		return "The request conflicts with the current state of the resource."
	case http.StatusInternalServerError:
		return "An unexpected error occurred while processing the request."
	}
	if title != "" {
		return title
	}
	return http.StatusText(status)
}
