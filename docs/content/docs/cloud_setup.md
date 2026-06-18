---
title: "Google Cloud & OAuth Setup"
weight: 15
---

# Google Cloud Platform & OAuth 2.0 Identity Configuration

This guide provides step-by-step instructions for configuring the Google Cloud Platform (GCP) resources, OAuth 2.0 Consent Screens, and Client Credentials required to run both the **React Admin Console** and the **GTasks Android Native Application** with secure Google Sign-In.

---

## 1. Google Cloud Project Setup

1. Open the [Google Cloud Console](https://console.cloud.google.com/).
2. Click the project dropdown in the top navigation bar and select **New Project**.
3. Name the project `Gemini-Task-Engine` (or your preferred identifier).
4. Assign an organization/billing account if required and click **Create**.

---

## 2. OAuth Consent Screen Configuration

Before generating credentials, you must configure the consent screen that users see when signing in.

1. Navigate to **APIs & Services** > **OAuth consent screen** in the GCP sidebar.
2. Choose your **User Type**:
   - **Internal:** (Recommended for Enterprise/Testing) Restricts authentication to users within your Google Workspace domain.
   - **External:** Permits any Google account to sign in (requires entering test users before publishing).
3. Click **Create**.
4. Fill in the **App Information**:
   - **App name:** `Gemini Task Engine`
   - **User support email:** Select your administrator email.
   - **Developer contact information:** Enter your email.
5. Click **Save and Continue**.
6. **Scopes:** Click **Add or Remove Scopes** and select:
   - `.../auth/userinfo.email` (View primary email address)
   - `.../auth/userinfo.profile` (View basic profile information)
   - `openid` (Associate identity with Google)
7. Click **Save and Continue**.
8. (If External) **Test Users:** Add the Google Accounts you intend to use for testing on your handset.
9. Review the summary and click **Back to Dashboard**.

---

## 3. Creating OAuth 2.0 Client Credentials

To secure the authentication pipeline, you must provision two separate Client IDs: one for the **React Web Console** (which acts as a Single Page Application) and one for the **Android Handset App** (which signs the request with your developer certificate).

### A. Web Application Credentials (Admin Console)

1. Navigate to **APIs & Services** > **Credentials**.
2. Click **+ Create Credentials** > **OAuth client ID**.
3. Select **Application type:** `Web application`.
4. Name the client: `Gemini Task Engine - Web Console`.
5. Under **Authorized JavaScript origins**, add:
   - `http://localhost:5173` (Vite Admin Console Dev Port)
   - `http://localhost:8080` (Go API Gateway Port)
6. Under **Authorized redirect URIs**, add:
   - `http://localhost:8080/api/v1/auth/callback` (Go oauth callback route)
7. Click **Create**.
8. **CRITICAL:** Copy the **Client ID** and **Client Secret** immediately. You will need these for your environment configuration.

---

### B. Android Application Credentials (GTasks Mobile App)

Google OAuth on Android validates both the application's package name and the cryptographical signature of your developer certificate to prevent client spoofing.

#### Step I: Extract your Developer SHA-1 Fingerprint
To authenticate your local debug builds, you must extract the SHA-1 fingerprint of your local Android debug keystore (created automatically by Gradle or Android Studio).

Run the following command in your terminal:

```bash
keytool -list -v \
  -keystore ~/.android/debug.keystore \
  -alias androiddebugkey \
  -storepass android \
  -keypass android
```

> [!NOTE]
> Under Windows, replace the keystore path with `%USERPROFILE%\.android\debug.keystore`.

Locate the **SHA-1** fingerprint in the terminal output. It will look similar to this:
`SHA1: DE:AD:BE:EF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF`

#### Step II: Register the Android Client in GCP
1. Go back to **APIs & Services** > **Credentials** in the Cloud Console.
2. Click **+ Create Credentials** > **OAuth client ID**.
3. Select **Application type:** `Android`.
4. Name the client: `Gemini Task Engine - Android Client`.
5. Enter the exact **Package name:** `com.google.gtasks` (defined in the app's `AndroidManifest.xml`).
6. Paste the **SHA-1 fingerprint** you extracted in Step I.
7. Click **Create**.
8. Copy the generated **Client ID**.

---

## 4. Environment Variable Configuration

Once you have provisioned your credentials, you must expose them to the Go API gateway using your local configuration files.

1. Locate the `.env.local.toml` file at the root of your workspace (create it if it does not exist).
2. Add or update the `[auth.google]` block:

```toml
[auth]
jwt_secret = "your-custom-secure-jwt-signing-key" # Define a secure token signing secret

[auth.google]
client_id = "PASTE-YOUR-WEB-CLIENT-ID-HERE.apps.googleusercontent.com"
client_secret = "PASTE-YOUR-WEB-CLIENT-SECRET-HERE"
```

> [!IMPORTANT]
> Keep your `.env.local.toml` out of version control. The root `.gitignore` is pre-configured to exclude `.env.local.toml` to prevent leaking enterprise client secrets.

---

## 5. Local Database Seeding
To match Google Sign-In identities to GTasks database records, ensure that the email address associated with your Google account is registered in the `users` table.

You can seed your developer account by editing your local `scripts/dev_events.sql` file and adding:

```sql
INSERT INTO users (id, name, email, o_auth_provider, o_auth_id, created_at, updated_at) 
VALUES (
  'b75c1a02-c884-40ed-a3f8-8b95f3ff7539', -- Your User UUID
  'Ryan McGuinness',                     -- Your Display Name
  'ryan@rmcguinness.altostrat.com',      -- Your Google Sign-In Email
  'google',
  '113329281712545510375',               -- Your Google User ID (optional, bound on first login)
  NOW(), NOW()
) ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name;
```

Reload the database seeds using the loader utility:
```bash
bazel run //cmd/db_loader -- -migrate -file=scripts/dev_env.sql,scripts/time_zones.sql,scripts/icao_codes.sql,scripts/dev_events.sql,scripts/seed_store_tasks.sql
```
