## Context

`backend-api` already mounts the protected Connect procedures and validates
Clerk bearer tokens correctly, but the browser cannot call them directly in the
local frontend-native workflow unless the API answers CORS preflight requests.
The current runtime has no browser CORS layer, so requests from the documented
frontend origins fail before the handler or auth interceptor is reached.

## Goals / Non-Goals

**Goals:**

- Allow the documented local frontend SPA origins to call the protected Connect
  procedures from the browser.
- Keep the CORS behavior narrow and local-only.
- Cover both successful and rejected preflight behavior in runtime tests.

**Non-Goals:**

- Add broad production CORS configuration or wildcard origin handling.
- Change Clerk auth verification, Connect procedure paths, or database logic.

## Decisions

### Add a small router-level browser CORS middleware

- Apply one router middleware ahead of the mounted Connect handlers.
- Only treat requests as browser CORS traffic when an `Origin` header is
  present.
- Handle preflight requests by returning the allow-origin, allow-methods, and
  allow-headers response directly.

Why:

- This keeps browser cross-origin behavior in the HTTP runtime layer where it
  belongs and avoids pushing CORS concerns into handlers or auth interceptors.

### Allow only the documented local frontend origins

- In `local`, allow `http://localhost:3000` and `http://127.0.0.1:3000`.
- Reject unknown origins for preflight requests.
- Leave non-local environments unchanged by default.

Why:

- The local stack already documents those frontend origins. Keeping the
  allowlist explicit reduces accidental exposure and matches the current
  development model.

## Risks / Trade-offs

- [A developer runs the frontend from an unexpected origin] -> The browser call
  will still fail until that origin is intentionally documented and added.
- [Future non-local browser deployments need different CORS behavior] -> Keep
  this baseline deliberately local-only so later environment-specific behavior
  can be added explicitly instead of inheriting a permissive default.

## Migration Plan

1. Add local-only browser CORS middleware in the API runtime.
2. Add route tests for allowed and rejected preflight requests.
3. Update runtime docs and the `api-runtime` spec summary.
4. Validate and archive the change after implementation is confirmed.
