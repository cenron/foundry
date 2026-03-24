# React Conventions

## Stack

Every React project uses this stack unless explicitly overridden:

- **Build tool:** Vite
- **Styling:** Tailwind CSS
- **Component library:** shadcn/ui
- **Server state:** TanStack Query (React Query)
- **API client:** Auto-generated from OpenAPI/Swagger specs

## Stack Freshness Check

Before installing or recommending any tool in this stack, verify it is still actively maintained. Check npm publish dates, GitHub activity, or official docs. If a tool appears abandoned (no releases in 12+ months, maintainer has signaled deprecation, or a widely adopted successor exists), **do not use it** — ask the user for a replacement decision before proceeding.

## Project Setup

When scaffolding a new React project:

1. Initialize with Vite (`npm create vite@latest` — React + TypeScript template)
2. Install and configure Tailwind CSS
3. Initialize shadcn/ui (`npx shadcn@latest init`)
4. Install TanStack Query (`@tanstack/react-query`)
5. Set up API code generation (see below)

## API Code Generation

Generate typed API clients and TanStack Query hooks from OpenAPI/Swagger specs using **@hey-api/openapi-ts**.

- Install: `@hey-api/openapi-ts` as a dev dependency.
- Config: `openapi-ts.config.ts` at the project root.
- Store the OpenAPI spec in the project (e.g., `docs/openapi.yaml`) or point at a remote URL.
- Add a `generate:api` script in `package.json` that runs `openapi-ts`.
- Generated code goes in `src/api/generated/` — never edit generated files directly.
- Use the TanStack Query plugin to generate type-safe query hooks directly.

## MANDATORY: Playwright E2E Testing

Every React project REQUIRES Playwright end-to-end tests. This is not optional — no React project is considered complete without E2E coverage.

### Setup

1. Install Playwright: `npm init playwright@latest`
2. Configure at least Chromium, Firefox, and WebKit browsers
3. Add `test:e2e` script to `package.json`
4. E2E tests live in an `e2e/` directory at the web root

### E2E Testing Rules

These rules exist because we shipped real bugs that E2E tests should have caught. Follow them exactly.

**1. Click, don't navigate.** E2E tests must reach pages by clicking links and buttons — the way a real user does. `page.goto()` is allowed ONLY for the initial entry point (e.g., the login page or home page). Every subsequent page transition must happen through UI interaction. `page.goto()` bypasses client-side routing entirely — a broken React Router pattern will be invisible to tests that use `goto()`.

**2. Assert API responses, not just UI rendering.** Every test that triggers an API call must verify the call succeeded. Pages can render stale cached data or optimistic UI even when the API returned an error. Use `page.waitForResponse()` to check HTTP status, assert success messages, or navigate away and back to confirm data persisted.

**3. Scope selectors to avoid ambiguity.** Never use bare `getByRole("link", { name: "Sign up" })` when the same text appears in multiple places (e.g., nav bar + login form). Scope to a landmark: `page.getByRole("banner").getByRole("link", { name: "Sign up" })`. Use `{ exact: true }` when text like "Followers" also appears inside "Followers (3)".

**4. Set up response listeners before the action, not after.** `page.waitForResponse()` must be called BEFORE the click that triggers the request. Otherwise the response fires before the listener is attached and the test times out.

**5. Don't assume element roles.** Verify what HTML element a component actually renders. shadcn `CardTitle` renders a `<div>`, not a heading — `getByRole("heading")` won't find it. Read the component source or use the accessibility snapshot to confirm roles before writing selectors.

**6. Test route matching for non-trivial patterns.** Any route with dynamic segments or special characters (e.g., `/:username` for `/@alice` URLs) needs a unit test using `matchRoutes()` to prove the pattern matches expected URLs and doesn't swallow static routes.

**7. Test auth integration.** At least one unit test must verify that authenticated API calls send the correct `Authorization` header. Mock `fetch` and inspect the headers — don't just test the config object.

**8. Handle shared mutable state in parallel tests.** When multiple browsers run in parallel against the same backend, tests that mutate shared state (follow/unfollow, create/delete) will race. Use `registerAndLogin(page, prefix)` to create unique users per test run. Keep tests fully parallel — don't use sequential browser projects (they double test time without fixing webkit auth timing issues).

**9. Wait for auth to settle after page.goto().** After any `page.goto()`, wait for `networkidle` AND verify auth-only UI elements are visible before interacting with authenticated features. Webkit's token refresh is slower — clicking before auth settles causes 401 errors.

## Component Patterns

- Use shadcn/ui components as the base — don't build custom UI primitives that shadcn already provides.
- Compose shadcn components into feature-specific components inside feature folders.
- Tailwind for all styling — no CSS modules, styled-components, or inline style objects.

## TanStack Query

- All server state goes through TanStack Query — no `useEffect` + `useState` for data fetching.
- Query keys follow a consistent factory pattern (e.g., `queryKeys.users.list()`, `queryKeys.users.detail(id)`).
- Mutations use `useMutation` with proper cache invalidation via `onSuccess`.
- Configure a `QueryClient` with sensible defaults at the app root.
