# Local Development Setup Guide

This guide walks you through setting up the Barracuda CLI scanner for local development and testing on a new machine.

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Node.js 18+** - [Install Node.js](https://nodejs.org/)
- **npm** - Comes with Node.js
- **Supabase CLI** (optional, for local Supabase) - [Install Supabase CLI](https://supabase.com/docs/guides/cli)
- **Git** - For cloning the repository

## Step 1: Clone and Initial Setup

```bash
# Clone the repository
git clone https://github.com/dillonlara115/barracuda.git
cd barracuda

# Install frontend dependencies
cd web
npm install
cd ..
```

## Step 2: Supabase Setup

You have two options for Supabase: **Remote (Cloud)** or **Local (Docker)**.

### Option A: Using Remote Supabase (Recommended for Quick Start)

1. **Create a Supabase Project** (if you don't have one):
   - Go to [https://supabase.com](https://supabase.com)
   - Sign up or log in
   - Click "New Project"
   - Choose your organization, name your project, set a database password
   - Wait for the project to be created (~2 minutes)

2. **Get Your Supabase Credentials**:
   - In your Supabase project dashboard, go to **Settings** → **API**
   - Copy the following values:
     - **Project URL** (e.g., `https://xxxxx.supabase.co`)
     - **anon/public key** (starts with `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`)
     - **service_role key** (starts with `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`) - Keep this secret!

3. **Configure Supabase Auth Redirects**:
   - Go to **Settings** → **Authentication** → **URL Configuration**
   - Set **Site URL** to: `http://localhost:5173`
   - Add **Redirect URLs**:
     ```
     http://localhost:5173
     http://localhost:5173/**
     http://localhost:8080
     http://localhost:8080/**
     ```

4. **Run Database Migrations**:
   ```bash
   # Install Supabase CLI if you haven't
   # macOS: brew install supabase/tap/supabase
   # Or download from: https://github.com/supabase/cli/releases
   
   # Link to your remote project
   supabase link --project-ref YOUR_PROJECT_REF
   # You can find your project ref in the Supabase dashboard URL or Settings → General
   
   # Push migrations to your remote database
   supabase db push
   ```

### Option B: Using Local Supabase (Docker Required)

1. **Start Local Supabase**:
   ```bash
   # Make sure Docker is running
   supabase start
   ```

2. **Get Local Credentials**:
   ```bash
   supabase status
   ```
   This will show you:
   - API URL: `http://127.0.0.1:54321`
   - anon key: `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`
   - service_role key: `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`

3. **Apply Migrations Locally**:
   ```bash
   # Migrations are automatically applied when you run `supabase start`
   # But you can reset and reapply if needed:
   supabase db reset
   ```

## Step 3: Environment Variables

Create environment variable files for local development.

### Root Directory `.env.local`

Create `.env.local` in the project root:

```bash
# Supabase Configuration
PUBLIC_SUPABASE_URL=https://your-project.supabase.co
PUBLIC_SUPABASE_ANON_KEY=your-anon-key-here
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key-here

# Cloud Run API URL (optional, for production API)
# VITE_CLOUD_RUN_API_URL=https://barracuda-api-your-env.a.run.app
```

**For Local Supabase**, use:
```bash
PUBLIC_SUPABASE_URL=http://127.0.0.1:54321
PUBLIC_SUPABASE_ANON_KEY=<from supabase status>
SUPABASE_SERVICE_ROLE_KEY=<from supabase status>
```

### Web Directory `.env.local`

Create `web/.env.local`:

```bash
# These should match the root .env.local
PUBLIC_SUPABASE_URL=https://your-project.supabase.co
PUBLIC_SUPABASE_ANON_KEY=your-anon-key-here

# Optional: For connecting to remote API
# VITE_CLOUD_RUN_API_URL=https://barracuda-api-your-env.a.run.app
```

**Important Notes:**
- `.env.local` files are gitignored and won't be committed
- The CLI loads `.env` first, then `.env.local` (if present)
- Frontend (Vite) only reads variables prefixed with `VITE_` or `PUBLIC_`

## Step 4: Build the Frontend

The frontend must be built before running the Go CLI:

```bash
# From project root
make frontend-build

# Or manually:
cd web
npm run build
cd ..
```

## Step 5: Build and Test the CLI

```bash
# Build the binary (includes embedded frontend)
make build

# Or manually:
go build -o bin/barracuda .

# Test a crawl
./bin/barracuda crawl https://example.com --max-pages 10

# Test with JSON export
./bin/barracuda crawl https://example.com \
  --format json \
  --export results.json \
  --graph-export graph.json
```

## Step 6: Run the API Server Locally

The API server connects to Supabase and provides REST endpoints:

```bash
# Using environment variables from .env.local
go run . api --port 8080

# Or with explicit flags
go run . api \
  --supabase-url https://your-project.supabase.co \
  --supabase-anon-key your-anon-key \
  --supabase-service-key your-service-role-key \
  --port 8080
```

The API will be available at `http://localhost:8080/api/v1/...`

## Step 7: Run the Frontend in Development Mode

For frontend development with hot reload:

```bash
# From project root
make frontend-dev

# Or manually:
cd web
npm run dev
```

The frontend will be available at `http://localhost:5173` and will proxy API requests to `http://localhost:8080`.

## Step 8: Serve Crawl Results Locally

After running a crawl, you can view results in the embedded dashboard:

```bash
# First, crawl with JSON export
./bin/barracuda crawl https://example.com \
  --format json \
  --export results.json \
  --graph-export graph.json

# Then serve the results
./bin/barracuda serve --results results.json --graph graph.json

# Or use the Makefile shortcut
make serve
```

Open `http://localhost:8080` in your browser to view the dashboard.

## Step 9: Testing the Full Stack

### Test Authentication Flow

1. **Start the API server**:
   ```bash
   go run . api --port 8080
   ```

2. **Start the frontend dev server** (in another terminal):
   ```bash
   cd web && npm run dev
   ```

3. **Open the app**:
   - Navigate to `http://localhost:5173`
   - You should see the login/signup page
   - Create an account or sign in

4. **Create a Project**:
   - After logging in, click "Create Project"
   - Enter a project name and domain
   - The project should appear in your projects list

### Test Crawl Upload

1. **Run a crawl**:
   ```bash
   ./bin/barracuda crawl https://example.com \
     --format json \
     --export results.json \
     --graph-export graph.json
   ```

2. **Upload to Supabase** (if you've implemented the upload feature):
   - The crawl results can be uploaded via the API
   - Check the API documentation in `docs/API_SERVER.md`

## Troubleshooting

### Frontend Build Errors

- **Missing dependencies**: Run `cd web && npm install`
- **Build fails**: Check that all dependencies in `package.json` are installed
- **Vite errors**: Clear `web/node_modules` and `web/package-lock.json`, then reinstall

### Supabase Connection Issues

- **"Missing Supabase configuration"**: Check that `.env.local` files exist and have correct values
- **Auth redirect errors**: Verify redirect URLs in Supabase dashboard match your local URLs
- **RLS errors**: Make sure migrations have been applied (`supabase db push`)

### API Server Issues

- **Port already in use**: Change the port with `--port 8081`
- **Supabase connection fails**: Verify `SUPABASE_SERVICE_ROLE_KEY` is set correctly
- **CORS errors**: The API should handle CORS automatically, but check if frontend URL matches

### Database Migration Issues

- **Migration errors**: Check migration files in `supabase/migrations/`
- **Schema out of sync**: Run `supabase db reset` (local) or `supabase db push` (remote)
- **RLS policies**: Review `docs/SUPABASE_SCHEMA.md` for policy requirements

## Quick Reference

### Common Commands

```bash
# Build everything
make build

# Frontend only
make frontend-build

# Run frontend dev server
make frontend-dev

# Run API server
go run . api --port 8080

# Run a crawl
./bin/barracuda crawl https://example.com

# Serve results
make serve

# Run tests
make test
```

### Environment Variables Checklist

- [ ] `PUBLIC_SUPABASE_URL` - Set in root `.env.local` and `web/.env.local`
- [ ] `PUBLIC_SUPABASE_ANON_KEY` - Set in root `.env.local` and `web/.env.local`
- [ ] `SUPABASE_SERVICE_ROLE_KEY` - Set in root `.env.local` (for API server)
- [ ] `VITE_CLOUD_RUN_API_URL` - Optional, for production API

### Useful URLs

- **Local Frontend**: http://localhost:5173
- **Local API**: http://localhost:8080
- **Local Supabase Studio** (if using local Supabase): http://localhost:54323
- **Supabase Dashboard** (remote): https://supabase.com/dashboard

## Next Steps

- Read `docs/API_SERVER.md` for API endpoint documentation
- Check `docs/SUPABASE_SCHEMA.md` for database schema details
- Review `docs/CLOUD_RUN_DEPLOYMENT.md` for production deployment
- See `docs/GSC_SETUP_CHECKLIST.md` for Google Search Console integration

## Getting Help

- Check existing documentation in the `docs/` directory
- Review error messages carefully - they often point to missing configuration
- Verify all environment variables are set correctly
- Ensure Supabase migrations have been applied

