package jobs

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/repositories"
)

const (
	DefaultRepositoryActiveStaleAfterHours = 48
	DefaultRepositoryActiveStaleAfter      = DefaultRepositoryActiveStaleAfterHours * time.Hour
	EnvCrawlStaleAfterHours                = "CRAWL_STALE_AFTER_HOURS"
)

type RepositoryActiveJob struct {
	repo       repositories.RepositoriesRepository
	staleAfter time.Duration
	runAtHour  int
}

func staleAfterFromEnv() time.Duration {
	if v := os.Getenv(EnvCrawlStaleAfterHours); v != "" {
		hours, err := strconv.Atoi(v)
		if err == nil && hours > 0 {
			return time.Duration(hours) * time.Hour
		}
		log.Printf("invalid %s value %q, using default %d hours", EnvCrawlStaleAfterHours, v, DefaultRepositoryActiveStaleAfterHours)
	}
	return DefaultRepositoryActiveStaleAfter
}

func NewRepositoryActiveJob(repo repositories.RepositoriesRepository) *RepositoryActiveJob {
	return &RepositoryActiveJob{
		repo:       repo,
		staleAfter: staleAfterFromEnv(),
		runAtHour:  13,
	}
}

func (j *RepositoryActiveJob) StaleAfter() time.Duration {
	return j.staleAfter
}

func (j *RepositoryActiveJob) Start(ctx context.Context) {
	go func() {
		for {
			wait := time.Until(nextRunAt(time.Now(), j.runAtHour))
			timer := time.NewTimer(wait)
			select {
			case <-timer.C:
				j.runOnce(ctx)
			case <-ctx.Done():
				timer.Stop()
				return
			}
		}
	}()
}

func (j *RepositoryActiveJob) runOnce(ctx context.Context) {
	cutoff := time.Now().UTC().Add(-j.staleAfter)
	if err := j.refreshRepositoryActiveFlags(ctx, cutoff); err != nil {
		log.Printf("repository active job failed: %v", err)
	}
}

func nextRunAt(now time.Time, hour int) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	if now.After(next) {
		next = next.AddDate(0, 0, 1)
	}
	return next
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
