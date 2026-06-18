---
title: "Android Handset App Setup"
weight: 20
---

# Android Native App (GTasks) Setup & Deployment

The **GTasks** application is a native Android associate portal built using **Kotlin, Jetpack Compose, and Retrofit**. It integrates directly with the Go API Gateway and supports real-time task queues, glassmorphic UI components, and peer-to-peer task trading.

This guide details how to set up your local development environment, import the project into Android Studio, compile using Gradle or Bazel, and deploy to emulators or physical handsets.

---

## 1. Prerequisites & Tooling

Ensure you have the following installed on your development machine:
- **Android Studio:** Ladybug (or newer stable release).
- **Android SDK:** API Level 35 (Android 15.0 / Vanilla Ice Cream).
- **Java Development Kit (JDK):** JDK 17 (or JDK 21). Gradle is pre-configured to compile using Java 17.
- **Android Debug Bridge (ADB):** Installed as part of the Android SDK Platform Tools (ensure `adb` is added to your shell `$PATH`).

---

## 2. Importing the Project into Android Studio

To ensure Android Studio indexes the Kotlin source files, layout resources, and Gradle dependencies correctly, you must **import only the `/apps/gtasks` subdirectory** rather than the root monorepo directory.

1. Open Android Studio.
2. Click **Open** (or **File** > **Open**).
3. Navigate to your local `gemini_task_engine` workspace.
4. Select the **`apps/gtasks`** directory and click **OK**.
5. Allow Android Studio to import and sync Gradle. This will download all Jetpack Compose and Retrofit dependencies and compile local indexing caches (takes 1–2 minutes on first load).

---

## 3. Configuring Local SDK Paths

Android Studio automatically creates a `local.properties` file inside `apps/gtasks/` pointing to your local Android SDK location. If you are building via CLI or encountering SDK location errors, verify that `apps/gtasks/local.properties` exists and contains:

```properties
sdk.dir=/Users/YOUR-USERNAME/Library/Android/sdk
```
*(On Windows, replace with `sdk.dir=C\:\\Users\\YOUR-USERNAME\\AppData\\Local\\Android\\Sdk`)*

---

## 4. Connecting the App to the Local Go Backend

The GTasks app communicates with the Go API Gateway on port `:8080`. Depending on whether you are using an Android Emulator or a physical USB-connected handset, use the appropriate network tunnel:

### A. Android Virtual Device (AVD Emulator)
Android Emulators run inside an isolated virtual network. To reach your laptop's `localhost:8080` from the emulator, the app is pre-configured to route API requests to:
`http://10.0.2.2:8080` *(which is the emulator's internal gateway to your host machine)*.

No extra routing configuration is needed.

---

### B. Physical Android Handsets (USB Debugging)
If you are running the app on a physical phone connected via USB, the phone cannot resolve `localhost` or `10.0.2.2` to your laptop.

To bridge this, you must set up **ADB Reverse Port Forwarding**. Run this command in your terminal:

```bash
adb reverse tcp:8080 tcp:8080
```

> [!TIP]
> This command instructs the ADB daemon on your phone to tunnel all outgoing traffic hitting `localhost:8080` on the phone's loopback interface directly over the USB cable to `localhost:8080` on your development laptop! Run this command every time you plug in your device.

---

## 5. Build & Compilation Workflows

We support two compilation pipelines: native **Gradle** (the Android standard) and **Bazel 8** (the monorepo standard).

### Method A: Standard Gradle (CLI)
You can compile the app directly using the root Gradle wrapper:

```bash
# Navigate to the mobile directory
cd apps/gtasks

# Compile the debug APK
./gradlew assembleDebug
```
*The compiled APK will be generated at:*  
`apps/gtasks/app/build/outputs/apk/debug/app-debug.apk`

---

### Method B: Monorepo Bazel (Recommended)
To keep builds hermetic and aligned with your backend pipelines, you can compile the mobile app using Bazel from the workspace root:

```bash
# Compile via Bazel shell target
bazel build //apps/gtasks:build
```
*Bazel will execute the `build.sh` script inside the sandbox and output the compiled APK to the same destination.*

---

## 6. Handset Installation & Deployment

### Step I: Ensure USB Debugging is Enabled
On your physical handset:
1. Go to **Settings** > **About Phone**.
2. Tap **Build Number** 7 times until you see "You are now a developer!".
3. Go back to **Settings** > **System** > **Developer Options**.
4. Enable **USB Debugging**.

---

### Step II: Install the APK
Once your device is recognized (verify by running `adb devices`), push the compiled APK to your phone by running:

```bash
adb install -r apps/gtasks/app/build/outputs/apk/debug/app-debug.apk
```

The `-r` flag tells ADB to replace the existing installation while preserving its local session cache, allowing you to log in once and deploy updates instantly.
