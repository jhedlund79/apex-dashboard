# Project Overview

Full-stack web application with a Go (Golang) REST API backend and a React + Vite + TypeScript frontend.

## Architecture

- **Backend**: Go 1.22+ with standard library `net/http` 
- **Frontend**: React 18 + TypeScript + Vite
- **Database**: PostgreSQL (update as needed)
- **Communication**: Frontend proxied to backend API at `/api/`
- **Auth**: JWT-based authentication (update as needed)

```
project-root/
├── CLAUDE.md
├── backend/
│   ├── cmd/                  # Application entrypoints
│   │   └── server/
│   │       └── main.go       # Main server entry
│   ├── internal/             # Private applicati22on code
│   │   ├── handlers/         # HTTP handlers
│   │   ├── middleware/        # HTTP middleware
│   │   ├── models/           # Data models and DB schema
│   │   ├── repository/       # Database access layer
│   │   ├── services/         # Business logic
│   │   └── config/           # Configuration loading
│   ├── pkg/                  # Public shared libraries
│   ├── migrations/           # SQL migration files
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── components/       # Reusable UI components
│   │   ├── pages/            # Route-level page components
│   │   ├── hooks/            # Custom React hooks
│   │   ├── services/         # API client functions
│   │   ├── types/            # TypeScript type definitions
│   │   ├── utils/            # Helper functions
│   │   ├── store/            # State management
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── public/
│   ├── index.html
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── tailwind.config.ts
│   └── package.json
└── docker-compose.yml
```

## Development Commands

### Backend (run from `backend/`)

```bash
# Run the server in development
go run ./cmd/server

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a specific package's tests
go test -v ./internal/handlers/...

# Run tests with coverage
go test -cover ./...

# Build the binary
go build -o bin/server ./cmd/server

# Lint (requires golangci-lint)
golangci-lint run

# Tidy dependencies
go mod tidy

# Run database migrations (update tool as needed)
migrate -path migrations -database "$DATABASE_URL" up
```

### Frontend (run from `frontend/`)

```bash
# Install dependencies
npm install

# Start dev server (default: http://localhost:5173)
npm run dev

# Type-check without emitting
npx tsc --noEmit

# Lint
npm run lint

# Run tests
npm run test

# Build for production
npm run build

# Preview production build
npm run preview
```

### Full Stack

```bash
# Start all services (DB, backend, frontend)
docker-compose up

# Start with rebuild
docker-compose up --build
```

## Code Style & Conventions

### Go Backend

- Follow standard Go conventions: `gofmt`, `goimports`, effective Go idioms
- Use `internal/` for all private application code — never import from another project
- Handlers receive `http.ResponseWriter` and `*http.Request`; return JSON via a shared `respond` helper
- Errors bubble up as values — use `fmt.Errorf("doing X: %w", err)` for wrapping
- No global state; pass dependencies via struct fields (dependency injection without a framework)
- Table-driven tests: use `[]struct{ name string; ... }` pattern with `t.Run`
- Name test files `*_test.go` next to the code they test
- Database access goes through the `repository` package — handlers never import `database/sql` directly
- Use context (`context.Context`) for cancellation and request-scoped values
- Environment config loaded once at startup via `config` package; no `os.Getenv` calls scattered in code

### React Frontend

- Functional components only — no class components
- Use TypeScript strict mode (`"strict": true` in tsconfig)
- Named exports for components, default export only for pages
- Colocate component-specific types in the same file; shared types in `types/`
- API calls go through `services/` — components never call `fetch` directly
- Use React Query (TanStack Query) for server state; local state with `useState`/`useReducer`
- Tailwind CSS for styling — no inline style objects, no CSS modules
- File naming: PascalCase for components (`UserCard.tsx`), camelCase for utilities (`formatDate.ts`)
- Prefer `const` arrow functions for components: `const UserCard = () => { ... }`
- All user-facing strings should be extracted if i18n is planned

## API Conventions

- RESTful routes under `/api/v1/`
- Use kebab-case for URL paths: `/api/v1/user-profiles`
- Use camelCase for JSON request/response fields
- Standard response envelope:
  ```json
  { "data": { ... }, "meta": { "page": 1, "total": 100 } }
  ```
- Error responses:
  ```json
  { "error": { "code": "NOT_FOUND", "message": "Resource not found" } }
  ```
- HTTP status codes: 200 OK, 201 Created, 204 No Content, 400 Bad Request, 401 Unauthorized, 403 Forbidden, 404 Not Found, 500 Internal Server Error
- Always validate request bodies at the handler level before passing to services

## Git Workflow

- Branch naming: `feature/short-description`, `fix/short-description`, `chore/short-description`
- Commit messages: conventional commits format (`feat:`, `fix:`, `chore:`, `docs:`, `test:`, `refactor:`)
- Run `go test ./...` and `npx tsc --noEmit` before committing
- Keep PRs focused — one feature or fix per branch

## Environment Variables

Backend expects (set via `.env` or environment):

```
PORT=8080
DATABASE_URL=postgres://user:pass@localhost:5432/dbname?sslmode=disable
JWT_SECRET=your-secret-key
ENVIRONMENT=development
```

Frontend uses Vite env vars (prefixed with `VITE_`):

```
VITE_API_BASE_URL=http://localhost:8080
```

## Common Pitfalls

- **Go**: Don't forget to close `resp.Body` when making HTTP client calls
- **Go**: Always check errors — never use `_` for error returns unless explicitly justified
- **Go**: Run `go mod tidy` after adding/removing imports
- **Vite**: Env vars must start with `VITE_` to be exposed to the client
- **Vite proxy**: Configure `vite.config.ts` proxy for `/api` to avoid CORS in dev
- **TypeScript**: Never use `any` — define proper types or use `unknown` with type guards
- **React**: Cleanup effects (timers, subscriptions) in `useEffect` return function

## Verification Checklist

After making changes, verify with:

1. `cd backend && go test ./...` — all Go tests pass
2. `cd backend && golangci-lint run` — no lint issues
3. `cd frontend && npx tsc --noEmit` — no type errors
4. `cd frontend && npm run lint` — no lint issues
5. `cd frontend && npm run build` — production build succeeds