package models

import "fmt"

type RepositorySortField string

type RepositorySortOrder string

const (
	RepositorySortTitle        RepositorySortField = "title"
	RepositorySortLastActivity RepositorySortField = "lastActivity"

	RepositorySortAscending  RepositorySortOrder = "asc"
	RepositorySortDescending RepositorySortOrder = "desc"
)

type RepositorySort struct {
	Field RepositorySortField
	Order RepositorySortOrder
}

type InvalidRepositorySortError struct {
	Parameter string
	Value     string
}

func (e InvalidRepositorySortError) Error() string {
	switch e.Parameter {
	case "sortBy":
		return fmt.Sprintf("sortBy %q is ongeldig; toegestane waarden zijn title en lastActivity", e.Value)
	case "sortOrder":
		return fmt.Sprintf("sortOrder %q is ongeldig; toegestane waarden zijn asc en desc", e.Value)
	default:
		return fmt.Sprintf("ongeldige sorteerparameter %q", e.Value)
	}
}

func ParseRepositorySort(sortBy, sortOrder *string) (RepositorySort, error) {
	field := RepositorySortTitle
	if sortBy != nil {
		field = RepositorySortField(*sortBy)
	}
	order := RepositorySortAscending
	if sortOrder != nil {
		order = RepositorySortOrder(*sortOrder)
	}

	switch field {
	case RepositorySortTitle, RepositorySortLastActivity:
	default:
		return RepositorySort{}, InvalidRepositorySortError{Parameter: "sortBy", Value: *sortBy}
	}

	switch order {
	case RepositorySortAscending, RepositorySortDescending:
	default:
		return RepositorySort{}, InvalidRepositorySortError{Parameter: "sortOrder", Value: *sortOrder}
	}

	return RepositorySort{Field: field, Order: order}, nil
}
