# chklst-go

**Deployment Checklist Tool** - Go backend with embedded Python microservice for Excel/PDF generation.

## Features

✅ **Core Features**:
- Projects, Components, and Deployments CRUD
- Library/Presets management (developers, servers, environments)
- Filtering by project, date (month/year)
- Real-time WebSocket updates
- Embedded Vue 3 frontend

✅ **NEW Features** (Go Bonuses):
- 📦 Database Export/Import - Backup entire SQLite database
- ⚙️ Settings Export/Import - Backup library presets to JSON
- 💚 Health Check - `/health` endpoint for monitoring
- 📝 Structured JSON Logging - Better debugging
- 🔄 Auto-Backup - Daily automatic database backups
- 🔍 Request Tracing - Unique IDs for every request

✅ **Python Microservice**:
- Excel export (pandas + openpyxl)
- PDF reports (reportlab + matplotlib)

## Quick Start

```bash
# Build
./build/build.sh

# Run
./chklst

# Access
http://localhost:8000
```

## Architecture

```
chklst-go (Single Binary ~50MB)
├── Go Backend (API, DB, WebSocket)
├── Vue 3 Frontend (Embedded)
└── Python Microservice (Excel/PDF)
```

## API Endpoints

### Projects
- `GET /api/v1/projects` - List all projects
- `POST /api/v1/projects` - Create project
- `GET /api/v1/projects/:id` - Get project
- `PUT /api/v1/projects/:id` - Update project
- `DELETE /api/v1/projects/:id` - Delete project

### Components
- `POST /api/v1/projects/:projectId/components` - Create component
- `PUT /api/v1/projects/:projectId/components/:componentId` - Update component
- `DELETE /api/v1/projects/:projectId/components/:componentId` - Delete component

### Deployments
- `GET /api/v1/deployments` - List deployments (with filters)
- `POST /api/v1/deployments` - Create deployment
- `GET /api/v1/deployments/:id` - Get deployment
- `PUT /api/v1/deployments/:id` - Update deployment
- `DELETE /api/v1/deployments/:id` - Delete deployment

### Library/Presets
- `GET /api/v1/library` - Get library
- `POST /api/v1/library/developers` - Add developer
- `DELETE /api/v1/library/developers/:name` - Remove developer
- Similar endpoints for servers and environments

### Reports
- `GET /api/v1/reports/excel?month=1&year=2025` - Export to Excel
- `GET /api/v1/reports/pdf?month=1&year=2025` - Export to PDF
- `GET /api/v1/reports/stats?month=1&year=2025` - Get statistics

### Admin & Monitoring
- `GET /health` - Health check
- `POST /api/v1/admin/backup/database` - Backup database
- `POST /api/v1/admin/restore/database` - Restore from backup
- `POST /api/v1/admin/export/settings` - Export settings to JSON
- `POST /api/v1/admin/import/settings` - Import settings from JSON
- `GET /api/v1/admin/backups` - List all backups

## Database Migration

Your existing `chklst.db` file works out of the box! Just copy it:

```bash
cp ../chklst.db ./chklst.db
```

## Development

```bash
# Install dependencies
go mod download

# Run in development
go run cmd/chklst/main.go

# Build
go build -o chklst cmd/chklst/main.go

# Cross-compile
GOOS=windows GOARCH=amd64 go build -o chklst.exe cmd/chklst/main.go
GOOS=darwin GOARCH=amd64 go build -o chklst-mac cmd/chklst/main.go
```

## Configuration

Set via environment variables or defaults:

- `DB_PATH` - Database file path (default: `./chklst.db`)
- `BACKUP_DIR` - Backup directory (default: `./backups`)
- `PORT` - Server port (default: `8000`)
- `LOG_LEVEL` - Logging level (default: `INFO`)
- `AUTO_BACKUP_HOURS` - Auto-backup interval (default: `24`)

## License

Personal use project by Kannan.
