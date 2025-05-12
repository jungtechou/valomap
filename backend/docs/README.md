# API Documentation with Swagger

This project uses [Swagger](https://swagger.io/) for API documentation. The documentation is automatically generated from code annotations during the build process.

## Accessing the Swagger UI

When the server is running, you can access the Swagger UI at:

```
http://localhost:3000/swagger/index.html
```

## Updating Documentation

Documentation is generated from annotations in the code. To properly document an API endpoint, add annotations before the handler function following this format:

```go
// HandlerName godoc
// @Summary Brief description of what the endpoint does
// @Description Detailed explanation of the endpoint's functionality
// @Tags category-tag
// @Accept json,xml,etc
// @Produce json,xml,etc
// @Param paramName paramType dataType isRequired description example:"example value"
// @Success statusCode {objectType} ReturnType "Description of success response"
// @Failure statusCode {objectType} ErrorType "Description of error response"
// @Router /path/to/endpoint [method]
func (h *Handler) HandlerName(c *gin.Context) {
    // Implementation
}
```

### Documentation Guidelines

1. **Every handler method** should include Swagger annotations
2. Use consistent tags to group related endpoints
3. Document all parameters, including query, path, header, and body parameters
4. Provide example values where helpful
5. Document all possible response codes and their structures
6. Include detailed descriptions for complex endpoints

## Regenerating Documentation

To manually regenerate the Swagger documentation, run:

```bash
./scripts/generate_swagger.sh
```

The documentation is also automatically generated during the Docker build process.

## Swagger Documentation Fields

Common annotation fields:

- `@Summary`: Brief summary of the endpoint
- `@Description`: Detailed description
- `@Tags`: Categorization tags (for UI grouping)
- `@Accept`: MIME types the endpoint can accept (json, xml, etc.)
- `@Produce`: MIME types the endpoint can return
- `@Param`: Parameters the endpoint accepts
  - Format: `@Param name in type required description example:"example value"`
  - `in` can be: query, path, header, body or formData
- `@Success`: Successful response details
- `@Failure`: Error response details
- `@Router`: API path and HTTP method

For more details on specific formatting, see the [Swaggo documentation](https://github.com/swaggo/swag).
