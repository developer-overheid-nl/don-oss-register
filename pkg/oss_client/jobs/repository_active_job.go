package jobs

import (
	"context"
	"log"
	"time"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/repositories"
)

const (
	DefaultRepositoryActiveRefreshInterval = 24 * time.Hour
	DefaultRepositoryActiveStaleAfter      = 48 * time.Hour
)

type RepositoryActiveJob struct {
	repo            repositories.RepositoriesRepository
	refreshInterval time.Duration
	staleAfter      time.Duration
}

func NewRepositoryActiveJob(repo repositories.RepositoriesRepository) *RepositoryActiveJob {
	return &RepositoryActiveJob{
		repo:            repo,
		refreshInterval: DefaultRepositoryActiveRefreshInterval,
		staleAfter:      DefaultRepositoryActiveStaleAfter,
	}
}

func (j *RepositoryActiveJob) Start(ctx context.Context) {
	run := func() {
		cutoff := time.Now().UTC().Add(-j.staleAfter)
		if err := j.refreshRepositoryActiveFlags(ctx, cutoff); err != nil {
			log.Printf("repository active job failed: %v", err)
		}
	}

	run()

	ticker := time.NewTicker(j.refreshInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				run()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (j *RepositoryActiveJob) refreshRepositoryActiveFlags(ctx context.Context, cutoff time.Time) error {
	repos, err := j.repo.AllRepositorys(ctx)
	if err != nil {
		return err
	}

	updated := 0
	for i := range repos {
		active := !repos[i].LastCrawledAt.Before(cutoff)
		if repos[i].Active == active {
			continue
		}
		repos[i].Active = active
		if err := j.repo.SaveRepository(ctx, &repos[i]); err != nil {
			return err
		}
		updated++
	}

	log.Printf("repository active job updated %d repositories", updated)
	return nil
}
