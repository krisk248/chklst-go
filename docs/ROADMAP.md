# CHKLST Roadmap

## Current State (January 2026)
- Go backend with Fiber v3
- Vue 3 frontend (QA mode)
- SQLite database
- Single binary deployment

---

## Phase A: Backend Foundation

### New Database Models

#### Enhanced Deployment Fields
```go
ChangeTicket     string    // JIRA/ServiceNow ticket
RollbackOf       *uint     // Points to original if rollback
IsRollback       bool      // Flag for filtering
PipelineID       string    // CI/CD reference
PipelineTrigger  string    // "manual" | "webhook" | "scheduled"
LeadTimeMinutes  int       // Commit to deploy time
DeployDuration   int       // Deploy duration (seconds)
WebhooksSent     bool      // Notification status
```

#### New Models
- **WebhookConfig**: Slack/Teams webhook configuration
- **DORAMetrics**: Daily DORA metrics snapshot

### New API Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | /api/v1/deployments/:id/rollback | Create rollback |
| GET/POST/PUT/DELETE | /api/v1/webhooks | Webhook CRUD |
| POST | /api/v1/webhooks/test/:id | Test webhook |
| GET | /api/v1/metrics/dora | DORA metrics |
| GET | /api/v1/export/toml | Export as TOML |
| POST | /api/v1/import/toml | Import from TOML |

---

## Phase B: QA Vue Enhancements

### Quick Deploy Updates
- Change Ticket field
- Pipeline ID field
- Lead Time input
- "Save & Notify" button

### New Pages
- Webhook Settings
- DORA Metrics Dashboard

### History Enhancements
- Rollback button per deployment
- Filter by rollbacks
- Filter by change ticket

---

## Phase C: Production Mode (Svelte)

### Architecture
```
./chklst              → Vue UI (QA) - chklst.db
./chklst --production → Svelte UI (Prod) - chklst-prod.db
```

### Svelte UI Pages
- Dashboard (overview + DORA summary)
- Projects/Components management
- Deployments
- TOML Import/Export
- Metrics (DORA charts)
- Settings (webhooks)

---

## DORA Metrics

| Metric | Calculation |
|--------|-------------|
| Deployment Frequency | Deploys per day |
| Lead Time | Avg lead_time_minutes |
| Change Failure Rate | Failed / Total × 100% |
| Mean Time to Restore | Avg time to fix failures |

---

## Git Branches
- `main` - Stable release
- `feature/production-mode` - Svelte UI work
- `feature/qa-enhancements` - QA improvements

---

## Future Ideas
- CLI mode: `./chklst deploy --project BRHUB`
- Docker image
- Helm chart for Kubernetes
- LDAP/SSO integration
- Multi-user with roles
