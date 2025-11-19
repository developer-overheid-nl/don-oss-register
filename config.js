const path = require('node:path');

const parseEnvBoolean = (value) => {
  if (value === undefined || value === null) {
    return false;
  }
  return ['1', 'true', 'yes', 'on'].includes(String(value).toLowerCase());
};

const config = {
  ROOT_DIR: __dirname,
  URL_PORT: 8080,
  URL_PATH: 'https://api.overheid.nl',
  BASE_VERSION: '/oss-register/v1',
  CONTROLLER_DIRECTORY: path.join(__dirname, 'controllers'),
  PROJECT_DIR: __dirname,
  USE_MOCKS: parseEnvBoolean(process.env.USE_MOCKS) || parseEnvBoolean(process.env.MOCKS_ENABLED),
};
config.OPENAPI_YAML = path.join(config.ROOT_DIR, 'api', 'openapi.yaml');
config.FULL_PATH = `${config.URL_PATH}:${config.URL_PORT}/${config.BASE_VERSION}`;
config.FILE_UPLOAD_PATH = path.join(config.PROJECT_DIR, 'uploaded_files');
config.MOCK_DIR = path.join(config.PROJECT_DIR, 'mocks');

module.exports = config;
