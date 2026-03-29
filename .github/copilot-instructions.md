# APEX Dashboard AI Coding Instructions

## Architecture Overview
Full-stack market dashboard: Go backend (net/http, no framework) serving REST API at `/api/v1/*`, React 18 + TypeScript + Vite frontend. Data flows from external APIs (Polygon.io equities, CoinGecko crypto) through repository interface → handlers → JSON responses wrapped in `{"data": T}` envelope.

## Backend Patterns
- **Dependency Injection**: Pass config via struct fields, no global state. See `backend/cmd/server/main.go` for wiring.
- **Handlers**: Receive `http.ResponseWriter, *http.Request`; use `h.respond()` for success (`{"data": data}`) or `h.storeError()` for errors (`{"error": {"code": "...", "message": "..."}}`). Example: `backend/internal/handlers/overview.go`.
- **Repository**: Implement `repository.Store` interface for data access. Currently cached live data (60s TTL) over static fallback. See `backend/internal/repository/cache/cache.go`.
- **Models**: Plain structs in `backend/internal/models/` matching frontend TypeScript types in `frontend/src/types/`.
- **Errors**: Bubble up as values; wrap with `fmt.Errorf("context: %w", err)`. `repository.ErrUnavailable` maps to 503.
- **Go Style**: Follow `.github/skills/go-style-guide/SKILL.md` for all Go code: config-driven constructors, sentinel errors, table-driven tests, no package logging.

## Frontend Patterns
- **API Calls**: All fetch via `frontend/src/services/market.ts` which unwraps `{"data": T}`. Never call `fetch` directly in components.
- **State Management**: Use hooks in `frontend/src/hooks/useMarketData.ts` for loading/error/retry. No global state libraries.
- **Components**: Functional, arrow functions, PascalCase files. Colocate types in same file; shared in `types/`.
- **Styling**: Tailwind CSS only, dark terminal theme. No inline styles or CSS modules.

## Development Workflow
- **Backend**: `cd backend && go run ./cmd/server` (listens on `:8080` via `APEX_ADDR`). Tests: `go test ./...` (table-driven pattern). Lint: `golangci-lint run`.
- **Frontend**: `cd frontend && npm run dev` (proxies `/api` to backend). Type-check: `npx tsc --noEmit`. Lint: `npm run lint`. Build: `npm run build`.
- **Full Stack**: Start backend first, then frontend. Vite dev server handles CORS via proxy.
- **Environment**: Backend needs `POLYGON_API_KEY` for live data. Frontend uses `VITE_API_BASE_URL` for production.

## Key Conventions
- **Go**: `internal/` for private code. No `os.Getenv` scattered; load once in `config/`. Context for cancellation. `slog` for logging.
- **TypeScript**: Strict mode, no `any`. CamelCase JSON fields. Functional components only.
- **Git**: Conventional commits (`feat:`, `fix:`). Run tests/type-check before commit.
- **API**: RESTful, kebab-case paths. Standard HTTP codes. Validate requests at handler level.

## Common Pitfalls
- Backend: Forget `go mod tidy` after imports. Always check errors. Close `resp.Body` in HTTP clients.
- Frontend: Env vars must start with `VITE_`. Use `unknown` over `any` with type guards.
- Data: Live APIs rate-limited; cache prevents issues. Fallback to static data on failures.