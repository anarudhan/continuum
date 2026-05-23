# 🏠 Self-Hosting Guide

Run Continuum on your own infrastructure.

---

## Docker Compose (Recommended)

The fastest path to production.

### 1. Clone and start

```bash
git clone https://github.com/anarudhan/continuum.git
cd continuum
docker compose up -d
```

This brings up:
- **Continuum API** on port `8080`
- **PostgreSQL 16** with pgvector on port `5432`
- **Redis 7** on port `6379`
- **React Dashboard** served alongside the API

### 2. Verify

```bash
curl http://localhost:8080/health
# → {"status":"healthy","version":"0.1.0"}
```

### 3. Create your first admin agent

The server auto-creates a default agent on first boot. Check the logs:

```bash
docker logs continuum-api | grep "API Key hint"
# → API Key hint: ctm_...XXXX (save this — shown once)
```

Use this key to create additional agents via `POST /admin/agents`.

---

## Environment Variables

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `CONTINUUM_DATABASE_URL` | `postgres://continuum:***@localhost:5432/continuum?sslmode=disable` | ✅ | PostgreSQL connection string |
| `CONTINUUM_REDIS_URL` | `redis://localhost:6379` | ✅ | Redis connection string |
| `CONTINUUM_PORT` | `8080` | ❌ | HTTP server port |

---

## Production Checklist

### Database

- [ ] Use a strong password for PostgreSQL
- [ ] Enable SSL (`sslmode=require`)
- [ ] Run migrations on every deploy (`make migrate`)
- [ ] Set up automated backups (pg_dump daily)

### Redis

- [ ] Enable AUTH if exposed to network
- [ ] Set `maxmemory-policy allkeys-lru`
- [ ] Use Redis Sentinel for HA if needed

### API

- [ ] Change default admin API key immediately
- [ ] Run behind a reverse proxy (Nginx, Traefik, Caddy)
- [ ] Enable HTTPS (Let's Encrypt)
- [ ] Set rate limits appropriate for your load

### Monitoring

- [ ] Scrape `/metrics` endpoint (Prometheus format)
- [ ] Alert on `/ready` failures
- [ ] Log aggregation (Loki, Fluentd, or CloudWatch)

---

## Kubernetes

Example deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: continuum
spec:
  replicas: 2
  selector:
    matchLabels:
      app: continuum
  template:
    metadata:
      labels:
        app: continuum
    spec:
      containers:
        - name: api
          image: ghcr.io/anarudhan/continuum:latest
          ports:
            - containerPort: 8080
          env:
            - name: CONTINUUM_DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: continuum-secrets
                  key: database-url
            - name: CONTINUUM_REDIS_URL
              valueFrom:
                secretKeyRef:
                  name: continuum-secrets
                  key: redis-url
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
```

---

## Building from Source

```bash
# Backend
cd /opt/continuum
go build -o continuum ./cmd/continuum

# Frontend
cd web
npm ci
npm run build

# Run
./continuum
```

Requires Go 1.24+ and Node 22+.

---

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `connection refused` on `/health` | API not started | Check `docker ps`, verify logs |
| `password authentication failed` | Wrong DB credentials | Verify `CONTINUUM_DATABASE_URL` |
| `migrations not found` | Missing migration files | Ensure `internal/models/migrations` is copied into image |
| `rate limited` | Too many requests | Increase limit in Redis or tune middleware |
| WebSocket disconnects | Reverse proxy timeout | Set `proxy_read_timeout 86400s` in Nginx |

---

## Security Hardening

- Rotate API keys monthly via `POST /admin/agents` + delete old ones
- Run PostgreSQL and Redis on internal networks only
- Use network policies (K8s) or security groups (cloud) to restrict access
- Enable audit logging for all admin operations
- Review `GET /agents` output regularly for unexpected agents
