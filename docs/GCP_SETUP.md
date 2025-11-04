# Google Cloud Platform Setup Guide

This guide walks you through setting up Google Cloud Platform for deploying Barracuda to Cloud Run.

## Step 1: Install gcloud CLI

### Linux Installation

```bash
# Download and install gcloud CLI
curl https://sdk.cloud.google.com | bash

# Restart your shell or run:
exec -l $SHELL

# Initialize gcloud
gcloud init
```

Follow the prompts to:
1. Log in with your Google account
2. Select or create a project
3. Choose a default region

### Alternative: Install via Package Manager

**For Debian/Ubuntu:**
```bash
echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key --keyring /usr/share/keyrings/cloud.google.gpg add -
sudo apt-get update && sudo apt-get install google-cloud-cli
```

**For Fedora/RHEL:**
```bash
sudo tee -a /etc/yum.repos.d/google-cloud-sdk.repo << EOM
[google-cloud-cli]
name=Google Cloud CLI
baseurl=https://packages.cloud.google.com/yum/repos/cloud-sdk-el8-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=0
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg
       https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOM
sudo yum install google-cloud-cli
```

## Step 2: Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Click "Select a project" → "New Project"
3. Enter project name: `barracuda` (or your preferred name)
4. Note your **Project ID** (different from project name)
5. Click "Create"

Or via CLI:
```bash
gcloud projects create barracuda --name="Barracuda SEO Scanner"
gcloud config set project barracuda
```

## Step 3: Enable Billing

**Important:** Cloud Run requires billing to be enabled (even though there's a free tier).

1. Go to [Billing](https://console.cloud.google.com/billing)
2. Link a billing account to your project
3. You'll get $300 free credit and a generous free tier

## Step 4: Enable Required APIs

Run these commands to enable the APIs needed for Cloud Run:

```bash
# Set your project ID (replace with your actual project ID)
export GCP_PROJECT_ID=your-project-id
gcloud config set project $GCP_PROJECT_ID

# Enable required APIs
gcloud services enable cloudbuild.googleapis.com
gcloud services enable run.googleapis.com
gcloud services enable artifactregistry.googleapis.com
gcloud services enable secretmanager.googleapis.com
```

## Step 5: Set Up Authentication

```bash
# Login to Google Cloud
gcloud auth login

# Set default project
gcloud config set project $GCP_PROJECT_ID

# Verify
gcloud config list
```

## Step 6: Create Artifact Registry Repository

```bash
# Set your preferred region (us-central1, us-east1, europe-west1, etc.)
export GCP_REGION=us-central1

# Create repository for Docker images
gcloud artifacts repositories create barracuda \
  --repository-format=docker \
  --location=$GCP_REGION \
  --description="Barracuda API Docker images"
```

## Step 7: Store Secrets in Secret Manager

**Get your Supabase service role key** from Supabase dashboard → Settings → API → service_role key

```bash
# Store Supabase service role key (never commit this!)
echo -n "your-service-role-key-here" | gcloud secrets create supabase-service-role-key \
  --data-file=- \
  --replication-policy="automatic"

# Verify secret was created
gcloud secrets list
```

## Step 8: Configure Docker Authentication

```bash
# Authenticate Docker to push to Artifact Registry
gcloud auth configure-docker $GCP_REGION-docker.pkg.dev
```

## Step 9: Test Locally First

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

## Step 10: Deploy to Cloud Run

Once everything is set up, deploy using the Makefile:

```bash
# Set environment variables
export GCP_PROJECT_ID=your-project-id
export GCP_REGION=us-central1
export PUBLIC_SUPABASE_URL=https://your-project.supabase.co
export PUBLIC_SUPABASE_ANON_KEY=your-anon-key

# Deploy (builds, pushes, and deploys)
make deploy
```

Or deploy manually:

```bash
# Build and push Docker image
make docker-build
make docker-push

# Deploy to Cloud Run
gcloud run deploy barracuda-api \
  --image $GCP_REGION-docker.pkg.dev/$GCP_PROJECT_ID/barracuda/barracuda-api:latest \
  --platform managed \
  --region $GCP_REGION \
  --allow-unauthenticated \
  --set-env-vars="PUBLIC_SUPABASE_URL=$PUBLIC_SUPABASE_URL,PUBLIC_SUPABASE_ANON_KEY=$PUBLIC_SUPABASE_ANON_KEY" \
  --set-secrets="SUPABASE_SERVICE_ROLE_KEY=supabase-service-role-key:latest" \
  --memory=512Mi \
  --cpu=1 \
  --timeout=300 \
  --max-instances=10 \
  --port=8080
```

## Verify Deployment

```bash
# Get the service URL
gcloud run services describe barracuda-api \
  --platform managed \
  --region $GCP_REGION \
  --format="value(status.url)"

# Test the deployed API
curl $(gcloud run services describe barracuda-api --platform managed --region $GCP_REGION --format="value(status.url)")/health
```

## Troubleshooting

### Check if APIs are enabled
```bash
gcloud services list --enabled
```

### View Cloud Run logs
```bash
gcloud run services logs read barracuda-api \
  --platform managed \
  --region $GCP_REGION
```

### Check service status
```bash
gcloud run services describe barracuda-api \
  --platform managed \
  --region $GCP_REGION
```

### Common Issues

1. **"Permission denied"** → Run `gcloud auth login`
2. **"API not enabled"** → Enable APIs (Step 4)
3. **"Billing not enabled"** → Enable billing (Step 3)
4. **"Secret not found"** → Create secret (Step 7)
5. **"Image not found"** → Build and push image first

## Next Steps

After Cloud Run is deployed:
1. Note the service URL (you'll need it for Vercel)
2. Test all API endpoints
3. Set up Vercel deployment with the Cloud Run URL

