# Cloud Run Deployment Guide

This guide walks you through deploying the Barracuda API server to Google Cloud Run.

## Prerequisites

1. **Google Cloud Account** with billing enabled
2. **gcloud CLI** installed and authenticated
3. **Docker** installed (for local testing)
4. **Supabase project** with migrations applied

## Step 1: Test API Locally

Before deploying, test the API server locally:

```bash
# Set environment variables
export PUBLIC_SUPABASE_URL=https://your-project.supabase.co
export SUPABASE_SERVICE_ROLE_KEY=your-service-role-key
export PUBLIC_SUPABASE_ANON_KEY=your-anon-key

# Test the API server
go run . api --port 8080

# In another terminal, test health endpoint
curl http://localhost:8080/health
```

## Step 2: Set Up Google Cloud Project

```bash
# Set your project ID (replace with your actual project ID)
export GCP_PROJECT_ID=your-project-id
export GCP_REGION=us-central1  # Choose your preferred region

# Set the project
gcloud config set project $GCP_PROJECT_ID

# Enable required APIs
gcloud services enable cloudbuild.googleapis.com
gcloud services enable run.googleapis.com
gcloud services enable artifactregistry.googleapis.com
gcloud services enable secretmanager.googleapis.com
```

## Step 3: Create Artifact Registry Repository

```bash
# Create repository for Docker images
gcloud artifacts repositories create barracuda \
  --repository-format=docker \
  --location=$GCP_REGION \
  --description="Barracuda API Docker images"
```

## Step 4: Store Secrets in Secret Manager

```bash
# Store Supabase service role key (never commit this!)
echo -n "your-service-role-key" | gcloud secrets create supabase-service-role-key \
  --data-file=- \
  --replication-policy="automatic"

# If you have OpenAI API key (optional)
# echo -n "your-openai-key" | gcloud secrets create openai-api-key \
#   --data-file=- \
#   --replication-policy="automatic"
```

## Step 5: Build and Push Docker Image

```bash
# Authenticate Docker to push to Artifact Registry
gcloud auth configure-docker $GCP_REGION-docker.pkg.dev

# Build and push (or use make commands)
make docker-build
make docker-push
```

## Step 6: Deploy to Cloud Run

```bash
# Deploy with environment variables
gcloud run deploy barracuda-api \
  --image $GCP_REGION-docker.pkg.dev/$GCP_PROJECT_ID/barracuda/barracuda-api:latest \
  --platform managed \
  --region $GCP_REGION \
  --allow-unauthenticated \
  --set-env-vars="PUBLIC_SUPABASE_URL=https://your-project.supabase.co,PUBLIC_SUPABASE_ANON_KEY=your-anon-key" \
  --set-secrets="SUPABASE_SERVICE_ROLE_KEY=supabase-service-role-key:latest" \
  --memory=512Mi \
  --cpu=1 \
  --timeout=300 \
  --max-instances=10 \
  --port=8080

# Get the service URL
gcloud run services describe barracuda-api \
  --platform managed \
  --region $GCP_REGION \
  --format="value(status.url)"
```

## Step 7: Test Deployed API

```bash
# Get the service URL
export API_URL=$(gcloud run services describe barracuda-api \
  --platform managed \
  --region $GCP_REGION \
  --format="value(status.url)")

# Test health endpoint
curl $API_URL/health
```

## Environment Variables Reference

- `PUBLIC_SUPABASE_URL` - Your Supabase project URL
- `PUBLIC_SUPABASE_ANON_KEY` - Supabase anon key (safe to expose)
- `SUPABASE_SERVICE_ROLE_KEY` - Service role key (from Secret Manager)
- `PORT` - Automatically set by Cloud Run (default: 8080)

## Updating the Deployment

After making code changes:

```bash
# Rebuild and push
make docker-build
make docker-push

# Redeploy (uses latest image)
gcloud run deploy barracuda-api \
  --image $GCP_REGION-docker.pkg.dev/$GCP_PROJECT_ID/barracuda/barracuda-api:latest \
  --platform managed \
  --region $GCP_REGION
```

## Troubleshooting

### View Logs
```bash
gcloud run services logs read barracuda-api \
  --platform managed \
  --region $GCP_REGION
```

### Check Service Status
```bash
gcloud run services describe barracuda-api \
  --platform managed \
  --region $GCP_REGION
```

### Common Issues

1. **"Permission denied"** - Make sure you're authenticated: `gcloud auth login`
2. **"Image not found"** - Ensure you've pushed the image to Artifact Registry
3. **"Secret not found"** - Verify secret exists: `gcloud secrets list`
4. **"API not enabled"** - Enable required APIs (see Step 2)

## Cost Considerations

Cloud Run pricing:
- **Free tier**: 2 million requests/month, 360,000 GB-seconds memory, 180,000 vCPU-seconds
- **After free tier**: Pay per use (requests, memory, CPU time)

For a small to medium application, Cloud Run is very cost-effective.

