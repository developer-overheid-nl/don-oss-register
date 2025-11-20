package handler

import (
	"fmt"

	problem "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/problem"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/util"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/services"
	"github.com/gin-gonic/gin"
)

// OSSController binds HTTP requests to the OSSController
type OSSController struct {
	Service *services.RepositoriesService
}

// newOSSController creates a new controller
func NewOSSController(s *services.RepositoriesService) *OSSController {
	return &OSSController{Service: s}
}

// ListRepositories handles GET /repositories
func (c *OSSController) ListRepositories(ctx *gin.Context, p *models.ListRepositoriesParams) ([]models.RepositorySummary, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 10
	}
	p.BaseURL = ctx.FullPath()
	repos, pagination, err := c.Service.ListRepositories(ctx.Request.Context(), p)
	if err != nil {
		return nil, err
	}
	util.SetPaginationHeaders(ctx.Request, ctx.Header, pagination)

	return repos, nil
}

// SearchRepositories handles GET /repositories/_search
func (c *OSSController) SearchRepositories(ctx *gin.Context, p *models.ListRepositoriesSearchParams) ([]models.RepositorySummary, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 10
	}
	p.BaseURL = ctx.FullPath()
	results, pagination, err := c.Service.SearchRepositories(ctx.Request.Context(), p)
	if err != nil {
		return nil, err
	}
	util.SetPaginationHeaders(ctx.Request, ctx.Header, pagination)
	return results, nil
}

// RetrieveRepositorie handles GET /repositories/:id
func (c *OSSController) RetrieveRepositorie(ctx *gin.Context, params *models.RepositorieParams) (*models.RepositorieDetail, error) {
	repositorie, err := c.Service.RetrieveRepositorie(ctx.Request.Context(), params.Id)
	if err != nil {
		return nil, err
	}
	if repositorie == nil {
		return nil, problem.NewNotFound(params.Id, "Repositorie not found")
	}
	return repositorie, nil
}

// CreateRepositorie handles POST /repositories
func (c *OSSController) CreateRepositorie(ctx *gin.Context, body *models.PostRepositorie) (*models.Repositorie, error) {
	created, err := c.Service.CreateRepositorie(ctx.Request.Context(), *body)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// ListOrganisations handles GET /organisations
func (c *OSSController) GetAllOrganisations(ctx *gin.Context) ([]models.OrganisationSummary, error) {
	orgs, total, err := c.Service.GetAllOrganisations(ctx.Request.Context())
	if err != nil {
		return nil, err
	}
	ctx.Header("Total-Count", fmt.Sprintf("%d", total))
	orgSummaries := make([]models.OrganisationSummary, len(orgs))
	for i, org := range orgs {
		orgSummaries[i] = models.OrganisationSummary{
			Uri:   org.Uri,
			Label: org.Label,
			Links: &models.Links{
				Repositories: &models.Link{
					Href: fmt.Sprintf("/v1/repositories?organisation=%s", org.Uri),
				},
			},
		}
	}
	return orgSummaries, nil
}

// CreateOrganisation handles POST /organisations
func (c *OSSController) CreateOrganisation(ctx *gin.Context, body *models.Organisation) (*models.Organisation, error) {
	created, err := c.Service.CreateOrganisation(ctx.Request.Context(), body)
	if err != nil {
		return nil, err
	}
	return created, nil
}
