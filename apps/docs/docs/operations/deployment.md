# Deployment Guide

Comprehensive guide for deploying Real Staging AI to production environments.

## Overview

Real Staging AI is a containerized application designed for flexible deployment across various platforms. **This guide recommends Render for compute/database and Backblaze B2 for storage** as the primary production deployment strategy, with alternative options documented for flexibility.

### Architecture Components

The application consists of:

- **API Service** - Go HTTP server (Echo framework)
- **Worker Service** - Go background job processor
- **PostgreSQL Database** - Primary data store (users, projects, images, subscriptions)
- **Redis** - Job queue and caching
- **S3-Compatible Storage** - Image and result storage (AWS S3, MinIO, etc.)
- **OpenTelemetry Collector** - Optional, recommended for observability

### Deployment Checklist

Before deploying:

- [x] Choose deployment platform: **Render (recommended)**
- [ ] Provision PostgreSQL database on Render
- [ ] Provision Redis instance on Render
- [ ] Set up Backblaze B2 bucket for storage
- [ ] Configure Auth0 application for production domain
- [ ] Configure Stripe webhook endpoint for production
- [ ] Prepare secrets (database credentials, API keys, etc.)
- [ ] Configure custom domain and SSL certificates
- [ ] Set up monitoring and alerting
- [ ] Plan backup strategy
- [ ] Test rollback procedures

## Prerequisites

### Required Services

1. **PostgreSQL 14+** - Managed service recommended:
   - AWS RDS PostgreSQL
   - Google Cloud SQL
   - Azure Database for PostgreSQL
   - DigitalOcean Managed Databases
   - Supabase

2. **Redis 6+** - Managed service recommended:
   - AWS ElastiCache
   - Google Cloud Memorystore
   - Azure Cache for Redis
   - Redis Cloud
   - Upstash

3. **S3-Compatible Storage**:
   - **Backblaze B2 (recommended)** - Cost-effective, S3-compatible
   - AWS S3
   - Google Cloud Storage (S3-compatible mode)
   - DigitalOcean Spaces
   - Cloudflare R2
   - MinIO (self-hosted)

### External Services

