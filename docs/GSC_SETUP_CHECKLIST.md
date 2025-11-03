# Google Search Console API Setup Checklist

## Step 1: Enable the Search Console API

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Select your project: **barracuda-477122**
3. Navigate to **APIs & Services** → **Library**
4. Search for **"Google Search Console API"**
5. Click on it and click **Enable**

## Step 2: Configure OAuth Consent Screen

1. Go to **APIs & Services** → **OAuth consent screen**
2. Fill out the required fields:
   - **User Type**: Choose "External" (unless you're using Google Workspace)
   - **App Name**: Barracuda SEO Crawler
   - **User support email**: Your email
   - **Developer contact**: Your email
3. Click **Save and Continue**

4. On the **Scopes** page:
   - Click **Add or Remove Scopes**
   - Search for: `https://www.googleapis.com/auth/webmasters.readonly`
   - **Check the box** to add it
   - Click **Update** then **Save and Continue**

5. On the **Test users** page (if your app is in Testing mode):
   - **IMPORTANT**: Click **+ ADD USERS**
   - Add your Google account email address (e.g., `dillonlara115@gmail.com`)
   - You can add multiple test users if needed
   - Click **Save and Continue**

6. Review and **Back to Dashboard**

## Common Issue: "Access blocked" Error

If you see "Access blocked: Barracuda has not completed the Google verification process":

**Solution**: Add yourself as a test user:
1. Go to **APIs & Services** → **OAuth consent screen**
2. Scroll to **Test users** section
3. Click **+ ADD USERS**
4. Enter your Google account email: `dillonlara115@gmail.com`
5. Click **Add**
6. Try connecting again

**Note**: Test users can only access the app while it's in Testing mode. For production use, you'll need to publish the app (requires Google verification).

## Step 3: Update Authorized Redirect URI

1. Go to **APIs & Services** → **Credentials**
2. Click on your OAuth 2.0 Client ID
3. Under **Authorized redirect URIs**, make sure you have:
   - `http://localhost:8080/api/gsc/callback`
4. Click **Save**

## Step 4: Verify Setup

The scope you need is:
- **Scope**: `https://www.googleapis.com/auth/webmasters.readonly`
- **Display Name**: "View Search Console data for your verified sites"

This is already configured in the code - you just need to enable it in Google Cloud Console!

## Testing

After setup:
1. Run `barracuda serve --results results.json`
2. Go to Recommendations tab
3. Click "Connect Google Search Console"
4. You should see the consent screen asking for permission to "View Search Console data"
5. Authorize and you're done!

