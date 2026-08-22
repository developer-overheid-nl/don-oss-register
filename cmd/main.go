package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/handler"
	problem "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/problem"
	util "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/util"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	commondatabase "github.com/developer-overheid-nl/don-register-common/database"
	commonlogging "github.com/developer-overheid-nl/don-register-common/logging"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/loopfz/gadgeto/tonic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	api "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/database"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/jobs"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/repositories"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/services"
)

const appName = "oss-register"

func newApplicationLogger(output io.Writer, configuredLevel string) (*slog.Logger, error) {
	return commonlogging.NewJSONLogger(output, appName, configuredLevel)
}

func invalidParamsFromBinding(c *gin.Context, err error) []problem.ErrorDetail {
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return []problem.ErrorDetail{{
			In:       inferLocation(c, ""),
			Location: "#/",
			Code:     "invalid",
			Detail:   err.Error(),
		}}
	}

	out := make([]problem.ErrorDetail, 0, len(verrs))
	for _, fe := range verrs {
		field := normalizeFieldName(fe.Field())
		out = append(out, problem.ErrorDetail{
			In:       inferLocation(c, field),
			Location: fmt.Sprintf("#/%s", field),
			Code:     fe.Tag(),
			Detail:   humanReason(fe),
		})
	}
	return out
}

func humanReason(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "url":
		return "must be a valid URL"
	default:
		return fe.Error()
	}
}

func normalizeFieldName(name string) string {
	if name == "" {
		return "body"
	}
	return strings.ToLower(name[:1]) + name[1:]
}

func inferLocation(c *gin.Context, field string) string {
	if strings.EqualFold(field, "id") {
		return "path"
	}
	if c.Request != nil && c.Request.Method == http.MethodGet {
		return "query"
	}
	return "body"
}

func init() {
	tonic.SetErrorHook(func(c *gin.Context, err error) (int, interface{}) {
		// 1) Bind/validate errors → 400 met correcte invalidParams
		var be tonic.BindError
		if errors.As(err, &be) || isValidationErr(err) {
			invalids := invalidParamsFromBinding(c, err)
			apiErr := problem.NewBadRequest("Request validation failed", invalids...)
			c.Header("Content-Type", "application/problem+json")
			return apiErr.Status, apiErr
		}

		// 2) Jouw eigen APIError → pass-through
		if apiErr, ok := err.(problem.ProblemJSON); ok {
			c.Header("Content-Type", "application/problem+json")
			return apiErr.Status, apiErr
		}

		// 3) Alles anders → 500
		logContext := context.Background()
		if c.Request != nil {
			logContext = c.Request.Context()
		}
		slog.ErrorContext(
			logContext,
			"HTTP request failed",
			"component", "http_server",
			"operation", "handle_error",
			"error", err,
		)
		internal := problem.NewInternalServerError("Internal server error")
		c.Header("Content-Type", "application/problem+json")
		return internal.Status, internal
	})
}

func isValidationErr(err error) bool {
	var verrs validator.ValidationErrors
	return errors.As(err, &verrs)
}

