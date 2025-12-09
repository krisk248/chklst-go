# 🚀 Quick Start - chklst-go with Vue Frontend

## ✅ What's Set Up

Your Vue 3 frontend is now integrated with the Go backend! Here's what you have:

```
chklst-go/
├── chklst                    ← 14MB Go binary (backend + serves frontend)
├── chklst.db                ← Your existing database
├── frontend/                ← Vue 3 frontend
│   ├── dist/               ← Built frontend (ready to serve!)
│   │   ├── index.html
│   │   └── assets/         ← CSS & JS bundles
│   ├── src/                ← Vue source code
│   └── package.json
└── cmd/chklst/main.go      ← Updated to serve frontend
```

## 🎯 How It Works

**Architecture:**
```
Browser ──────────────────────> Go Server (Port 8000)
    │                                │
    │ GET /                          │ Serves index.html (Vue app)
    │ GET /assets/...                │ Serves CSS/JS files
    │                                │
    │ POST /api/v1/projects          │ Go API handlers
    │ GET /api/v1/deployments        │ Database operations
    │ GET /health                    │ Health check
```

**What Happens:**
1. Go serves your Vue app from `frontend/dist/`
2. Vue Router handles client-side routing (no page reloads!)
3. Vue makes API calls to `/api/v1/*` (same server)
4. Go processes API requests and talks to SQLite database

## 🏃 Running the Application

### Option 1: Just Run It! (Simplest)

```bash
cd /home/kannan/Projects/Active/chklst/chklst-go

# Start the server
./chklst
```

**Access your app:**
- **Full App**: http://localhost:8000
- **Health Check**: http://localhost:8000/health
- **API**: http://localhost:8000/api/v1/projects

### Option 2: Development Mode (with auto-rebuild)

```bash
# Terminal 1: Run Go backend
make run
# OR: go run cmd/chklst/main.go

# Terminal 2 (if you need to rebuild Vue):
cd frontend
npm run dev
```

## 📱 What You'll See

### 1. Open Browser

```bash
# Your default browser should open automatically, or go to:
http://localhost:8000
```

### 2. You'll See Your Vue App!

The Vue frontend will load with:
- ✅ Projects list
- ✅ Deployments dashboard
- ✅ Components management
- ✅ Library/presets
- ✅ All your existing data from chklst.db

### 3. Check the Console

**Go Backend Logs:**
```json
{"timestamp":"2025-12-09T15:00:00+04:00","level":"INFO","message":"Starting chklst-go application"}
{"timestamp":"2025-12-09T15:00:00+04:00","level":"INFO","message":"Database initialized successfully"}
{"timestamp":"2025-12-09T15:00:00+04:00","level":"INFO","message":"Starting HTTP server","data":{"port":"8000"}}
```

**Browser Console (F12):**
```
Vue app loaded ✓
API connected ✓
Projects: 5
Deployments: 23
```

## 🧪 Testing the Integration

### Test 1: Health Check

```bash
curl http://localhost:8000/health
```

**Expected:**
```json
{
  "status": "healthy",
  "database": "connected",
  "timestamp": "2025-12-09T15:00:00+04:00",
  "version": "1.0.0"
}
```

### Test 2: API Endpoints

```bash
# Get all projects
curl http://localhost:8000/api/v1/projects

# Create a project
curl -X POST http://localhost:8000/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Project","build_server":"Jenkins"}'

# Get deployments with filters
curl "http://localhost:8000/api/v1/deployments?month=12&year=2025"
```

### Test 3: Frontend Routes

**These should all work (no 404s):**
- http://localhost:8000/ (home)
- http://localhost:8000/projects (projects page)
- http://localhost:8000/deployments (deployments page)
- http://localhost:8000/library (library page)

**SPA Magic:** Vue Router handles these client-side, no server reload!

## 🔧 Rebuilding Frontend (If Needed)

If you make changes to the Vue code:

```bash
cd frontend

# Install dependencies (if not already)
npm install

# Development mode (hot reload)
npm run dev

# Production build (for Go to serve)
npm run build
# This updates frontend/dist/
```

After rebuilding, **restart Go server** to serve new files:
```bash
# Stop with Ctrl+C, then:
./chklst
```

## 📊 What's Different From Python Version

### Same:
- ✅ Same Vue 3 frontend
- ✅ Same database (chklst.db)
- ✅ Same API endpoints
- ✅ Same features

### Better:
- ✅ **Single server** (was Python + Vue dev server)
- ✅ **One port** (8000 only, was 8000 + 5173)
- ✅ **No CORS issues** (same origin now!)
- ✅ **Production ready** (static file serving optimized)
- ✅ **14MB binary** (vs 100+ Python files)

