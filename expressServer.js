// const { Middleware } = require('swagger-express-middleware');
const http = require('node:http');
const fs = require('node:fs');
const path = require('node:path');
const jsYaml = require('js-yaml');
const express = require('express');
const cors = require('cors');
const cookieParser = require('cookie-parser');
const bodyParser = require('body-parser');
const OpenApiValidator = require('express-openapi-validator');
const logger = require('./logger');
const config = require('./config');

class ExpressServer {
  static sanitizeOperationId(operationId) {
    if (!operationId || typeof operationId !== 'string') {
      return null;
    }
    let result = operationId.trim();
    if (result.length === 0) {
      return null;
    }
    result = result.replace(/[_-]+/g, ' ');
    result = result.replace(/[^a-zA-Z0-9_$]+/g, ' ');
    result = result.split(' ').filter((segment) => segment.length > 0)
      .map((segment, index) => {
        if (index === 0) {
          return segment;
        }
        return segment.charAt(0).toUpperCase() + segment.slice(1);
      }).join('');
    result = result.replace(/^[^a-zA-Z_$]+/, '');
    if (result.length === 0) {
      return null;
    }
    return result.charAt(0).toLowerCase() + result.slice(1);
  }

  static sanitizeTagName(tagName) {
    if (!tagName || typeof tagName !== 'string') {
      return null;
    }
    let result = tagName.trim();
    if (result.length === 0) {
      return null;
    }
    result = result.replace(/[_-]+/g, ' ');
    result = result.replace(/[^a-zA-Z0-9_$]+/g, ' ');
    const parts = result.split(' ').filter((segment) => segment.length > 0)
      .map((segment) => segment.charAt(0).toUpperCase() + segment.slice(1));
    if (parts.length === 0) {
      return null;
    }
    return parts.join('');
  }

  static normalizeOperationIds(schema) {
    if (!schema || !schema.paths) {
      return;
    }
    const methods = [
      'get',
      'put',
      'post',
      'delete',
      'options',
      'head',
      'patch',
      'trace',
    ];
    for (const pathKey of Object.keys(schema.paths)) {
      const pathItem = schema.paths[pathKey];
      for (const method of methods) {
        const operation = pathItem[method];
        if (operation && operation.operationId) {
          const normalizedId = ExpressServer.sanitizeOperationId(operation.operationId);
          if (normalizedId && normalizedId !== operation.operationId) {
            operation['x-original-operationId'] = operation.operationId;
            operation.operationId = normalizedId;
          }
        }
      }
    }
  }

  constructor(port, openApiYaml) {
    this.port = port;
    this.app = express();
    try {
      this.schema = jsYaml.safeLoad(fs.readFileSync(openApiYaml));
      if (this.schema && this.schema?.components) {
        const { components } = this.schema;
        const componentMirrors = [
          'schemas',
          'responses',
          'parameters',
          'examples',
          'requestBodies',
          'headers',
          'securitySchemes',
          'links',
          'callbacks',
          'pathItems',
        ];
        for (const key of componentMirrors) {
          if (!this.schema[key] && components[key]) {
            this.schema[key] = components[key];
          }
        }
      }
      ExpressServer.normalizeOperationIds(this.schema);
    } catch (e) {
      logger.error('failed to start Express Server', e.message);
    }
    this.setupMiddleware();
  }

  static getStatusText(status) {
    const statusTexts = {
      400: 'Bad Request',
      401: 'Unauthorized',
      403: 'Forbidden',
      404: 'Not Found',
      405: 'Method Not Allowed',
      409: 'Conflict',
      422: 'Unprocessable Entity',
      429: 'Too Many Requests',
      500: 'Internal Server Error',
      501: 'Not Implemented',
      502: 'Bad Gateway',
      503: 'Service Unavailable',
      504: 'Gateway Timeout',
    };
    return statusTexts[status] || 'Unknown Error';
  }

