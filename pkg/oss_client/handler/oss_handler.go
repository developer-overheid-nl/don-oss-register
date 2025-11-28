package handler

import (
	problem "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/problem"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/util"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/services"
	"github.com/gin-gonic/gin"
)

// OSSController binds HTTP requests to the OSSController
type OSSController struct {
	Service *services.RepositoryService
}

// newOSSController creates a new controller
func NewOSSController(s *services.RepositoryService) *OSSController {
	return &OSSController{Service: s}
}

// ListRepositorys handles GET /Repositorys
func (c *OSSController) ListRepositorys(ctx *gin.Context, p *models.ListRepositorysParams) ([]models.RepositorySummary, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 10
	}
	p.BaseURL = ctx.FullPath()
	repos, pagination, err := c.Service.ListRepositorys(ctx.Request.Context(), p)
	if err != nil {
		return nil, err
	}
	util.SetPaginationHeaders(ctx.Request, ctx.Header, pagination)

	return repos, nil
}

// SearchRepositorys handles GET /Repositorys/_search
func (c *OSSController) SearchRepositorys(ctx *gin.Context, p *models.ListRepositorysSearchParams) ([]models.RepositorySummary, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 10
	}
	p.BaseURL = ctx.FullPath()
	results, pagination, err := c.Service.SearchRepositorys(ctx.Request.Context(), p)
	if err != nil {
		return nil, err
	}
	util.SetPaginationHeaders(ctx.Request, ctx.Header, pagination)
	return results, nil
}

// RetrieveRepository handles GET /Repositorys/:id
func (c *OSSController) RetrieveRepository(ctx *gin.Context, params *models.RepositoryParams) (*models.RepositoryDetail, error) {
	Repository, err := c.Service.RetrieveRepository(ctx.Request.Context(), params.Id)
	if err != nil {
		return nil, err
	}
	if Repository == nil {
		return nil, problem.NewNotFound(params.Id, "Repository not found")
	}
	return Repository, nil
}

// CreateRepository handles POST /Repositorys
func (c *OSSController) CreateRepository(ctx *gin.Context, body *models.PostRepository) (*models.RepositoryDetail, error) {
	created, err := c.Service.CreateRepository(ctx.Request.Context(), *body)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// ListOrganisations handles GET /organisations
func (c *OSSController) ListOrganisations(ctx *gin.Context, p *models.ListOrganisationsParams) ([]models.OrganisationSummary, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 10
	}
	p.BaseURL = ctx.FullPath()

	orgs, pagination, err := c.Service.ListOrganisations(ctx.Request.Context(), p)
	if err != nil {
		return nil, err
	}
	util.SetPaginationHeaders(ctx.Request, ctx.Header, pagination)

	return orgs, nil
}

// CreateOrganisation handles POST /organisations
func (c *OSSController) CreateOrganisation(ctx *gin.Context, body *models.Organisation) (*models.Organisation, error) {
	created, err := c.Service.CreateOrganisation(ctx.Request.Context(), body)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// ListGitOrganisations handles GET /GitRepositorys
func (c *OSSController) ListGitOrganisations(ctx *gin.Context, p *models.ListGitOrganisationsParams) ([]models.GitOrganisatie, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 10
	}
	p.BaseURL = ctx.FullPath()
	gitOrganisations, pagination, err := c.Service.ListGitOrganisations(ctx.Request.Context(), p)
	if err != nil {
		return nil, err
	}
	util.SetPaginationHeaders(ctx.Request, ctx.Header, pagination)

	return gitOrganisations, nil
}

// CreateGitOrganisation handles POST /GitOrganisation
func (c *OSSController) CreateGitOrganisation(ctx *gin.Context, body *models.PostGitOrganisatie) (*models.GitOrganisatie, error) {
	created, err := c.Service.CreateGitOrganisatie(ctx.Request.Context(), *body)
	if err != nil {
		return nil, err
	}
	return created, nil
}
