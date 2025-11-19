/* eslint-disable no-unused-vars */
const Service = require('./Service');

/**
 * Create repository
 * Registreer een nieuwe OSS repository in het register.
 *
 * postRepository PostRepository 
 * returns Repository
 */
// const createRepository = async ({ postRepository }) => {
const createRepository = async (params) => {
  try {
    const mockResult = await Service.applyMock('RepositoriesService', 'createRepository', params);
    if (mockResult !== undefined) {
      if (mockResult.action === 'reject') {
        throw mockResult.value;
      }
      return mockResult.value;
    }
    return Service.successResponse(params);
  } catch (e) {
    const status = typeof e.status === 'number' && e.status > 0 ? e.status : 400;
    const message = e && e.message ? e.message : 'Er is een fout opgetreden.';
    throw Service.rejectResponse({
      message,
      detail: e.detail || message,
    }, status);
  }
};

/**
 * Get repository by id
 * Geeft één OSS repository terug op basis van het id.
 *
 * id UUID Het id van de repository.
 * returns Repository
 */
// const getRepositoryById = async ({ id }) => {
const getRepositoryById = async (params) => {
  try {
    const mockResult = await Service.applyMock('RepositoriesService', 'getRepositoryById', params);
    if (mockResult !== undefined) {
      if (mockResult.action === 'reject') {
        throw mockResult.value;
      }
      return mockResult.value;
    }
    return Service.successResponse(params);
  } catch (e) {
    const status = typeof e.status === 'number' && e.status > 0 ? e.status : 400;
    const message = e && e.message ? e.message : 'Er is een fout opgetreden.';
    throw Service.rejectResponse({
      message,
      detail: e.detail || message,
    }, status);
  }
};

/**
 * List repositories
 * Geeft een lijst terug met OSS repositories die in het register zijn opgenomen.
 *
 * status String Filter op publiccode-status. all = alle repositories, withPublicCode = alleen repositories met een publiccode.yaml, withoutPublicCode = alleen repositories zonder publiccode.yaml. (optional)
 * page Integer Paginanummer (1-based). (optional)
 * perPage Integer Aantal resultaten per pagina. (optional)
 * organisation String Filter op organisatie (bijvoorbeeld de organisation URI). (optional)
 * ids String Kommagescheiden lijst met repository-id's (uuid). (optional)
 * returns RepositoryListResponse
 */
// const listRepositories = async ({ status, page, perPage, organisation, ids }) => {
const listRepositories = async (params) => {
  try {
    const mockResult = await Service.applyMock('RepositoriesService', 'listRepositories', params);
    if (mockResult !== undefined) {
      if (mockResult.action === 'reject') {
        throw mockResult.value;
      }
      return mockResult.value;
    }
    return Service.successResponse(params);
  } catch (e) {
    const status = typeof e.status === 'number' && e.status > 0 ? e.status : 400;
    const message = e && e.message ? e.message : 'Er is een fout opgetreden.';
    throw Service.rejectResponse({
      message,
      detail: e.detail || message,
    }, status);
  }
};

module.exports = {
  createRepository,
  getRepositoryById,
  listRepositories,
};
