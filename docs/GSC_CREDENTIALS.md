# Google Search Console OAuth Credentials

## Setup

Create a `.env` file in the project root (or set environment variables) with your GSC credentials:

```bash
GSC_CLIENT_ID=your-client-id.apps.googleusercontent.com
GSC_CLIENT_SECRET=your-client-secret
```

Or use a JSON credentials file:

```bash
GSC_CREDENTIALS_JSON='{"web":{"client_id":"...","client_secret":"...","redirect_uris":["http://localhost:8080/api/gsc/callback"]}}'
```

## Usage

### Option 1: Using .env file (Recommended)

1. Create `.env` file in project root:
   ```bash
   cp .env.example .env
   ```

2. Add your credentials to `.env`:
   ```bash
   GSC_CLIENT_ID=your-client-id.apps.googleusercontent.com
   GSC_CLIENT_SECRET=your-client-secret
   ```

3. Run the server:
   ```bash
   # Load .env and run
   export $(cat .env | xargs) && barracuda serve --results results.json
   ```

### Option 2: Export environment variables

```bash
export GSC_CLIENT_ID='your-client-id.apps.googleusercontent.com'
export GSC_CLIENT_SECRET='your-client-secret'
barracuda serve --results results.json
```

### Option 3: Use in shell script

Create `start-server.sh`:
```bash
#!/bin/bash
export GSC_CLIENT_ID='your-client-id.apps.googleusercontent.com'
export GSC_CLIENT_SECRET='your-client-secret'
barracuda serve --results results.json "$@"
```

Then run: `chmod +x start-server.sh && ./start-server.sh`

## Security Notes

⚠️ **Never commit credentials to git!**

- `.env` files are already in `.gitignore`
- Keep credentials secure and private
- Use different credentials for development vs production
- Rotate credentials if they're exposed

## Getting Credentials

See `docs/GSC_SETUP_CHECKLIST.md` for detailed instructions on creating Google Cloud OAuth credentials.

