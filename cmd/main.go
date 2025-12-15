package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/handler"
	problem "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/problem"
	util "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/util"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/loopfz/gadgeto/tonic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	api "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/database"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/repositories"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/services"
)

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
		internal := problem.NewInternalServerError(err.Error())
		c.Header("Content-Type", "application/problem+json")
		return internal.Status, internal
	})
}

func isValidationErr(err error) bool {
	var verrs validator.ValidationErrors
	return errors.As(err, &verrs)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file ", err)
	}

	version, err := util.LoadOASVersion("./api/openapi.json")
	if err != nil {
		log.Fatalf("failed to load OAS version: %v", err)
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
		log.Fatalf("Geen databaseverbinding: %v", err)
	}
	repo := repositories.NewRepositoriesRepository(db)
	repositoriesService := services.NewRepositoryService(repo)
	controller := handler.NewOSSController(repositoriesService)
	if _, err := repositoriesService.CreateOrganisation(context.Background(), &models.Organisation{Uri: "https://www.gpp-woo.nl", Label: "GPP-Woo"}); err != nil {
		fmt.Printf("[GPP-Woo-import] create org warning: %v\n", err)
	}
	if _, err := repositoriesService.CreateOrganisation(context.Background(), &models.Organisation{Uri: "https://www.geonovum.nl", Label: "Stichting Geonovum"}); err != nil {
		fmt.Printf("[Geonovum-import] create org warning: %v\n", err)
	}
	if _, err := repositoriesService.CreateOrganisation(context.Background(), &models.Organisation{Uri: "https://www.ictu.nl", Label: "ICTU"}); err != nil {
		fmt.Printf("[ICTU-import] create org warning: %v\n", err)
	}
	if _, err := repositoriesService.CreateOrganisation(context.Background(), &models.Organisation{Uri: "https://vng.nl", Label: "Vereniging van Nederlandse Gemeenten"}); err != nil {
		fmt.Printf("[VNG-import] create org warning: %v\n", err)
	}
	if _, err := repositoriesService.CreateOrganisation(context.Background(), &models.Organisation{Uri: "https://developer.overheid.nl/", Label: "Developer overheid"}); err != nil {
		fmt.Printf("[Developer-overheid-import] create org warning: %v\n", err)
	}

	// Start server
	router := api.NewRouter(version, controller)

	log.Println("Server is running on port 1337")
	log.Fatal(http.ListenAndServe(":1337", router))
}