## 🎨 Frontend Features Working

Your Vue app has these working features:

**Pages:**
1. **Dashboard** - Deployment overview
2. **Projects** - CRUD for projects
3. **Deployments** - List, filter, create deployments
4. **Components** - Manage project components
5. **Library** - Developers, servers, environments presets

**Features:**
- ✅ Vue Router (client-side routing)
- ✅ Pinia stores (state management)
- ✅ Axios/Fetch API calls to Go backend
- ✅ Tailwind CSS styling
- ✅ TypeScript type safety

## 🐛 Troubleshooting

### Issue: "Cannot GET /"

**Solution:**
```bash
# Make sure you're in the right directory
cd /home/kannan/Projects/Active/chklst/chklst-go

# Check frontend/dist exists
ls -la frontend/dist/

# Rebuild if needed
cd frontend && npm run build && cd ..

# Restart server
./chklst
```

### Issue: "404 Not Found" for CSS/JS

**Check:**
```bash
# Verify assets exist
ls frontend/dist/assets/

# Should show:
# index-*.css
# index-*.js
```

**Fix:** Rebuild frontend
```bash
cd frontend
npm run build
```

### Issue: API calls fail (CORS errors)

**Check:** Make sure you're accessing via http://localhost:8000 (not localhost:5173)

The Go server serves both frontend AND API, no CORS needed!

### Issue: Changes not showing

**Two possibilities:**

1. **Frontend changes:** Rebuild Vue
   ```bash
   cd frontend && npm run build
   ```

2. **Backend changes:** Rebuild Go
   ```bash
   go build -o chklst cmd/chklst/main.go
   ```

Then restart: `./chklst`

## 📸 Expected Screenshots

### 1. Terminal (Go Server Running)

```
{"timestamp":"2025-12-09T15:00:00+04:00","level":"INFO","message":"Starting chklst-go application"}
{"timestamp":"2025-12-09T15:00:00+04:00","level":"INFO","message":"Configuration loaded","data":{"port":"8000"}}
{"timestamp":"2025-12-09T15:00:00+04:00","level":"INFO","message":"Database initialized successfully"}
{"timestamp":"2025-12-09T15:00:00+04:00","level":"INFO","message":"Starting HTTP server","data":{"port":"8000"}}

    _______ __
   / ____(_) /_  ___  _____
  / /_  / / __ \/ _ \/ ___/
 / __/ / / /_/ /  __/ /
/_/   /_/_.___/\___/_/          v3.0.0-beta.3
--------------------------------------------------
INFO Server started on: 	http://127.0.0.1:8000
INFO Total handlers count: 	27
```

### 2. Browser (Vue App Loaded)

You'll see your Vue deployment checklist interface with:
- Navigation menu
- Project cards
- Deployment table
- Filter controls
- All your existing data

### 3. Browser DevTools Network Tab

```
Status  Method  File                Type        Size
200     GET     /                   document    506 B
200     GET     /assets/index-*.js  javascript  252 KB
200     GET     /assets/index-*.css stylesheet  22 KB
200     GET     /api/v1/projects    xhr         1.2 KB
200     GET     /api/v1/deployments xhr         5.4 KB
```

## 🎉 Success Checklist

✅ **Running:**
- [ ] Go server starts without errors
- [ ] Logs show "Server started on: http://127.0.0.1:8000"
- [ ] 27 handlers registered

✅ **Frontend:**
- [ ] Browser opens to http://localhost:8000
- [ ] Vue app loads (no blank page)
- [ ] Navigation works
- [ ] Data displays from database

✅ **API:**
- [ ] /health returns healthy status
- [ ] /api/v1/projects returns your projects
- [ ] Can create/update/delete projects
- [ ] Deployments filtering works

✅ **Integration:**
- [ ] No CORS errors in console
- [ ] API calls succeed (200 OK)
- [ ] Data saves to database
- [ ] Page refreshes maintain state

## 🚀 Next Steps

1. **Try it now!** Run `./chklst` and open http://localhost:8000
2. **Make changes:** Edit Vue components, rebuild, refresh
3. **Add features:** Both backend (Go) and frontend (Vue) are ready for changes
4. **Deploy:** Single 14MB binary + frontend/dist folder = production ready!

## 📞 Quick Commands Reference

```bash
# Start server
./chklst

# Rebuild Go
make dev

# Rebuild Vue
cd frontend && npm run build

# Check logs
./chklst 2>&1 | jq .

# Test API
curl http://localhost:8000/health
curl http://localhost:8000/api/v1/projects

# Stop server
Ctrl + C
```

---

**Ready?** Run `./chklst` and enjoy your fully integrated Go + Vue application! 🎊