  setupMiddleware() {
    // this.setupAllowedMedia();
    this.app.use(cors());
    this.app.use(bodyParser.json({ limit: '14MB' }));
    this.app.use(express.json());
    this.app.use(express.urlencoded({ extended: false }));
    this.app.use(cookieParser());
    this.app.use((req, res, next) => {
      res.set('API-Version', this.schema.info.version);
      next();
    });
    this.app.get('/openapi.json', (req, res) => res.json(this.schema));
    this.app.get('/login-redirect', (req, res) => {
      res.status(200);
      res.json(req.query);
    });
    this.app.get('/oauth2-redirect.html', (req, res) => {
      res.status(200);
      res.json(req.query);
    });
    this.app.use(
      OpenApiValidator.middleware({
        apiSpec: this.schema,
        validateApiSpec: false,
        validateSecurity: false,
        validateRequests: {
          coerceTypes: false,
          allowUnknownBodyProperties: false,
        },
        operationHandlers: {
          basePath: path.join(__dirname),
        },
        fileUploader: { dest: config.FILE_UPLOAD_PATH },
      }),
    );
    ExpressServer.registerRoutes(this.app, this.schema);
  }

  static toExpressPath(openApiPath) {
    return openApiPath.replace(/{/g, ':').replace(/}/g, '');
  }

  static resolveOperationHandler(operation) {
    let handlerRef = operation['x-eov-operation-handler'];
    if (!handlerRef && operation.tags?.length > 0) {
      const tagName = ExpressServer.sanitizeTagName(operation.tags[0]);
      if (tagName) {
        handlerRef = `controllers/${tagName}Controller`;
      }
    }
    if (!handlerRef) {
      logger.warn(`No handler reference found for operation ${operation.operationId}`);
      return null;
    }
    const normalizedRef = handlerRef.replace(/\\/g, '/');
    const modulePath = normalizedRef.endsWith('.js')
      ? path.join(__dirname, normalizedRef)
      : path.join(__dirname, `${normalizedRef}.js`);
    if (!fs.existsSync(modulePath)) {
      logger.warn(`Handler module missing for ${operation.operationId}: ${modulePath}`);
      return null;
    }
    // eslint-disable-next-line global-require, import/no-dynamic-require
    const controllerModule = require(modulePath);
    const handler = controllerModule[operation.operationId];
    if (typeof handler !== 'function') {
      logger.warn(`Handler function ${operation.operationId} not found in ${modulePath}`);
      return null;
    }
    return handler;
  }

  static registerRoutes(app, schema) {
    if (!schema || !schema.paths) {
      return;
    }
    const methods = [
      'get',
      'put',
      'post',
      'delete',
      'options',
      'head',
      'patch',
      'trace',
    ];
    for (const pathKey of Object.keys(schema.paths)) {
      const expressPath = ExpressServer.toExpressPath(pathKey);
      const pathItem = schema.paths[pathKey];
      for (const method of methods) {
        const operation = pathItem[method];
        if (!operation) {
          continue;
        }
        const handler = ExpressServer.resolveOperationHandler(operation);
        if (!handler) {
          continue;
        }
        app[method](expressPath, async (req, res, next) => {
          logger.info(`Incoming request for ${method.toUpperCase()} ${expressPath}`);
          req.openapi = req.openapi || {};
          req.openapi.schema = req.openapi.schema || operation;
          req.openapi.pathParams = req.openapi.pathParams || req.params;
          try {
            await Promise.resolve(handler(req, res, next));
          } catch (error) {
            next(error);
          }
        });
      }
    }
  }

  launch() {
    // eslint-disable-next-line no-unused-vars
    this.app.use((err, req, res, next) => {
      // format errors using RFC 7807 Problem Details format
      const status = err.status || 500;
      const problemDetails = {
        type: `https://httpstatuses.com/${status}`,
        title: ExpressServer.getStatusText(status),
        status,
        detail: err.message || err.toString(),
      };

      // Add instance URI if available
      if (req.originalUrl) {
        problemDetails.instance = req.originalUrl;
      }

      // Add additional error details if available
      if (err.status === 400 && err.errors && Array.isArray(err.errors) && err.errors.length > 0) {
        problemDetails.invalidParams = err.errors;
      }

      // Set the proper content type for problem+json
      res.set('Content-Type', 'application/problem+json');
      res.status(status).json(problemDetails);
    });

    http.createServer(this.app).listen(this.port);
    console.log(`Listening on port ${this.port}`);
  }

  async close() {
    if (this.server !== undefined) {
      await this.server.close();
      console.log(`Server on port ${this.port} shut down`);
    }
  }
}

module.exports = ExpressServer;
