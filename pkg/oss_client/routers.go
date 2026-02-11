package oss_client

import (
	"net/http"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/handler"
	problem "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/problem"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/loopfz/gadgeto/tonic"
	"github.com/wI2L/fizz"
	"github.com/wI2L/fizz/openapi"
)

var (
	apiVersionHeader = fizz.Header(
		"API-Version",
		"De API-versie van de response",
		"",
	)
)

func NewRouter(apiVersion string, controller *handler.OSSController) *fizz.Fizz {
	//gin.SetMode(gin.ReleaseMode)
	g := gin.Default()
	g.HandleMethodNotAllowed = true

	// Configure CORS to allow access from everywhere
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "API-Version", "X-Api-Key"}
	config.ExposeHeaders = []string{"API-Version", "Link", "Total-Count", "Total-Pages", "Per-Page", "Current-Page"}
	g.Use(cors.New(config))

	g.Use(APIVersionMiddleware(apiVersion))
	g.NoMethod(func(c *gin.Context) {
		apiErr := problem.New(http.StatusMethodNotAllowed, "Method not allowed")
		c.Header("API-Version", apiVersion)
		c.Header("Content-Type", "application/problem+json")
		c.AbortWithStatusJSON(apiErr.Status, apiErr)
	})
	g.NoRoute(func(c *gin.Context) {
		apiErr := problem.NewNotFound("Resource does not exist")
		c.Header("API-Version", apiVersion)
		c.Header("Content-Type", "application/problem+json")
		c.AbortWithStatusJSON(apiErr.Status, apiErr)
	})
	f := fizz.NewFromEngine(g)

	root := f.Group("/v1", "OSS v1", "OSS Register V1 routes")

	root.GET("/repositories/_search",
		[]fizz.OperationOption{
			fizz.ID("searchRepositories"),
			fizz.Summary("Search repositories"),
			fizz.Description("Geeft een lijst terug met OSS repositories die in het register zijn opgenomen."),
			fizz.Security(&openapi.SecurityRequirement{
				"clientCredentials": {},
			}),
			apiVersionHeader,
		},
		tonic.Handler(controller.SearchRepositorys, 200),
	)

	root.GET("/repositories",
		[]fizz.OperationOption{
			fizz.ID("listRepositories"),
			fizz.Summary("List repositories"),
			fizz.Description("Geeft een lijst terug met OSS repositories die in het register zijn opgenomen."),
			fizz.Security(&openapi.SecurityRequirement{
				"clientCredentials": {},
			}),
			apiVersionHeader,
		},
		tonic.Handler(controller.ListRepositorys, 200),
	)

	root.GET("/repositories/:id",
		[]fizz.OperationOption{
			fizz.ID("getRepositoryById"),
			fizz.Summary("Get repository by id"),
			fizz.Description("Geeft één OSS repository terug op basis van het id."),
			fizz.Security(&openapi.SecurityRequirement{
				"apiKey":            {},
				"clientCredentials": {"repositories:read"},
			}),
			apiVersionHeader,
		},
		tonic.Handler(controller.RetrieveRepository, 200),
	)

	root.PUT("/repositories/:id",
		[]fizz.OperationOption{
			fizz.ID("updateRepository"),
			fizz.Summary("Specifieke repository updaten"),
			fizz.Description("Specifieke repository updaten"),
			fizz.Security(&openapi.SecurityRequirement{
				"clientCredentials": {"repositories:write"},
			}),
			apiVersionHeader,
		},
		tonic.Handler(controller.UpdateRepository, 200),
	)

	root.POST("/repositories",
		[]fizz.OperationOption{
			fizz.ID("createRepository"),
			fizz.Summary("Create repository"),
			fizz.Description("Registreer een nieuwe OSS repository in het register."),
			fizz.Security(&openapi.SecurityRequirement{
				"clientCredentials": {"repositories:write"},
			}),
			apiVersionHeader,
		},
		tonic.Handler(controller.CreateRepository, 201),
	)

	root.GET("/git-organisations",
		[]fizz.OperationOption{
			fizz.ID("listGitOrganisations"),
			fizz.Summary("List git organisations"),
			fizz.Description("Geeft een lijst terug met git organisations die in het register zijn opgenomen."),
			fizz.Security(&openapi.SecurityRequirement{
				"clientCredentials": {},
			}),
			apiVersionHeader,
		},
		tonic.Handler(controller.ListGitOrganisations, 200),
	)

	root.POST("/git-organisations",
		[]fizz.OperationOption{
			fizz.ID("createGitOrganisation"),
			fizz.Summary("Create git organisation"),
			fizz.Description("Registreer een nieuwe git organisatie in het register."),
			fizz.Security(&openapi.SecurityRequirement{
				"clientCredentials": {"gitOrganisations:write"},
			}),
			apiVersionHeader,
		},
		tonic.Handler(controller.CreateGitOrganisation, 201),
	)

	root.GET("/organisations",
		[]fizz.OperationOption{
			fizz.ID("listOrganisations"),
			fizz.Summary("Alle organisaties ophalen"),
			fizz.Description("Alle organisaties ophalen"),
			fizz.Security(&openapi.SecurityRequirement{
				"apiKey":            {},
				"clientCredentials": {"organisations:read"},
			}),
			apiVersionHeader,
		},
		tonic.Handler(controller.ListOrganisations, 200),
	)

	root.POST("/organisations",
		[]fizz.OperationOption{
			fizz.ID("createOrganisation"),
			fizz.Summary("Organisatie aanmaken"),
			fizz.Description("Maak een nieuwe organisatie aan."),
			fizz.Security(&openapi.SecurityRequirement{
				"clientCredentials": {"organisations:write"},
			}),
			apiVersionHeader,
		},
		tonic.Handler(controller.CreateOrganisation, 201),
	)
	// 6) OpenAPI documentatie
	g.StaticFile("/v1/openapi.json", "./api/openapi.json")

	return f
}

type apiVersionWriter struct {
	gin.ResponseWriter
	version string
}

func (w *apiVersionWriter) WriteHeader(code int) {
	w.Header().Set("API-Version", w.version)
	w.ResponseWriter.WriteHeader(code)
}

func APIVersionMiddleware(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer = &apiVersionWriter{c.Writer, version}
		c.Next()
	}
}
