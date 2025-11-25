# Deployment Guide

How to deploy the M-Pesa MCP Server in various environments.

## Local Development

```bash
# 1. Setup environment
cp .env.example .env
# Edit .env with your Daraja sandbox credentials

# 2. Run
go run cmd/server/main.go
```

## Production Deployment

### Option 1: Docker

```dockerfile
# Dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o mpesa-mcp cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/mpesa-mcp .
EXPOSE 8080
CMD ["./mpesa-mcp"]
```

```bash
# Build and run
docker build -t mpesa-mcp .
docker run -p 8080:8080 \
  -e MPESA_CONSUMER_KEY=your_key \
  -e MPESA_CONSUMER_SECRET=your_secret \
  -e BASE_URL=https://api.safaricom.co.ke \
  -e BUSINESS_SHORTCODE=your_shortcode \
  -e PASSKEY=your_passkey \
  -e CALLBACK_URL=https://your-domain.com/callback \
  mpesa-mcp
```

### Option 2: Systemd Service (Linux)

```ini
# /etc/systemd/system/mpesa-mcp.service
[Unit]
Description=M-Pesa MCP Server
After=network.target

[Service]
Type=simple
User=mpesa
WorkingDirectory=/opt/mpesa-mcp
EnvironmentFile=/opt/mpesa-mcp/.env
ExecStart=/opt/mpesa-mcp/mpesa-mcp
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start
sudo systemctl enable mpesa-mcp
sudo systemctl start mpesa-mcp
sudo systemctl status mpesa-mcp
```

### Option 3: Cloud Platforms

#### Google Cloud Run

```bash
# Build and deploy
gcloud builds submit --tag gcr.io/PROJECT_ID/mpesa-mcp
gcloud run deploy mpesa-mcp \
  --image gcr.io/PROJECT_ID/mpesa-mcp \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars MPESA_CONSUMER_KEY=xxx,MPESA_CONSUMER_SECRET=xxx
```

#### AWS ECS/Fargate

```bash
# Push to ECR
aws ecr create-repository --repository-name mpesa-mcp
docker tag mpesa-mcp:latest AWS_ACCOUNT.dkr.ecr.REGION.amazonaws.com/mpesa-mcp:latest
docker push AWS_ACCOUNT.dkr.ecr.REGION.amazonaws.com/mpesa-mcp:latest

# Deploy via ECS console or CLI
```

#### Heroku

```bash
# Create Procfile
echo "web: ./mpesa-mcp" > Procfile

# Deploy
heroku create mpesa-mcp
heroku config:set MPESA_CONSUMER_KEY=xxx MPESA_CONSUMER_SECRET=xxx
git push heroku main
```

### Option 4: Kubernetes

```yaml
# k8s-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mpesa-mcp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: mpesa-mcp
  template:
    metadata:
      labels:
        app: mpesa-mcp
    spec:
      containers:
      - name: mpesa-mcp
        image: your-registry/mpesa-mcp:latest
        ports:
        - containerPort: 8080
        env:
        - name: MPESA_CONSUMER_KEY
          valueFrom:
            secretKeyRef:
              name: mpesa-secrets
              key: consumer-key
        - name: MPESA_CONSUMER_SECRET
          valueFrom:
            secretKeyRef:
              name: mpesa-secrets
              key: consumer-secret
---
apiVersion: v1
kind: Service
metadata:
  name: mpesa-mcp
spec:
  selector:
    app: mpesa-mcp
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

## Production Checklist

- [ ] Use production Daraja credentials (not sandbox)
- [ ] Set up HTTPS with valid SSL certificate
- [ ] Implement callback endpoint to receive payment confirmations
- [ ] Add request logging and monitoring
- [ ] Set up health checks and alerts
- [ ] Configure firewall rules
- [ ] Implement rate limiting
- [ ] Add authentication for MCP endpoints (if needed)
- [ ] Set up backup and disaster recovery
- [ ] Document runbooks for common issues