1. **Auth0 Account** - [auth0.com](https://auth0.com)
   - Configure production application
   - Set callback URLs to production domain
   - Generate API credentials

2. **Stripe Account** - [stripe.com](https://stripe.com)
   - Complete business verification
   - Switch to Live Mode
   - Configure webhook endpoint
   - Get live API keys

3. **Replicate Account** - [replicate.com](https://replicate.com)
   - Get API token
   - Ensure sufficient credits for production load

## Building Docker Images

### Build Locally

Build images from the project root:

```bash
# Build API service
docker build -t realstaging/api:latest ./apps/api

# Build Worker service
docker build -t realstaging/worker:latest ./apps/worker

# Build with specific version tag
VERSION=1.0.0
docker build -t realstaging/api:${VERSION} ./apps/api
docker build -t realstaging/worker:${VERSION} ./apps/worker
```

### Multi-Platform Builds

For deploying to ARM-based servers or mixed architectures:

```bash
# Enable Docker buildx
docker buildx create --name multiarch --use

# Build for multiple platforms
docker buildx build --platform linux/amd64,linux/arm64 \
  -t realstaging/api:latest \
  --push \
  ./apps/api

docker buildx build --platform linux/amd64,linux/arm64 \
  -t realstaging/worker:latest \
  --push \
  ./apps/worker
```

### Push to Registry

```bash
# Docker Hub
docker push realstaging/api:latest
docker push realstaging/worker:latest

# GitHub Container Registry
docker tag realstaging/api:latest ghcr.io/yourorg/realstaging-api:latest
docker push ghcr.io/yourorg/realstaging-api:latest

# AWS ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 123456789.dkr.ecr.us-east-1.amazonaws.com
docker tag realstaging/api:latest 123456789.dkr.ecr.us-east-1.amazonaws.com/realstaging-api:latest
docker push 123456789.dkr.ecr.us-east-1.amazonaws.com/realstaging-api:latest
```

## Deployment Strategies

### Option 1: Render (Recommended for Production)

**Best for:** Production deployments, managed services, automatic SSL, zero DevOps

**Why Render?**
- Fully managed PostgreSQL and Redis
- Automatic SSL/TLS certificates
- Zero-downtime deployments
- Built-in monitoring and alerting
- Automatic scaling
- Free SSL certificates
- Easy GitHub integration

#### Prerequisites

1. **Render Account** - [render.com](https://render.com)
2. **Backblaze B2 Account** - [backblaze.com](https://www.backblaze.com/b2/cloud-storage.html)
3. **Auth0 Production App** configured
4. **Stripe Live API Keys** and webhook endpoint

#### Step 1: Set Up Backblaze B2 Storage

1. **Create B2 Account and Bucket:**
   ```bash
   # Log into Backblaze B2 dashboard
   # Create a new bucket: realstaging-prod
   # Set bucket to Private
   # Note the bucket region (e.g., us-west-004)
   ```

2. **Create Application Key:**
   - Go to "App Keys" in B2 dashboard
   - Create new key with access to your bucket
   - Save the `keyID` (access key) and `applicationKey` (secret key)
   - **Important:** Save these credentials securely - you won't see them again

3. **Configure B2 CORS (for presigned uploads):**
   ```json
   [
     {
       "corsRuleName": "allowUploads",
       "allowedOrigins": [
         "https://real-staging.ai"
       ],
       "allowedOperations": [
         "s3_put",
         "s3_get",
         "s3_head"
       ],
       "allowedHeaders": ["*"],
       "exposeHeaders": ["ETag"],
       "maxAgeSeconds": 3600
     }
   ]
   ```

4. **Note Your B2 Configuration:**
   - **Endpoint:** `https://s3.us-west-004.backblazeb2.com` (adjust region)
   - **Bucket Name:** `realstaging-prod`
   - **Region:** `us-west-004` (your bucket's region)
   - **Access Key ID:** Your B2 keyID
   - **Secret Access Key:** Your B2 applicationKey

#### Step 2: Create Render Blueprint

Create `render.yaml` in your repository root:

```yaml
services:
  # API Service
  - type: web
    name: realstaging-api
    runtime: docker
    dockerfilePath: ./apps/api/Dockerfile
    dockerContext: ./apps/api
    region: oregon  # or your preferred region
    plan: starter  # or standard for production
    numInstances: 1  # scale as needed
    healthCheckPath: /health
    envVars:
      - key: APP_ENV
        value: production
      - key: PORT
        value: 8080
      - key: DATABASE_URL
        fromDatabase:
          name: realstaging-db
          property: connectionString
      - key: REDIS_ADDR
        fromService:
          name: realstaging-redis
          type: redis
          property: connectionString
      - key: AUTH0_DOMAIN
        sync: false  # Set in Render dashboard
      - key: AUTH0_AUDIENCE
        sync: false
      - key: S3_ENDPOINT
        value: https://s3.us-west-004.backblazeb2.com
      - key: S3_REGION
        value: us-west-004
      - key: S3_BUCKET_NAME
        value: realstaging-prod
      - key: S3_ACCESS_KEY
        sync: false  # Set in Render dashboard (B2 keyID)
      - key: S3_SECRET_KEY
        sync: false  # Set in Render dashboard (B2 applicationKey)
      - key: S3_USE_PATH_STYLE
        value: false
      - key: STRIPE_SECRET_KEY
        sync: false
      - key: STRIPE_WEBHOOK_SECRET
        sync: false
      - key: REPLICATE_API_TOKEN
        sync: false
      - key: FRONTEND_URL
        value: https://app.yourdomain.com

  # Worker Service
  - type: worker
    name: realstaging-worker
    runtime: docker
    dockerfilePath: ./apps/worker/Dockerfile
    dockerContext: ./apps/worker
    region: oregon
    plan: starter
    numInstances: 1
    envVars:
      - key: APP_ENV
        value: production
      - key: DATABASE_URL
        fromDatabase:
          name: realstaging-db
          property: connectionString
      - key: REDIS_ADDR
        fromService:
          name: realstaging-redis
          type: redis
          property: connectionString
      - key: S3_ENDPOINT
        value: https://s3.us-west-004.backblazeb2.com
      - key: S3_REGION
        value: us-west-004
      - key: S3_BUCKET_NAME
        value: realstaging-prod
      - key: S3_ACCESS_KEY
        sync: false
      - key: S3_SECRET_KEY
        sync: false
      - key: S3_USE_PATH_STYLE
        value: false
      - key: REPLICATE_API_TOKEN
        sync: false
      - key: WORKER_CONCURRENCY
        value: 5

databases:
  - name: realstaging-db
    plan: starter  # or standard for production
    region: oregon
    databaseName: realstaging
    user: realstaging

  - name: realstaging-redis
    plan: starter
    region: oregon
    maxmemoryPolicy: allkeys-lru
```

#### Step 3: Deploy to Render

1. **Push to GitHub:**
   ```bash
   git add render.yaml
   git commit -m "feat(infra): add Render deployment configuration"
   git push origin main
   ```

2. **Connect in Render Dashboard:**
   - Go to [render.com/dashboard](https://render.com/dashboard)
   - Click "New" → "Blueprint"
   - Connect your GitHub repository
   - Render will detect `render.yaml` automatically
   - Click "Apply"

3. **Configure Environment Variables:**
   In Render dashboard, for each service, add the secret values:
   
   **API Service:**
   - `AUTH0_DOMAIN`: `your-tenant.us.auth0.com`
   - `AUTH0_AUDIENCE`: `https://api.yourdomain.com`
   - `S3_ACCESS_KEY`: Your B2 keyID
   - `S3_SECRET_KEY`: Your B2 applicationKey
   - `STRIPE_SECRET_KEY`: `sk_live_...`
   - `STRIPE_WEBHOOK_SECRET`: `whsec_...`
   - `REPLICATE_API_TOKEN`: `r8_...`

   **Worker Service:**
   - `S3_ACCESS_KEY`: Your B2 keyID
   - `S3_SECRET_KEY`: Your B2 applicationKey
   - `REPLICATE_API_TOKEN`: `r8_...`

4. **Run Database Migrations:**
   ```bash
   # Get shell access to API service
   # In Render dashboard: realstaging-api → Shell
   
   # Run migrations
   /app/migrate up
   
   # Verify migration status
   /app/migrate version
   ```

5. **Verify Deployment:**
   ```bash
   # Check health endpoint
   curl https://api.real-staging.ai/health
   
   # Should return:
   # {"status":"healthy","database":"connected","redis":"connected"}
   ```

#### Step 4: Configure Custom Domain (Optional)

1. **In Render Dashboard:**
   - Go to realstaging-api settings
   - Click "Custom Domains"
   - Add `api.yourdomain.com`
   - Add DNS records as shown

2. **Update Environment Variables:**
   - Update `FRONTEND_URL` if needed
   - Update Auth0 callback URLs
   - Update Stripe webhook URLs

#### Step 5: Set Up Monitoring

1. **Enable Render Metrics:**
   - Available in Render dashboard
   - CPU, memory, request rate
   - Response time percentiles

2. **Set Up Alerts:**
   - Go to service settings → Alerts
   - Configure alerts for:
     - High error rate
     - High response time
     - Service down

#### Scaling on Render

**Vertical Scaling:**
```yaml
# Update render.yaml
services:
  - type: web
    plan: standard  # or pro
```

**Horizontal Scaling:**
```yaml
# Update render.yaml
services:
  - type: web
    numInstances: 3  # scale API
  - type: worker
    numInstances: 2  # scale workers
```

**Database Scaling:**
- Upgrade plan in Render dashboard
- Enable connection pooling
- Consider read replicas for high load

#### Cost Estimation (Render + Backblaze B2)

**Render (Monthly):**
- API (Starter): $7/month
- Worker (Starter): $7/month
- PostgreSQL (Starter): $7/month
- Redis (Starter): $10/month
- **Total: ~$31/month**

**Backblaze B2 (Monthly):**
- Storage: $0.005/GB (~$5 for 1TB)
- Downloads: First 3x storage free, then $0.01/GB
- API Calls: Generous free tier
- **Estimated: $5-20/month** (depending on usage)

**Replicate AI:**
- Pay per image processed (~$0.011/image)

**Total estimated: $40-60/month** plus AI costs

---

### Option 2: Docker Compose (Small to Medium Scale)

**Best for:** Single-server deployments, small teams, 100-1000 users

#### Production Docker Compose Setup

Create `docker-compose.prod.yml`:

```yaml
version: '3.9'

services:
  api:
    image: realstaging/api:latest
    restart: always
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_ADDR=${REDIS_HOST}:${REDIS_PORT}
      - AUTH0_DOMAIN=${AUTH0_DOMAIN}
      - AUTH0_AUDIENCE=${AUTH0_AUDIENCE}
      - S3_BUCKET_NAME=${S3_BUCKET_NAME}
      - S3_REGION=${S3_REGION}
      - S3_ACCESS_KEY=${S3_ACCESS_KEY}
      - S3_SECRET_KEY=${S3_SECRET_KEY}
      - STRIPE_SECRET_KEY=${STRIPE_SECRET_KEY}
      - STRIPE_WEBHOOK_SECRET=${STRIPE_WEBHOOK_SECRET}
      - REPLICATE_API_TOKEN=${REPLICATE_API_TOKEN}
      - FRONTEND_URL=${FRONTEND_URL}
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 40s
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  worker:
    image: realstaging/worker:latest
    restart: always
    environment:
      - APP_ENV=production
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_ADDR=${REDIS_HOST}:${REDIS_PORT}
      - S3_BUCKET_NAME=${S3_BUCKET_NAME}
      - S3_REGION=${S3_REGION}
      - S3_ACCESS_KEY=${S3_ACCESS_KEY}
      - S3_SECRET_KEY=${S3_SECRET_KEY}
      - REPLICATE_API_TOKEN=${REPLICATE_API_TOKEN}
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8081/health"]
      interval: 30s
      timeout: 5s
      retries: 3
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  # Optional: Nginx reverse proxy with SSL
  nginx:
    image: nginx:alpine
    restart: always
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - api
```

#### Deploy

1. **Set up server:**
   ```bash
   # Install Docker
   curl -fsSL https://get.docker.com | sh
   sudo usermod -aG docker $USER
   
   # Install Docker Compose
   sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
   sudo chmod +x /usr/local/bin/docker-compose
   ```

2. **Prepare environment:**
   ```bash
   # Copy production compose file
   scp docker-compose.prod.yml user@server:/opt/realstaging/
   
   # Create .env file with secrets (never commit!)
   ssh user@server
   cd /opt/realstaging
   nano .env  # Add all environment variables
   chmod 600 .env
   ```

3. **Run migrations:**
   ```bash
   # SSH to server
   docker compose -f docker-compose.prod.yml run --rm api /app/migrate up
   ```

4. **Start services:**
   ```bash
   docker compose -f docker-compose.prod.yml up -d
   ```

5. **Verify:**
   ```bash
   docker compose -f docker-compose.prod.yml ps
   docker compose -f docker-compose.prod.yml logs -f api
   curl http://localhost:8080/health
   ```

#### Scaling

```bash
# Scale workers horizontally
docker compose -f docker-compose.prod.yml up -d --scale worker=3

# Scale API (behind load balancer)
docker compose -f docker-compose.prod.yml up -d --scale api=2
```

### Option 2: Kubernetes (Large Scale)

**Best for:** High availability, auto-scaling, 1000+ users

#### Namespace Setup

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: realstaging
```

#### Secrets

```yaml
# secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: realstaging-secrets
  namespace: realstaging
type: Opaque
stringData:
  DATABASE_URL: "postgres://user:pass@host:5432/dbname?sslmode=require"
  REDIS_ADDR: "redis.example.com:6379"
  AUTH0_DOMAIN: "your-tenant.us.auth0.com"
  AUTH0_AUDIENCE: "https://api.realstaging.ai"
  S3_ACCESS_KEY: "AKIAIOSFODNN7EXAMPLE"
  S3_SECRET_KEY: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  STRIPE_SECRET_KEY: "sk_live_..."
  STRIPE_WEBHOOK_SECRET: "whsec_..."
  REPLICATE_API_TOKEN: "r8_..."
```

Apply secrets:
```bash
kubectl apply -f secrets.yaml
```

#### ConfigMap

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: realstaging-config
  namespace: realstaging
data:
  APP_ENV: "production"
  S3_BUCKET_NAME: "realstaging-prod"
  S3_REGION: "us-east-1"
  FRONTEND_URL: "https://app.realstaging.ai"
```

#### API Deployment

```yaml
# api-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: realstaging
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api
  template:
    metadata:
      labels:
        app: api
    spec:
      containers:
      - name: api
        image: realstaging/api:1.0.0
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: realstaging-config
        - secretRef:
            name: realstaging-secrets
        resources:
          requests:
            cpu: "500m"
            memory: "512Mi"
          limits:
            cpu: "1000m"
            memory: "1Gi"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: api
  namespace: realstaging
spec:
  selector:
    app: api
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
```

#### Worker Deployment

```yaml
# worker-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: worker
  namespace: realstaging
spec:
  replicas: 2
  selector:
    matchLabels:
      app: worker
  template:
    metadata:
      labels:
        app: worker
    spec:
      containers:
      - name: worker
        image: realstaging/worker:1.0.0
        envFrom:
        - configMapRef:
            name: realstaging-config
        - secretRef:
            name: realstaging-secrets
        resources:
          requests:
            cpu: "1000m"
            memory: "1Gi"
          limits:
            cpu: "2000m"
            memory: "2Gi"
        livenessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 30
          periodSeconds: 10
```

#### Ingress (NGINX)

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: api-ingress
  namespace: realstaging
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - api.realstaging.ai
    secretName: api-tls
  rules:
  - host: api.realstaging.ai
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api
            port:
              number: 80
```

#### Horizontal Pod Autoscaler

```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-hpa
  namespace: realstaging
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: worker-hpa
  namespace: realstaging
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: worker
  minReplicas: 2
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 80
```

#### Deploy to Kubernetes

```bash
# Create namespace
kubectl apply -f namespace.yaml

# Apply secrets and config
kubectl apply -f secrets.yaml
kubectl apply -f configmap.yaml

# Run migrations (one-time job)
kubectl run migrate --image=realstaging/api:1.0.0 \
  --namespace=realstaging \
  --restart=Never \
  --env-from=configmap/realstaging-config \
  --env-from=secret/realstaging-secrets \
  --command -- /app/migrate up

# Wait for migration to complete
kubectl wait --for=condition=complete job/migrate -n realstaging --timeout=300s

# Deploy services
kubectl apply -f api-deployment.yaml
kubectl apply -f worker-deployment.yaml
kubectl apply -f ingress.yaml
kubectl apply -f hpa.yaml

# Verify
kubectl get pods -n realstaging
kubectl get svc -n realstaging
kubectl get ing -n realstaging

# Check logs
kubectl logs -f -n realstaging -l app=api
```

### Option 3: Fly.io (Alternative)

**Best for:** Global low-latency, easy scaling, Heroku alternative

#### Install Fly CLI

```bash
curl -L https://fly.io/install.sh | sh
fly auth login
```

#### Create Fly Apps

```bash
# Create API app
fly apps create realstaging-api --org your-org

# Create Worker app
fly apps create realstaging-worker --org your-org
```

#### Configure fly.toml for API

```toml
# apps/api/fly.toml
app = "realstaging-api"
primary_region = "iad"

[build]
  dockerfile = "Dockerfile"

[env]
  APP_ENV = "production"
  PORT = "8080"

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = false
  auto_start_machines = true
  min_machines_running = 2

[[http_service.checks]]
  interval = "30s"
  timeout = "5s"
  grace_period = "10s"
  method = "GET"
  path = "/health"

[processes]
  app = "./api-server"

[[services]]
  protocol = "tcp"
  internal_port = 8080

  [[services.ports]]
    port = 80
    handlers = ["http"]

  [[services.ports]]
    port = 443
    handlers = ["tls", "http"]
```

#### Configure fly.toml for Worker

```toml
# apps/worker/fly.toml
app = "realstaging-worker"
primary_region = "iad"

[build]
  dockerfile = "Dockerfile"

[env]
  APP_ENV = "production"

[processes]
  app = "./worker"

[deploy]
  max_unavailable = 0.5
```

#### Set Secrets

```bash
# API secrets
fly secrets set -a realstaging-api \
  DATABASE_URL="postgres://..." \
  REDIS_ADDR="redis://..." \
  AUTH0_DOMAIN="..." \
  AUTH0_AUDIENCE="..." \
  S3_BUCKET_NAME="..." \
  S3_ACCESS_KEY="..." \
  S3_SECRET_KEY="..." \
  STRIPE_SECRET_KEY="..." \
  STRIPE_WEBHOOK_SECRET="..." \
  REPLICATE_API_TOKEN="..."

# Worker secrets
fly secrets set -a realstaging-worker \
  DATABASE_URL="postgres://..." \
  REDIS_ADDR="redis://..." \
  S3_BUCKET_NAME="..." \
  S3_ACCESS_KEY="..." \
  S3_SECRET_KEY="..." \
  REPLICATE_API_TOKEN="..."
```

#### Provision Postgres

```bash
# Create Fly Postgres cluster
fly postgres create --name realstaging-db --region iad

# Attach to API
fly postgres attach realstaging-db -a realstaging-api

# Attach to Worker
fly postgres attach realstaging-db -a realstaging-worker
```

#### Provision Redis

```bash
# Create Upstash Redis
fly redis create --org your-org --name realstaging-redis --region iad

# Get connection string
fly redis status realstaging-redis
```

#### Deploy

```bash
# Deploy API
cd apps/api
fly deploy

# Deploy Worker
cd apps/worker
fly deploy

# Scale
fly scale count 3 -a realstaging-api
fly scale count 2 -a realstaging-worker

# Check status
fly status -a realstaging-api
fly logs -a realstaging-api
```

### Option 4: Fly.io (Alternative Platform)

**Best for:** Global low-latency, easy scaling, edge deployments

See full Fly.io deployment instructions earlier in this document for detailed setup

## Secrets Management

### Best Practices

1. **Never commit secrets to version control**
2. **Use environment variables** for all sensitive data
3. **Rotate secrets regularly** (quarterly recommended)
4. **Use separate secrets** for each environment
5. **Limit access** - only admins should see production secrets

### Secrets Management Solutions

#### AWS Secrets Manager

```bash
# Store secret
aws secretsmanager create-secret \
  --name realstaging/prod/database-url \
  --secret-string "postgres://user:pass@host/db"

# Retrieve in application
aws secretsmanager get-secret-value \
  --secret-id realstaging/prod/database-url \
  --query SecretString --output text
```

#### HashiCorp Vault

```bash
# Store secret
vault kv put secret/realstaging/prod \
  database_url="postgres://..." \
  stripe_key="sk_live_..."

# Retrieve secret
vault kv get -field=database_url secret/realstaging/prod
```

#### Kubernetes Secrets (from external source)

Use [External Secrets Operator](https://external-secrets.io/) to sync from AWS/Vault:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secrets
  namespace: realstaging
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-east-1

---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: realstaging-secrets
  namespace: realstaging
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets
    kind: SecretStore
  target:
    name: realstaging-secrets
  data:
  - secretKey: DATABASE_URL
    remoteRef:
      key: realstaging/prod/database-url
```

## Production Configuration

### Required Environment Variables

For a production environment, configure:

| Variable | Description | Example |
|----------|-------------|---------|
| `APP_ENV` | Environment name | `production` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@host:5432/db?sslmode=require` |
| `REDIS_ADDR` | Redis connection address | `redis.example.com:6379` |
| `AUTH0_DOMAIN` | Auth0 tenant domain | `your-tenant.us.auth0.com` |
| `AUTH0_AUDIENCE` | Auth0 API audience | `https://api.realstaging.ai` |
| `S3_BUCKET_NAME` | S3 bucket for images | `realstaging-prod` |
| `S3_REGION` | S3 region | `us-east-1` |
| `S3_ACCESS_KEY` | S3 access key | `AKIAIOSFODNN7EXAMPLE` |
| `S3_SECRET_KEY` | S3 secret key | `wJalrXUtnFEMI/K7MDENG/...` |
| `STRIPE_SECRET_KEY` | Stripe API key (live) | `sk_live_...` |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook secret | `whsec_...` |
| `REPLICATE_API_TOKEN` | Replicate API token | `r8_...` |
| `FRONTEND_URL` | Frontend base URL | `https://app.realstaging.ai` |

### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP port | `8080` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OpenTelemetry endpoint | (disabled) |
| `LOG_LEVEL` | Logging level | `info` |
| `MAX_WORKERS` | Worker concurrency | `4` |

For detailed explanations, see [Configuration Guide](../guides/configuration.md).

## Health Checks

### API Health Endpoint

**GET /health**

Returns service status and dependencies:

```json
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": "2025-10-17T18:30:00Z",
  "checks": {
    "database": "ok",
    "redis": "ok",
    "s3": "ok"
  }
}
```

### Health Check Implementations

#### Docker Compose

```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
  interval: 30s
  timeout: 5s
  retries: 3
  start_period: 40s
```

#### Kubernetes

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 2
```

### Monitoring Health

```bash
# Check health manually
curl https://api.realstaging.ai/health

# Monitor with watch
watch -n 5 curl -s https://api.realstaging.ai/health | jq

# Check from Kubernetes
kubectl get pods -n realstaging
kubectl exec -it -n realstaging deploy/api -- curl localhost:8080/health
```

## Rollback Procedures

### Docker Compose Rollback

```bash
# Tag current version before deploying
docker tag realstaging/api:latest realstaging/api:backup-$(date +%Y%m%d)

# If deployment fails, rollback
docker compose -f docker-compose.prod.yml down
docker tag realstaging/api:backup-20251017 realstaging/api:latest
docker compose -f docker-compose.prod.yml up -d

# Rollback database migration (if needed)
docker compose -f docker-compose.prod.yml run --rm api /app/migrate down 1
```

### Kubernetes Rollback

```bash
# Check rollout status
kubectl rollout status deployment/api -n realstaging

# View rollout history
kubectl rollout history deployment/api -n realstaging

# Rollback to previous version
kubectl rollout undo deployment/api -n realstaging

# Rollback to specific revision
kubectl rollout undo deployment/api -n realstaging --to-revision=3

# Pause rollout if issues detected
kubectl rollout pause deployment/api -n realstaging

# Resume after fix
kubectl rollout resume deployment/api -n realstaging
```

### Database Migration Rollback

```bash
# Check current migration version
psql $DATABASE_URL -c "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;"

# Rollback one migration
./migrate -database $DATABASE_URL -path ./migrations down 1

# Rollback to specific version
./migrate -database $DATABASE_URL -path ./migrations goto 5

# Verify migration status
./migrate -database $DATABASE_URL -path ./migrations version
```

### Rollback Checklist

When rolling back:

- [ ] Identify the issue (check logs, metrics, alerts)
- [ ] Determine the last known good version
- [ ] Check if database migrations were run
- [ ] Announce maintenance window to users
- [ ] Rollback application code
- [ ] Rollback database migrations (if needed)
- [ ] Verify health checks pass
- [ ] Test critical user flows
- [ ] Monitor for errors
- [ ] Document incident and root cause

## Troubleshooting Deployments

### Common Issues

#### 1. Container Won't Start

**Symptoms:**
- Pod in CrashLoopBackOff
- Container exits immediately

**Diagnosis:**
```bash
# Check logs
docker logs <container-id>
kubectl logs -n realstaging deploy/api

# Check events
kubectl describe pod -n realstaging <pod-name>
```

**Common causes:**
- Missing environment variables
- Database connection failure
- Port already in use
- Insufficient resources

#### 2. Health Checks Failing

**Symptoms:**
- Pods marked as unhealthy
- Traffic not routed to pods

**Diagnosis:**
```bash
# Test health endpoint directly
kubectl exec -it -n realstaging <pod-name> -- curl localhost:8080/health

# Check resource usage
kubectl top pods -n realstaging
```

**Common causes:**
- Slow startup (adjust `initialDelaySeconds`)
- Database connection pool exhausted
- Memory/CPU limits too low

#### 3. Database Connection Errors

**Symptoms:**
- "connection refused" errors
- "too many connections" errors

**Diagnosis:**
```bash
# Test connection from pod
kubectl exec -it -n realstaging <pod-name> -- psql $DATABASE_URL -c "SELECT 1;"

# Check connection count
psql $DATABASE_URL -c "SELECT count(*) FROM pg_stat_activity;"
```

**Solutions:**
- Verify `DATABASE_URL` format
- Check database firewall rules
- Increase connection pool size
- Scale down replicas temporarily

#### 4. S3 Upload Failures

**Symptoms:**
- "AccessDenied" errors
- "NoSuchBucket" errors

**Diagnosis:**
```bash
# Test S3 access from pod
kubectl exec -it -n realstaging <pod-name> -- aws s3 ls s3://$S3_BUCKET_NAME
```

**Solutions:**
- Verify S3 credentials
- Check bucket policy and CORS
- Verify region matches bucket location

### Getting Help

If issues persist:

1. Check the **Troubleshooting** section of the [Guides](../guides/index.md) for detailed scenarios
2. Review [Monitoring](monitoring.md) for observability setup
3. Check application logs for error details
4. Verify all configuration matches [Configuration Guide](../guides/configuration.md)

## Post-Deployment Checklist

After deploying to production:

- [ ] All services healthy and passing health checks
- [ ] Database migrations applied successfully
- [ ] SSL certificates valid and auto-renewing
- [ ] Monitoring and alerting configured
- [ ] Log aggregation working
- [ ] Backups scheduled and tested
- [ ] DNS records pointing to correct endpoints
- [ ] Auth0 callbacks configured for production domain
- [ ] Stripe webhooks configured for production
- [ ] Test user signup and login flow
- [ ] Test image upload and processing
- [ ] Test subscription checkout flow
- [ ] Verify images stored in S3
- [ ] Check error rates in monitoring
- [ ] Document deployment date and version
- [ ] Update runbook with any issues encountered

## Related Documentation

- [Configuration Guide](../guides/configuration.md) - All environment variables
- [Monitoring Guide](monitoring.md) - Observability and alerting
- [Troubleshooting Guide](deployment.md#troubleshooting-deployments) - Common issues and fixes
- [Stripe Billing Guide](../guides/stripe-billing.md) - Payment configuration