func main() {
	envErr := godotenv.Load()
	logger, err := newApplicationLogger(os.Stdout, os.Getenv("LOG_LEVEL"))
	if err != nil {
		fallbackLogger, _ := newApplicationLogger(os.Stdout, "info")
		fallbackLogger.Error(
			"invalid logging configuration",
			"component", "application",
			"operation", "configure_logging",
			"error", err,
		)
		os.Exit(1)
		return
	}
	slog.SetDefault(logger)
	commondatabase.ConfigureDefaultLogging(logger)
	gin.DisableConsoleColor()
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = commonlogging.NewSlogWriter(
		logger,
		slog.LevelError,
		"http_server",
		"recovery",
	)

	if envErr != nil {
		slog.Error(
			"failed to load environment file",
			"component", "application",
			"operation", "load_environment",
			"error", envErr,
		)
		os.Exit(1)
		return
	}

	version, err := util.LoadOASVersion("./api/openapi.json")
	if err != nil {
		slog.Error(
			"failed to load OAS version",
			"component", "application",
			"operation", "load_oas_version",
			"error", err,
		)
		os.Exit(1)
		return
	}
	host := os.Getenv("DB_HOSTNAME")
	user := os.Getenv("DB_USERNAME")
	pass := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_DBNAME")
	schema := os.Getenv("DB_SCHEMA")

	u := &url.URL{
		Scheme: "postgres",
		Host:   host + ":5432",
		Path:   dbname,
	}
	u.User = url.UserPassword(user, pass)

	q := u.Query()
	// q.Set("sslmode", "require")
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()

	dbcon := u.String()
	db, err := database.Connect(dbcon)
	if err != nil {
		slog.Error(
			"database connection failed",
			"component", "database",
			"operation", "connect",
			"error", err,
		)
		os.Exit(1)
		return
	}
	commondatabase.ConfigureLogging(db, logger)
	repo := repositories.NewRepositoriesRepository(db)
	repositoriesService := services.NewRepositoryService(repo)
	controller := handler.NewOSSController(repositoriesService)
	if _, err := repositoriesService.CreateOrganisation(context.Background(), &models.Organisation{Uri: "https://www.gpp-woo.nl", Label: "GPP-Woo"}); err != nil {
		slog.Warn(
			"failed to seed organisation",
			"component", "organisation_seed",
			"operation", "create",
			"organisation_uri", "https://www.gpp-woo.nl",
			"error", err,
		)
	}
	if _, err := repositoriesService.CreateOrganisation(context.Background(), &models.Organisation{Uri: "https://www.geonovum.nl", Label: "Stichting Geonovum"}); err != nil {
		slog.Warn(
			"failed to seed organisation",
			"component", "organisation_seed",
			"operation", "create",
			"organisation_uri", "https://www.geonovum.nl",
			"error", err,
		)
	}
	if _, err := repositoriesService.CreateOrganisation(context.Background(), &models.Organisation{Uri: "https://www.ictu.nl", Label: "ICTU"}); err != nil {
		slog.Warn(
			"failed to seed organisation",
			"component", "organisation_seed",
			"operation", "create",
			"organisation_uri", "https://www.ictu.nl",
			"error", err,
		)
	}
	if _, err := repositoriesService.CreateOrganisation(context.Background(), &models.Organisation{Uri: "https://vng.nl", Label: "Vereniging van Nederlandse Gemeenten"}); err != nil {
		slog.Warn(
			"failed to seed organisation",
			"component", "organisation_seed",
			"operation", "create",
			"organisation_uri", "https://vng.nl",
			"error", err,
		)
	}
	if _, err := repositoriesService.CreateOrganisation(context.Background(), &models.Organisation{Uri: "https://developer.overheid.nl/", Label: "Developer overheid"}); err != nil {
		slog.Warn(
			"failed to seed organisation",
			"component", "organisation_seed",
			"operation", "create",
			"organisation_uri", "https://developer.overheid.nl/",
			"error", err,
		)
	}
	if err := repositoriesService.PublishAllRepositoriesToTypesense(context.Background()); err != nil {
		slog.Error(
			"initial Typesense synchronization failed",
			"component", "typesense",
			"operation", "bulk_index",
			"error", err,
		)
		os.Exit(1)
		return
	}
	jobs.NewRepositoryActiveJob(repo).Start(context.Background())

	// Start server
	router := api.NewRouter(version, controller)

	slog.Info(
		"server started",
		"component", "http_server",
		"operation", "listen",
		"address", ":1337",
	)
	if err := http.ListenAndServe(":1337", router); err != nil {
		slog.Error(
			"HTTP server stopped",
			"component", "http_server",
			"operation", "listen",
			"error", err,
		)
		os.Exit(1)
	}
}
