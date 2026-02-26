# Flutterize

Cyber security learning project — Go REST API + Flutter Android + MockMail + MockSMS.

> [!CAUTION]
> **This application contains intentional security vulnerabilities.**  
> It is designed **strictly for educational and authorized security testing purposes only.**  
> Do NOT deploy this application in any production environment or expose it to untrusted networks.

> [!WARNING]
> By using this project, you acknowledge that:
> - The included vulnerabilities can be exploited to compromise data and systems
> - You are solely responsible for how you use this software
> - Use this only in isolated lab environments that you own or have explicit permission to test
> - See [VULNERABILITIES.md](VULNERABILITIES.md) for the full list of intentional vulnerabilities

## Prerequisites

- Docker & Docker Compose
- mkcert (`sudo apt install mkcert && mkcert -install`)
- Flutter SDK (for building the Android app)
- Go 1.25+ (for local development)

## Quick Start

```bash
# 1. Copy environment file and adjust values
cp .env.example .env

# 2. Generate TLS certificates
make gen-certs

# 3. Start all services (API + MySQL + MockMail + MockSMS)
make up

# 4. View logs
make logs
```

## Vulnerability Mode

Each vulnerability can be toggled individually by swapping source files between secure and vulnerable versions.

```bash
make list                  # Show all vulnerabilities with ON/OFF status
make vulnerable 1 3 10     # Enable specific vulnerabilities by number
make vulnerable all        # Enable all vulnerabilities
make secure                # Disable all (restore secure code)
```

> [!NOTE]
> Vulns **4, 5, 6, 7** (mass assignment, user enumeration, token leak, stored XSS) share the same source files and toggle as a group.
> Flutter client vulns (**8, 9**) require hot-reloading the app after toggling.

See [VULNERABILITIES.md](VULNERABILITIES.md) for exploitation guides and remediation hints.

## Services

| Service   | URL                          | Description                    |
|-----------|------------------------------|--------------------------------|
| API       | https://localhost:8443       | Go REST API (HTTPS)            |
| MockMail  | http://localhost:8025        | Email viewer (SMTP on :2525)   |
| MockSMS   | http://localhost:7600        | SMS viewer                     |
| MySQL     | localhost:3306               | Database                       |

## Seed Users

| Email                  | Phone          | Password     | Points |
|------------------------|----------------|------------- |--------|
| admin@flutterize.lab   | 081234567890   | password123  | 1000   |
| user@flutterize.lab    | 081234567891   | password123  | 500    |

## Flutter App

```bash
cd app
flutter pub get
flutter run
```

> For Android emulator, the API is reachable at `https://10.0.2.2:8443`

## License

This project is licensed under the [MIT License](LICENSE).
