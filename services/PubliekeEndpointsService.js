/* eslint-disable no-unused-vars */
const Service = require('./Service');

/**
 * Organisatie aanmaken
 * Maak een nieuwe organisatie aan.
 *
 * organisationSummary OrganisationSummary 
 * returns OrganisationSummary
 */
// const createOrganisation = async ({ organisationSummary }) => {
const createOrganisation = async (params) => {
  try {
    const mockResult = await Service.applyMock('PubliekeEndpointsService', 'createOrganisation', params);
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
 * Alle organisaties ophalen
 * Alle organisaties ophalen
 *
 * returns listOrganisations_200_response
 */
// const listOrganisations = async () => {
const listOrganisations = async (params) => {
  try {
    const mockResult = await Service.applyMock('PubliekeEndpointsService', 'listOrganisations', params);
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
  createOrganisation,
  listOrganisations,
};
