# AGENTS.md

## Project goal
Build a production-ready backend service for a simple H5 landing page and admin panel.

## Business context
This project serves a H5 page and an admin management panel.

Public side:
- Fetch one content item by date

Admin side:
- Login
- Create content
- Update content
- Delete content
- List content with pagination

## Source of truth
- Read and follow docs/api-spec.md strictly
- If examples conflict, prefer the normalized field names in docs/api-spec.md

## API field rules
- Use `bg_url` as the only background image field
- Do not use `bg`
- `tags` must be a string array
- `date` format must be `YYYY-MM-DD`

## Tech stack
- Go 1.22+
- Gin
- GORM
- SQLite for local/dev
- MySQL for prod
- JWT for admin auth

## Repo layout
- cmd/server: application entry
- internal/config: configuration loading
- internal/model: database models and DTOs
- internal/repo: DB access
- internal/service: business logic
- internal/handler: HTTP handlers
- internal/middleware: auth, logging, recovery, CORS
- internal/router: route registration
- scripts: helper scripts
- deploy: deployment configs

## Engineering conventions
- Keep handlers thin
- Put business logic in service layer
- Put DB operations in repo layer
- Add request validation
- Return stable JSON structures
- Do not hardcode secrets
- Use environment variables for configuration
- Add seed admin initialization
- Write README commands that actually work

## Done when
- Project builds successfully
- `go test ./...` passes
- All API routes in docs/api-spec.md are implemented
- Admin routes are protected by JWT except login
- README explains how to run locally and deploy

## Workflow
- First analyze and propose a plan
- Then implement in small steps
- After each step, run build/tests and report what changed
