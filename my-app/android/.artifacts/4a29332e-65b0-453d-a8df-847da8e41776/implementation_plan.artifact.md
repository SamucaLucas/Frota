# Fix for Mixed Content Error (HTTPS to HTTP)

The application is running on `https://localhost` (Capacitor default) and trying to access an API at `http://192.168.3.38:8082`. The Android WebView blocks this by default as it is "Mixed Content" (insecure HTTP request from a secure HTTPS page).

## Proposed Changes

### [app] (file:///C:/Users/Samuel%20Lucas/Documents/Sistemas%20Freelancer/Sistema%20de%20Frotas%20-%20Eduardo/Sistema-Frota/my-app/android/app)

#### [MODIFY] [MainActivity.java](file:///C:/Users/Samuel%20Lucas/Documents/Sistemas%20Freelancer/Sistema%20de%20Frotas%20-%20Eduardo/Sistema-Frota/my-app/android/app/src/main/java/com/sousa/app/MainActivity.java)

- Override the `onStart` method to configure the WebView to allow mixed content.
- Use `WebSettings.MIXED_CONTENT_ALWAYS_ALLOW` to permit HTTP requests from the HTTPS local server.

#### [NEW] [network_security_config.xml](file:///C:/Users/Samuel%20Lucas/Documents/Sistemas%20Freelancer/Sistema%20de%20Frotas%20-%20Eduardo/Sistema-Frota/my-app/android/app/src/main/res/xml/network_security_config.xml)

- Explicitly allow cleartext traffic for the local IP `192.168.3.38` and `localhost`.
- This ensures the Android network layer doesn't block the request even if the WebView allows it.

#### [MODIFY] [AndroidManifest.xml](file:///C:/Users/Samuel%20Lucas/Documents/Sistemas%20Freelancer/Sistema%20de%20Frotas%20-%20Eduardo/Sistema-Frota/my-app/android/app/src/main/AndroidManifest.xml)

- Reference the new `network_security_config.xml` in the `<application>` tag.
- (Optional but recommended) Keep `android:usesCleartextTraffic="true"`.

## Verification Plan

### Manual Verification
- Deploy the app to the device.
- Attempt to login again.
- Verify in the Inspector that the "Mixed Content" error no longer appears and the request succeeds.
