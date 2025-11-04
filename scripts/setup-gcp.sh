#!/bin/bash
# Google Cloud Platform Setup Script for Barracuda

set -e

echo "🚀 Barracuda - Google Cloud Platform Setup"
echo "==========================================="
echo ""

# Check if gcloud is available
if ! command -v gcloud &> /dev/null; then
    echo "❌ gcloud CLI not found. Make sure it's installed and in your PATH."
    echo "   Try: source ~/.zshrc"
    exit 1
fi

echo "✓ gcloud CLI found: $(gcloud --version | head -1)"
echo ""

# Step 1: Login
echo "Step 1: Authenticate with Google Cloud"
echo "---------------------------------------"
echo "This will open a browser for authentication..."
gcloud auth login

echo ""
echo "✓ Authentication complete!"
echo ""

# Step 2: List/create project
echo "Step 2: Select or Create a Google Cloud Project"
echo "------------------------------------------------"
echo ""
echo "Your existing projects:"
gcloud projects list --format="table(projectId,name)" || echo "No projects found"
echo ""

read -p "Enter your project ID (or press Enter to create a new one): " PROJECT_ID

if [ -z "$PROJECT_ID" ]; then
    read -p "Enter name for new project: " PROJECT_NAME
    PROJECT_ID="${PROJECT_NAME}-$(date +%s | tail -c 5)"
    echo "Creating project: $PROJECT_ID"
    gcloud projects create "$PROJECT_ID" --name="$PROJECT_NAME" || {
        echo "Project creation failed. You may need to create it manually at:"
        echo "https://console.cloud.google.com/projectcreate"
        exit 1
    }
fi

gcloud config set project "$PROJECT_ID"
echo "✓ Project set to: $PROJECT_ID"
echo ""

# Step 3: Enable billing (check)
echo "Step 3: Billing Setup"
echo "---------------------"
echo "⚠️  Cloud Run requires billing to be enabled (even with free tier)."
echo ""
read -p "Have you enabled billing for this project? (y/n): " BILLING_ENABLED

if [ "$BILLING_ENABLED" != "y" ] && [ "$BILLING_ENABLED" != "Y" ]; then
    echo ""
    echo "Please enable billing:"
    echo "1. Go to: https://console.cloud.google.com/billing"
    echo "2. Link a billing account to project: $PROJECT_ID"
    echo "3. Run this script again"
    exit 1
fi

echo "✓ Billing check complete"
echo ""

# Step 4: Enable APIs
echo "Step 4: Enabling Required APIs"
echo "-------------------------------"
echo "This may take a minute..."

gcloud services enable cloudbuild.googleapis.com
gcloud services enable run.googleapis.com
gcloud services enable artifactregistry.googleapis.com
gcloud services enable secretmanager.googleapis.com

echo "✓ APIs enabled"
echo ""

# Step 5: Set region
echo "Step 5: Select Region"
echo "---------------------"
echo "Available regions:"
echo "  - us-central1 (Iowa, USA)"
echo "  - us-east1 (South Carolina, USA)"
echo "  - europe-west1 (Belgium)"
echo "  - asia-southeast1 (Singapore)"
echo ""
read -p "Enter region (default: us-central1): " REGION
REGION=${REGION:-us-central1}

echo "✓ Region set to: $REGION"
echo ""

# Step 6: Create Artifact Registry
echo "Step 6: Creating Artifact Registry Repository"
echo "-----------------------------------------------"
gcloud artifacts repositories create barracuda \
    --repository-format=docker \
    --location="$REGION" \
    --description="Barracuda API Docker images" || {
    echo "⚠️  Repository might already exist, continuing..."
}

echo "✓ Artifact Registry ready"
echo ""

# Step 7: Configure Docker auth
echo "Step 7: Configuring Docker Authentication"
echo "------------------------------------------"
gcloud auth configure-docker "${REGION}-docker.pkg.dev" --quiet
echo "✓ Docker authentication configured"
echo ""

# Step 8: Create secret (if Supabase key provided)
echo "Step 8: Secret Manager Setup"
echo "------------------------------"
read -p "Do you want to store your Supabase service role key now? (y/n): " STORE_SECRET

if [ "$STORE_SECRET" = "y" ] || [ "$STORE_SECRET" = "Y" ]; then
    read -sp "Enter your Supabase service role key: " SERVICE_KEY
    echo ""
    echo -n "$SERVICE_KEY" | gcloud secrets create supabase-service-role-key \
        --data-file=- \
        --replication-policy="automatic" || {
        echo "⚠️  Secret might already exist. Update it with:"
        echo "   echo -n 'your-key' | gcloud secrets versions add supabase-service-role-key --data-file=-"
    }
    echo "✓ Secret stored"
else
    echo "You can store it later with:"
    echo "  echo -n 'your-key' | gcloud secrets create supabase-service-role-key --data-file=-"
fi

echo ""

# Summary
echo "========================================="
echo "✅ Setup Complete!"
echo "========================================="
echo ""
echo "Project ID: $PROJECT_ID"
echo "Region: $REGION"
echo ""
echo "Next steps:"
echo "1. Make sure you have these environment variables set:"
echo "   export GCP_PROJECT_ID=$PROJECT_ID"
echo "   export GCP_REGION=$REGION"
echo "   export PUBLIC_SUPABASE_URL=https://your-project.supabase.co"
echo "   export PUBLIC_SUPABASE_ANON_KEY=your-anon-key"
echo ""
echo "2. Test the API locally:"
echo "   go run . api --port 8080"
echo ""
echo "3. Deploy to Cloud Run:"
echo "   make deploy"
echo ""
echo "Or see docs/GCP_SETUP.md for detailed instructions"
echo ""

