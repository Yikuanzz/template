# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Full-stack application with Go backend (Gin + uber/fx) and React frontend (Vite + TypeScript).

## Commands

### Development (using Taskfile)
```bash
task backend:run    # Start backend (formats code first)
task frontend:run   # Start frontend (bun run dev)
task run           # Start both
task backend:swagger   # Generate Swagger docs
task backend:fmt       # Format Go code
```

### Frontend Commands
```bash
cd frontend
bun run dev        # Dev server
bun run build      # TypeScript check + Vite build
bun run lint       # ESLint
npx shadcn@latest add [component]  # Install shadcn/ui components
```

### Backend Commands
```bash
cd backend
go run app/cmd/main.go --env app/cmd/.env
go test ./...
swag init -g app/cmd/main.go -o app/docs --parseDependency --parseInternal --parseDepth 1
```

## Architecture

### Backend (Go)

Three-layer architecture with uber/fx dependency injection:

- **Handler** (`app/internal/handler/`): HTTP request/response, parameter binding, calls Logic
- **Logic** (`app/internal/logic/`): Business logic, calls Repo
- **Repo** (`app/internal/repo/`): Data access via GORM

Each module has: `*_handler.go`, `*_model.go`, `*_logic.go`, `*_repo.go`, `provider.go`

Key directories:
- `app/model/` - Database models (GORM)
- `app/types/dto/` - Data transfer objects
- `app/types/errorn/` - Error code definitions
- `app/server/middleware/` - Auth, CORS, API logging middleware
- `utils/` - Shared utilities (errorx, bind, handle, logs, etc.)

Interface pattern: Handler defines Logic interface, Logic defines Repo interface.

### Frontend (React + TypeScript)

- `src/api/` - API calls using Axios
- `src/components/ui/` - shadcn/ui components (install via CLI only)
- `src/pages/` - Page components
- `src/store/` - Zustand state management
- `utils/http.ts` - Axios instance with interceptors
- `utils/env.ts` - Environment variable access
- `types/` - TypeScript type definitions

Path alias: `@/` maps to `src/`

## Key Patterns

### Backend
- All handlers need Swagger annotations (`@Summary`, `@Router`, etc.)
- Use `errorx.New()` / `errorx.Wrap()` for errors with codes from `app/types/errorn/`
- Use `handle.Success()` / `handle.HandleErrorWithContext()` for responses
- Use `bind.ShouldBindJSON()` for request validation
- Always pass `context.Context` to functions
- Database queries must use `.WithContext(ctx)`

### Frontend
- Use shadcn/ui components exclusively (install via `npx shadcn@latest add`)
- Access env vars through `utils/env.ts`, not `import.meta.env` directly
- Use `http` instance from `utils/http.ts`, not raw axios
- Store persistence: implement `init()` method to restore from localStorage
- Use `cn()` from `@/lib/utils` for conditional class names
