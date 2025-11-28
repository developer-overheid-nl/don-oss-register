package oss_client

import (
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/handler"
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

	notFoundResponse = fizz.Response(
		"404",
		"Not Found",
		nil,
		nil,
		nil,
	)
)

func NewRouter(apiVersion string, controller *handler.OSSController) *fizz.Fizz {
	//gin.SetMode(gin.ReleaseMode)
	g := gin.Default()

	// Configure CORS to allow access from everywhere
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "API-Version"}
	config.ExposeHeaders = []string{"API-Version", "Link", "Total-Count", "Total-Pages", "Per-Page", "Current-Page"}
	g.Use(cors.New(config))

	g.Use(APIVersionMiddleware(apiVersion))
	f := fizz.NewFromEngine(g)

	root := f.Group("/v1", "OSS v1", "OSS Register V1 routes")

	read := root.Group("", "Publieke endpoints", "Alleen lezen endpoints")
	read.GET("/repositories/_search",
		[]fizz.OperationOption{
			fizz.ID("searchRepositories"),
			fizz.Summary("Zoek repositories"),
			fizz.Description("Zoekt geregistreerde repositories op basis van titel."),
			fizz.Security(&openapi.SecurityRequirement{
				"apiKey":            {},
				"clientCredentials": {"oss:read"},
			}),
			apiVersionHeader,
			notFoundResponse,
		},
		tonic.Handler(controller.SearchRepositorys, 200),
	)
	read.GET("/repositories",
		[]fizz.OperationOption{
			fizz.ID("listRepositories"),
			fizz.Summary("Alle repositories ophalen"),
			fizz.Description("Geeft een lijst met alle geregistreerde repositories terug."),
			fizz.Security(&openapi.SecurityRequirement{
				"apiKey":            {},
				"clientCredentials": {"oss:read"},
			}),
			apiVersionHeader,
			notFoundResponse,
		},
		tonic.Handler(controller.ListRepositorys, 200),
	)

	read.GET("/repositories/:id",
		[]fizz.OperationOption{
			fizz.ID("retrieveRepositorie"),
			fizz.Summary("Specifieke repository ophalen"),
			fizz.Description("Geeft de details van een specifieke repository terug."),
			fizz.Security(&openapi.SecurityRequirement{
				"apiKey":            {},
				"clientCredentials": {"oss:read"},
			}),
			apiVersionHeader,
			notFoundResponse,
		},
		tonic.Handler(controller.RetrieveRepository, 200),
	)

	read.GET("/gitOrganisations",
		[]fizz.OperationOption{
			fizz.ID("getGitOrganisation"),
			fizz.Summary("Haal een de git organisaties op"),
			fizz.Description("Haal de nieuwe git organisaties met een OpenAPI URL."),
			fizz.Security(&openapi.SecurityRequirement{
				"apiKey":            {},
				"clientCredentials": {"oss:write"},
			}),
			apiVersionHeader,
			notFoundResponse,
		},
		tonic.Handler(controller.ListGitOrganisations, 201),
	)

	readOrg := root.Group("", "Private endpoints", "Alleen lezen endpoints")
	readOrg.GET("/organisations",
		[]fizz.OperationOption{
			fizz.ID("listOrganisations"),
			fizz.Summary("Alle organisaties ophalen"),
			fizz.Description("Geeft een lijst van alle organisaties terug."),
			fizz.Security(&openapi.SecurityRequirement{
				"apiKey":            {},
				"clientCredentials": {"organisations:read"},
			}),
			apiVersionHeader,
			notFoundResponse,
		},
		tonic.Handler(controller.ListOrganisations, 200),
	)
	writeOrg := root.Group("", "Private endpoints", "Alleen lezen endpoints")
	writeOrg.POST("/organisations",
		[]fizz.OperationOption{
			fizz.ID("createOrganisation"),
			fizz.Summary("Voeg een nieuwe organisatie toe"),
			fizz.Description("Voeg een organisatie toe op basis van URI en label."),
			fizz.Security(&openapi.SecurityRequirement{
				"apiKey":            {},
				"clientCredentials": {"organisations:write"},
			}),
			apiVersionHeader,
			notFoundResponse,
		},
		tonic.Handler(controller.CreateOrganisation, 201),
	)

	write := root.Group("", "Private endpoints", "Bewerken van repositories")
	write.POST("/repositories",
		[]fizz.OperationOption{
			fizz.ID("createRepositorie"),
			fizz.Summary("Registreer een nieuwe repository"),
			fizz.Description("Registreer een nieuwe repository met een OpenAPI URL."),
			fizz.Security(&openapi.SecurityRequirement{
				"apiKey":            {},
				"clientCredentials": {"oss:write"},
			}),
			apiVersionHeader,
			notFoundResponse,
		},
		tonic.Handler(controller.CreateRepository, 201),
	)

	write.POST("/gitOrganisations",
		[]fizz.OperationOption{
			fizz.ID("createGitOrganisation"),
			fizz.Summary("Registreer een nieuwe git organisatie"),
			fizz.Description("Registreer een nieuwe git organisatie met een OpenAPI URL."),
			fizz.Security(&openapi.SecurityRequirement{
				"apiKey":            {},
				"clientCredentials": {"oss:write"},
			}),
			apiVersionHeader,
			notFoundResponse,
		},
		tonic.Handler(controller.CreateGitOrganisation, 201),
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
	if code >= 200 && code < 300 {
		w.Header().Set("API-Version", w.version)
	}
	w.ResponseWriter.WriteHeader(code)
}

func APIVersionMiddleware(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer = &apiVersionWriter{c.Writer, version}
		c.Next()
	}
}
