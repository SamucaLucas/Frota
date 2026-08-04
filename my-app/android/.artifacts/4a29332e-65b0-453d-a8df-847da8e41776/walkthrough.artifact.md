# Walkthrough - Fix for Login Error in Native App (Testing Mode)

I have updated the Capacitor configuration and the frontend code to allow the native app to communicate with your local HTTP server (`http://192.168.3.38:8082`) without being blocked by security policies.

## Changes Made

### 1. Capacitor Configuration
- **File**: [capacitor.config.json](file:///C:/Users/Samuel%20Lucas/Documents/Sistemas%20Freelancer/Sistema%20de%20Frotas%20-%20Eduardo/Sistema-Frota/my-app/capacitor.config.json)
- **Change**: Added `androidScheme: "http"` and enabled the `CapacitorHttp` plugin.
- **Why**: By default, Capacitor uses `https`, which blocks calls to `http` (Mixed Content). Using `http://localhost` as the app's internal origin solves this. `CapacitorHttp` also helps bypass CORS issues.

### 2. Frontend Security (Service Workers)
- **Files**: [login.html](file:///C:/Users/Samuel%20Lucas/Documents/Sistemas%20Freelancer/Sistema%20de%20Frotas%20-%20Eduardo/Sistema-Frota/my-app/www/Usuario/login.html) and [index.html](file:///C:/Users/Samuel%20Lucas/Documents/Sistemas%20Freelancer/Sistema%20de%20Frotas%20-%20Eduardo/Sistema-Frota/my-app/www/index.html)
- **Change**: Added a check to skip Service Worker registration when running as a native app.
- **Why**: Service Workers have extremely strict security that often blocks HTTP traffic even if the WebView is configured to allow it.

## 🛡️ How to Revert for Production (Safe Mode)

When you are ready to move to production and use HTTPS for everything, follow these steps:

> [!IMPORTANT]
> **1. Capacitor Config:**
> In `capacitor.config.json`, change `androidScheme` back to `"https"` or remove the line.

> [!IMPORTANT]
> **2. Service Workers:**
> In `login.html` and `index.html`, remove the condition `!(window.Capacitor && window.Capacitor.isNativePlatform())` from the `navigator.serviceWorker.register` check.

> [!IMPORTANT]
> **3. Android Manifest:**
> Remove `android:usesCleartextTraffic="true"` and `android:networkSecurityConfig="@xml/network_security_config"` from `AndroidManifest.xml` if you no longer need local testing.

## Next Steps for You

Since I modified files in the `www` folder, you need to synchronize these changes with the Android project:

1.  Open your terminal in the project root.
2.  Run: `npx cap copy android`
3.  Build and run the app again in Android Studio.
