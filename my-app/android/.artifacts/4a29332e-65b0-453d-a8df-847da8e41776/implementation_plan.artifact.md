# Fix for Login Error in Native App (Mixed Content and CORS)

The app works in the mobile browser because it's likely being accessed via the same IP/Origin as the API, or the browser is less strict with `http` to `http` requests. In the native app, Capacitor uses `https://localhost` as the origin, which causes two problems when calling an `http` API:
1. **Mixed Content**: HTTPS pages cannot call HTTP APIs (Blocked by Chromium).
2. **CORS**: The server might not be configured to allow requests from `https://localhost`.

## Proposed Changes

### [Capacitor Config] (file:///C:/Users/Samuel%20Lucas/Documents/Sistemas%20Freelancer/Sistema%20de%20Frotas%20-%20Eduardo/Sistema-Frota/my-app/capacitor.config.json)

#### [MODIFY] [capacitor.config.json](file:///C:/Users/Samuel%20Lucas/Documents/Sistemas%20Freelancer/Sistema%20de%20Frotas%20-%20Eduardo/Sistema-Frota/my-app/capacitor.config.json)

- Change the `androidScheme` to `http`. This converts the app's origin to `http://localhost`, eliminating the "Mixed Content" security block when calling your `http://192.168.3.38` API.
- Enable `CapacitorHttp` plugin. This will automatically intercept `fetch` calls and run them through native Android code, which bypasses CORS restrictions.

### [Frontend Code] (file:///C:/Users/Samuel%20Lucas/Documents/Sistemas%20Freelancer/Sistema%20de%20Frotas%20-%20Eduardo/Sistema-Frota/my-app/www/Usuario/login.html)

#### [MODIFY] [login.html](file:///C:/Users/Samuel%20Lucas/Documents/Sistemas%20Freelancer/Sistema%20de%20Frotas%20-%20Eduardo/Sistema-Frota/my-app/www/Usuario/login.html)

- Prevent Service Worker registration when running in a native environment. Service Workers have extremely strict security and often block non-HTTPS traffic regardless of WebView settings.

## Verification Plan

### Manual Verification
- Rebuild the app: `npm run build` and then `npx cap copy android`.
- Run the app on the device.
- Try to login.
- Check logcat for any remaining `chromium` errors.
