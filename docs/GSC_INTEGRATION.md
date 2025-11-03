# Google Search Console Integration

This document explains how to set up and use the Google Search Console (GSC) integration with Barracuda.

## Overview

The GSC integration enhances SEO recommendations by:
- Prioritizing fixes based on actual search traffic
- Identifying CTR optimization opportunities
- Providing context-aware recommendations based on performance data

## How It Works

**Good news:** You don't need to create your own Google Cloud project! Just click "Connect" and authorize Barracuda with your Google account.

OAuth works like this:
- Barracuda provides OAuth credentials (like any Google app)
- You authorize Barracuda to access your Search Console data
- No Google Cloud project needed on your end!

## Setup

### Step 1: Enable Search Console API in Google Cloud

Before using GSC integration, you need to:

1. **Enable the API:**
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Select your project (**barracuda-477122**)
   - Navigate to **APIs & Services** → **Library**
   - Search for "Google Search Console API"
   - Click **Enable**

2. **Configure OAuth Consent Screen:**
   - Go to **APIs & Services** → **OAuth consent screen**
   - Fill out required fields (App name, email, etc.)
   - On **Scopes** page, add: `https://www.googleapis.com/auth/webmasters.readonly`
   - **IMPORTANT**: On **Test users** page, click **+ ADD USERS** and add your email (`dillonlara115@gmail.com`)
   - Click **Save and Continue**

3. **Update Redirect URI:**
   - Go to **APIs & Services** → **Credentials**
   - Click your OAuth 2.0 Client ID
   - Add redirect URI: `http://localhost:8080/api/gsc/callback`
   - Save

See `docs/GSC_SETUP_CHECKLIST.md` for detailed step-by-step instructions.

### Step 2: Use Barracuda

Once the API is enabled:

1. Start the server: `barracuda serve --results results.json`
2. Go to the **Recommendations** tab
3. Click **Connect Google Search Console**
4. Authorize with your Google account
5. Select your Search Console property
6. Click **Enrich Issues with GSC Data**
7. Done! 🎉

### Option 2: Use Your Own Credentials (Advanced)

If you want to use your own Google Cloud project instead of the default credentials:

1. Create a Google Cloud Project:
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project or select an existing one
   - Enable the **Google Search Console API**

2. Create OAuth Credentials:
   - Navigate to **APIs & Services** > **Credentials**
   - Click **Create Credentials** > **OAuth client ID**
   - Choose **Web application**
   - Add authorized redirect URI: `http://localhost:8080/api/gsc/callback`
   - Save your **Client ID** and **Client Secret**

3. Configure Barracuda:

Set environment variables before running the server:

```bash
export GSC_CLIENT_ID="your-client-id.apps.googleusercontent.com"
export GSC_CLIENT_SECRET="your-client-secret"
```

Or use a JSON credentials file:

```bash
export GSC_CREDENTIALS_JSON='{"web":{"client_id":"...","client_secret":"...","redirect_uris":["http://localhost:8080/api/gsc/callback"]}}'
```

### Install Dependencies

Run `go mod tidy` to install required packages:
- `golang.org/x/oauth2`
- `google.golang.org/api/searchconsole/v1`

## Usage

### 1. Start the Server

```bash
barracuda serve --results results.json
```

### 2. Connect GSC

1. Navigate to the **Recommendations** tab in the dashboard
2. Click **Connect Google Search Console**
3. Authorize Barracuda in the popup window
4. Select your property from the dropdown
5. Click **Enrich Issues with GSC Data**

### 3. View Enhanced Recommendations

Once connected, recommendations will show:
- **Traffic-based priority**: High-traffic pages get higher priority
- **CTR insights**: Pages with low CTR but high impressions get optimization suggestions
- **Performance context**: Recommendations include actual search performance data

## API Endpoints

- `GET /api/gsc/connect` - Get OAuth authorization URL
- `GET /api/gsc/callback` - OAuth callback handler
- `GET /api/gsc/properties` - List available GSC properties
- `POST /api/gsc/performance` - Fetch performance data for a property
- `POST /api/gsc/enrich-issues` - Merge GSC data with crawl issues

## How It Works

1. **OAuth Flow**: User authorizes access via Google OAuth
2. **Property Selection**: User selects which GSC property to use
3. **Data Fetching**: System fetches last 30 days of performance data
4. **URL Matching**: GSC URLs are matched with crawled URLs
5. **Enrichment**: Issues are enriched with:
   - Impressions count
   - Clicks count
   - CTR (Click-Through Rate)
   - Average position
   - Top queries

6. **Priority Calculation**: Enhanced priority scores consider:
   - Traffic volume (high-traffic pages prioritized)
   - CTR opportunities (low CTR with high impressions)
   - Ranking position (pages close to top 10)

## Example Enhanced Recommendation

**Before GSC:**
> Missing meta description - Medium impact

**After GSC:**
> Missing meta description - **HIGH impact**
> 
> This page has high search visibility (15,000 impressions/month). Fixing this issue could significantly impact your SEO performance.
> 
> GSC Data:
> - Impressions: 15,000/month
> - Clicks: 480
> - CTR: 3.2%
> - Position: 8.5

## Troubleshooting

### "Access Blocked" Error (403: access_denied)

**This is the most common issue!** If you see:
> "Access blocked: Barracuda has not completed the Google verification process"

**Solution**: Add yourself as a test user in Google Cloud Console:

1. Go to [Google Cloud Console](https://console.cloud.google.com/) → **APIs & Services** → **OAuth consent screen**
2. Scroll down to the **Test users** section
3. Click **+ ADD USERS**
4. Enter your Google account email: `dillonlara115@gmail.com`
5. Click **Add**
6. Try connecting again in Barracuda

**Why this happens**: While your app is in Testing mode, only explicitly added test users can access it. This is Google's security measure.

**For production**: To allow anyone to use the app, you'll need to publish it (which requires Google verification - a longer process).

### GSC Integration Disabled

If you see "GSC integration disabled" in the server logs:
- Check that environment variables are set correctly
- Verify the redirect URI matches your server URL
- Ensure Google Search Console API is enabled in your Google Cloud project

### No Properties Found

- Verify you have access to properties in Google Search Console
- Check that the OAuth flow completed successfully
- Refresh properties using the "Refresh Properties" button

### Missing Performance Data

- Some pages may not have search data (new pages, low-traffic pages)
- Performance data is limited to last 16 months (GSC API limit)
- Ensure the property URL matches your crawled domain

## Security Notes

- OAuth tokens are stored in memory (single server instance)
- For production, consider persistent token storage
- Tokens expire and are automatically refreshed
- Never expose client secrets in public repositories

## Limitations

- Performance data is cached for the session
- Maximum 25,000 URLs per property (GSC API limit)
- Data refresh requires manual action
- Single user per server instance (IP-based identification)

