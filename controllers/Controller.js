const fs = require('node:fs');
const path = require('node:path');
const config = require('../config');
const Service = require('../services/Service');

class Controller {
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
    };
    return statusTexts[status] || 'Unknown Error';
  }

  static sendResponse(response, payload) {
    /**
    * The default response-code is 200. We want to allow to change that. in That case,
    * payload will be an object consisting of a code and a payload. If not customized
    * send 200 and the payload as received in this method.
    */
    response.status(payload.code || 200);
    const responsePayload = payload.payload !== undefined ? payload.payload : payload;
    if (responsePayload instanceof Object) {
      response.json(responsePayload);
    } else {
      response.end(responsePayload);
    }
  }

  static sendError(response, error) {
    const status = error.code || 500;
    const reason = error.message || (error.error && error.error.message) || 'Unexpected error';
    const detail = error.detail || reason;
    let invalidParams = [];
    if (Array.isArray(error.invalidParams)) {
      invalidParams = error.invalidParams;
    } else if (error.field !== undefined || error.reason !== undefined) {
      invalidParams = [{
        name: error.field || 'body',
        reason: error.reason || detail,
      }];
    }

    const problem = {
      type: `https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Status/${status}`,
      title: Controller.getStatusText(status),
      status,
      detail,
      instance: error.instance || 'body',
    };
    if (invalidParams.length > 0) {
      problem.invalidParams = invalidParams;
    }

    response.status(status).json(problem);
  }

  /**
  * Files have been uploaded to the directory defined by config.js as upload directory
  * Files have a temporary name, that was saved as 'filename' of the file object that is
  * referenced in request.files array.
  * This method finds the file and changes it to the file name that was originally called
  * when it was uploaded. To prevent files from being overwritten, a timestamp is added between
  * the filename and its extension
  * @param request
  * @param fieldName
  * @returns {string}
  */
  static collectFile(request, fieldName) {
    let uploadedFileName = '';
    if (request.files && request.files.length > 0) {
      const fileObject = request.files.find((file) => file.fieldname === fieldName);
      if (fileObject) {
        const fileArray = fileObject.originalname.split('.');
        const extension = fileArray.pop();
        fileArray.push(`_${Date.now()}`);
        uploadedFileName = `${fileArray.join('')}.${extension}`;
        fs.renameSync(path.join(config.FILE_UPLOAD_PATH, fileObject.filename),
          path.join(config.FILE_UPLOAD_PATH, uploadedFileName));
      }
    }
    return uploadedFileName;
  }

  static getRequestBodyName(request) {
    const codeGenDefinedBodyName = request.openapi.schema['x-codegen-request-body-name'];
    if (codeGenDefinedBodyName !== undefined) {
      return codeGenDefinedBodyName;
    }
    const refObjectPath = request.openapi.schema.requestBody.content['application/json'].schema.$ref;
    if (refObjectPath !== undefined && refObjectPath.length > 0) {
      return (refObjectPath.substr(refObjectPath.lastIndexOf('/') + 1));
    }
    return 'body';
  }

  static aliasRequestBodyParam(params, bodyName, value) {
    if (!bodyName || value === undefined) {
      return params;
    }
    const result = { ...params };
    if (!Object.prototype.hasOwnProperty.call(result, bodyName)) {
      result[bodyName] = value;
    }
    if (!Object.prototype.hasOwnProperty.call(result, 'body')) {
      result.body = value;
    }
    const sanitizedName = Service.sanitizeOperationId(bodyName);
    if (sanitizedName && !Object.prototype.hasOwnProperty.call(result, sanitizedName)) {
      result[sanitizedName] = value;
    }
    const lowerCaseName = bodyName.charAt(0).toLowerCase() + bodyName.slice(1);
    if (lowerCaseName && !Object.prototype.hasOwnProperty.call(result, lowerCaseName)) {
      result[lowerCaseName] = value;
    }
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      Object.keys(value).forEach((key) => {
        if (!Object.prototype.hasOwnProperty.call(result, key)) {
          result[key] = value[key];
        }
      });
    }
    return result;
  }

  static collectRequestParams(request) {
    let requestParams = {};
    const { requestBody } = request.openapi.schema || {};
    if (requestBody) {
      const { content } = requestBody;
      if (content['application/json'] !== undefined) {
        const requestBodyName = this.getRequestBodyName(request);
        const schemaObject = content['application/json'].schema || {};
        let schemaDefinition = schemaObject;
        if (schemaObject.$ref) {
          const resolved = Service.resolveRef(schemaObject.$ref);
          if (resolved) {
            schemaDefinition = resolved;
          }
        }
        const requiredProperties = Array.isArray(schemaDefinition.required)
          ? schemaDefinition.required
          : [];
        const payload = request.body || {};
        if (requiredProperties.length > 0) {
          const missing = requiredProperties.filter((prop) => payload[prop] === undefined);
          if (missing.length > 0) {
            throw Service.rejectResponse({
              message: `Missing required properties: ${missing.join(', ')}`,
              field: requestBodyName,
              detail: `Missing required properties: ${missing.join(', ')}`,
            }, 400);
          }
        }
        const declaredProperties = schemaDefinition && schemaDefinition.properties
          ? Object.keys(schemaDefinition.properties)
          : [];
        const allowAdditional = (() => {
          if (schemaDefinition && Object.prototype.hasOwnProperty.call(schemaDefinition, 'additionalProperties')) {
            return schemaDefinition.additionalProperties !== false;
          }
          return false;
        })();
        if (!allowAdditional) {
          const unknownProperties = Object.keys(payload).filter((prop) => !declaredProperties.includes(prop));
          if (unknownProperties.length > 0) {
            throw Service.rejectResponse({
              message: `Unknown properties: ${unknownProperties.join(', ')}`,
              field: requestBodyName,
              detail: `Unknown properties: ${unknownProperties.join(', ')}`,
            }, 400);
          }
        }
        if (Object.keys(payload).length > 0) {
          requestParams = Controller.aliasRequestBodyParam({
            ...requestParams,
            [requestBodyName]: payload,
          }, requestBodyName, payload);
        }
      } else if (content['multipart/form-data'] !== undefined) {
        Object.keys(content['multipart/form-data'].schema.properties).forEach(
          (property) => {
            const propertyObject = content['multipart/form-data'].schema.properties[property];
            if (propertyObject.format !== undefined && propertyObject.format === 'binary') {
              requestParams[property] = this.collectFile(request, property);
            } else {
              requestParams[property] = request.body[property];
            }
          },
        );
      }
    }

    if (request.openapi.schema.parameters !== undefined) {
      request.openapi.schema.parameters.forEach((param) => {
        if (param.in === 'path') {
          requestParams[param.name] = request.openapi.pathParams[param.name];
        } else if (param.in === 'query') {
          requestParams[param.name] = request.query[param.name];
        } else if (param.in === 'header') {
          requestParams[param.name] = request.headers[param.name];
        }
      });
    }
    return requestParams;
  }

  static async handleRequest(request, response, serviceOperation) {
    try {
      const serviceResponse = await serviceOperation(this.collectRequestParams(request));
      Controller.sendResponse(response, serviceResponse);
    } catch (error) {
      Controller.sendError(response, error);
    }
  }
}

module.exports = Controller;
