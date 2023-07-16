## Style Guild:

- OpenApi Generate: https://github.com/OpenAPITools/openapi-generator
- swagger-typescript-api: https://github.com/acacode/swagger-typescript-api

## How Tos

- How to generate TS interface files by yaml files:
  ```
  npx swagger-typescript-api -p modules/shared/workflow/lib/api/workflow.yaml -t open-api-templates -o modules/shared/workflow/lib/api -n workflowApi
  ```
