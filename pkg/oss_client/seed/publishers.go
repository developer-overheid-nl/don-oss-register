package seed

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/repositories"
)

type publishersFile struct {
	Data []publisherEntry `json:"data"`
}

type publisherEntry struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Tooi        int64             `json:"tooi"`
	CodeHosting []publisherSource `json:"codeHosting"`
	CreatedAt   string            `json:"createdAt"`
	UpdatedAt   string            `json:"updatedAt"`
}

type publisherSource struct {
	URL string `json:"url"`
}

// Publish loads the publishers file and ensures organisations and repositories exist.
func Publish(ctx context.Context, repo repositories.RepositoriesRepository, path string) error {
	fh, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open publishers file: %w", err)
	}
	defer fh.Close()

	var payload publishersFile
	if err := json.NewDecoder(fh).Decode(&payload); err != nil {
		return fmt.Errorf("decode publishers file: %w", err)
	}

	for _, entry := range payload.Data {
		if entry.Tooi == 0 || entry.Description == "" {
			continue
		}
		orgURI := fmt.Sprintf("https://identifier.overheid.nl/tooi/id/oorg/%d", entry.Tooi)
		if err := ensureOrganisation(ctx, repo, orgURI, entry.Description); err != nil {
			return err
		}

		if entry.ID == "" {
			continue
		}
		existingRepo, err := repo.GetRepositorieByID(ctx, entry.ID)
		if err != nil {
			return err
		}
		if existingRepo != nil {
			continue
		}

		url := entry.primaryURL()
		created := entry.parseTime(entry.CreatedAt)
		updated := entry.parseTime(entry.UpdatedAt)
		repoModel := &models.Repositorie{
			Id:             entry.ID,
			Name:           entry.Description,
			Description:    entry.Description,
			OrganisationID: &orgURI,
			RepositorieUri: url,
			CreatedAt:      created,
			UpdatedAt:      updated,
		}
		if err := repo.SaveRepositorie(ctx, repoModel); err != nil {
			return fmt.Errorf("save repository %s: %w", entry.ID, err)
		}
	}
	return nil
}

func ensureOrganisation(ctx context.Context, repo repositories.RepositoriesRepository, uri, label string) error {
	current, err := repo.FindOrganisationByURI(ctx, uri)
	if err != nil {
		return fmt.Errorf("find organisation %s: %w", uri, err)
	}
	if current != nil {
		return nil
	}
	org := &models.Organisation{Uri: uri, Label: label}
	if err := repo.SaveOrganisatie(org); err != nil {
		return fmt.Errorf("save organisation %s: %w", uri, err)
	}
	return nil
}

func (p publisherEntry) primaryURL() string {
	if len(p.CodeHosting) == 0 {
		return ""
	}
	return p.CodeHosting[0].URL
}

func (p publisherEntry) parseTime(value string) int64 {
	if value == "" {
		return time.Now().Unix()
	}
	t, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Now().Unix()
	}
	return t.Unix()
}
