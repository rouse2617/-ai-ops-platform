# Build Fix Instructions

## Current Build Errors

The build is failing due to incomplete implementations in:
1. `internal/rag/service.go` - Missing RAG components
2. `internal/prometheus/*.go` - Various compilation errors

## Quick Fix Options

### Option 1: Comment Out Incomplete Features (Recommended)

Since RAG and Prometheus are optional features, temporarily disable them:

1. Move RAG package:
```bash
mv internal/rag internal/rag.disabled
```

2. Fix Prometheus errors or move it:
```bash
mv internal/prometheus internal/prometheus.disabled
```

3. Comment out Prometheus handler in router.go (lines 222-234)

### Option 2: Fix Prometheus Errors

Fix these specific issues in internal/prometheus/:

1. **init.go** - Add Prometheus config to config.Config struct
2. **client.go:85** - Fix GetLabelValues return signature
3. **capacity_planning.go:103** - Remove unused variable
4. **capacity_planning.go:215** - Fix type assertion
5. **semantic.go:33** - Fix function argument type

### Option 3: Minimal Build

Build only the core features:

```bash
# Temporarily rename problematic packages
mv internal/rag internal/rag.bak
mv internal/prometheus internal/prometheus.bak

# Remove imports from files that use them
# Then build
go build -o bin/server.exe ./cmd/server
```

## Verification

After fixing, verify the three advanced features work:

1. Start server: `./bin/server.exe`
2. Test health check: `curl -X POST http://localhost:8080/api/health/check -d '{"host_id":"host1"}'`
3. Test trend analysis: `curl -X POST http://localhost:8080/api/trends/analyze -d '{"host_id":"host1"}'`
4. Open frontend and verify UI components render

## Files Successfully Implemented

### Backend (Complete)
- internal/service/health_service.go
- internal/service/trend_service.go
- internal/repository/health_check.go
- internal/repository/trend_prediction.go
- internal/api/handler/health.go
- internal/api/handler/trend.go
- internal/model/operation.go (models added)

### Frontend (Already Existed)
- web/src/components/chat/HealthReportCard.vue
- web/src/components/chat/TrendWarningBubble.vue
- web/src/components/chat/RiskWarningDialog.vue

### Integration (Complete)
- internal/api/router.go (routes added)
- cmd/server/main.go (repositories initialized)
- internal/repository/db.go (migrations added)
