## Why

The local authenticated frontend now reaches the protected API from a browser,
but the API runtime did not respond to browser CORS preflight requests for the
protected Connect procedures. That blocks the protected profile bootstrap flow
before auth verification can even run.

## What Changes

- Add local-only browser CORS handling in the API runtime for the documented
  frontend SPA origins.
- Allow CORS preflight requests to the protected Connect procedures to succeed
  for the local frontend origins and reject unknown origins.
- Update runtime docs and tests so the local browser-to-API contract is
  explicit.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `api-runtime`: local browser access to protected Connect procedures now
  includes the required CORS preflight behavior for the documented frontend SPA
  origins.

## Impact

- Affected code: `internal/api/`, runtime tests, runtime docs, and the
  `api-runtime` OpenSpec capability.
- Affected systems: local frontend-native development against the browser-facing
  API.
