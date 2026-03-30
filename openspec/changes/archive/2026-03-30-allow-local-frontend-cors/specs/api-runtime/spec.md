## MODIFIED Requirements

### Requirement: Runtime endpoint surface

The API runtime SHALL expose metadata and health endpoints at `GET /`,
`GET /healthz`, and `GET /readyz`, and it SHALL mount the generated Connect
procedures `POST /blueprint.user.v1.UserService/EnsureCurrentUserProfile` and
`POST /blueprint.user.v1.UserService/GetCurrentUser`. In `local`, the runtime
MUST allow browser CORS requests from the documented frontend SPA origins so
the frontend can call those procedures directly.

#### Scenario: Protected Connect procedures are mounted

- **WHEN** the API runtime is started with valid configuration
- **THEN** the generated `UserService` procedures are available at their Connect
  HTTP paths

#### Scenario: Local frontend origin can preflight protected Connect procedures

- **WHEN** a browser from the documented local frontend origin sends a CORS
  preflight request to a protected Connect procedure path
- **THEN** the API runtime returns the corresponding allow-origin,
  allow-methods, and allow-headers response needed for the frontend SPA call
