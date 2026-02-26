#!/bin/bash
# toggle-vuln.sh — Switch between vulnerable and secure code
set -e

MODE="$1"
BASE_DIR="$(cd "$(dirname "$0")/.." && pwd)"

if [ "$MODE" != "vulnerable" ] && [ "$MODE" != "secure" ]; then
    echo "Usage: $0 <vulnerable|secure>"
    echo "  vulnerable  — Apply all vulnerability patches"
    echo "  secure      — Revert to secure code"
    exit 1
fi

SOURCE_DIR="$BASE_DIR/$MODE"

echo "==> Switching to $MODE mode..."

# Backend file mappings
cp "$SOURCE_DIR/middleware.go"        "$BASE_DIR/backend/middleware/middleware.go"
cp "$SOURCE_DIR/handler_auth.go"      "$BASE_DIR/backend/handler/auth.go"
cp "$SOURCE_DIR/handler_user.go"      "$BASE_DIR/backend/handler/user.go"
cp "$SOURCE_DIR/handler_points.go"    "$BASE_DIR/backend/handler/points.go"
cp "$SOURCE_DIR/repository_user.go"   "$BASE_DIR/backend/repository/user.go"
cp "$SOURCE_DIR/repository_points.go" "$BASE_DIR/backend/repository/points.go"
echo "    ✓ Backend files swapped"

# Flutter file mappings
cp "$SOURCE_DIR/main.dart"         "$BASE_DIR/app/lib/main.dart"
cp "$SOURCE_DIR/api_service.dart"  "$BASE_DIR/app/lib/services/api_service.dart"
echo "    ✓ Flutter files swapped"

# Rebuild backend containers
echo "==> Rebuilding backend containers..."
cd "$BASE_DIR"
docker compose down 2>/dev/null || true
docker compose up -d --build

echo ""
echo "============================================"
if [ "$MODE" = "vulnerable" ]; then
    echo "  ⚠️  VULNERABLE MODE ACTIVE"
    echo "  All security vulnerabilities are enabled"
    echo ""
    echo "  Vulnerabilities enabled:"
    echo "    #1  JWT none algorithm"
    echo "    #2  IDOR on profile"
    echo "    #3  Negative amount transfer"
    echo "    #4  Mass assignment"
    echo "    #5  User enumeration"
    echo "    #6  Reset token leak"
    echo "    #7  Stored XSS (no sanitization)"
    echo "    #8  SSL pinning bypassed"
    echo "    #9  Root detection disabled"
    echo "    #10 Race condition"
else
    echo "  🔒 SECURE MODE ACTIVE"
    echo "  All security patches applied"
    echo ""
    echo "  Protections enabled:"
    echo "    JWT algorithm validation"
    echo "    Profile access via JWT only"
    echo "    Amount validation"
    echo "    Points balance hardcoded"
    echo "    Generic error messages"
    echo "    Token via email/SMS only"
    echo "    HTML sanitization"
    echo "    SSL pinning enforced"
    echo "    Root/emulator detection"
    echo "    Row-level locking"
fi
echo "============================================"
echo ""
echo "  Note: Hot-reload the Flutter app to apply"
echo "  client-side changes (#8, #9)."
