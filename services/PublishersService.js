/* eslint-disable no-unused-vars */
const Service = require('./Service');

/**
 * List publishers
 * Geeft een lijst terug met publishers (organisaties) die OSS registreren.
 *
 * returns PublisherListResponse
 */
// const listPublishers = async () => {
const listPublishers = async (params) => {
  try {
    const mockResult = await Service.applyMock('PublishersService', 'listPublishers', params);
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
  listPublishers,
};
