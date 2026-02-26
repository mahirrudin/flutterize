#!/bin/bash
# toggle-vuln.sh — Granular vulnerability toggle via file-swap
set -e

BASE_DIR="$(cd "$(dirname "$0")/.." && pwd)"
STATE_FILE="$BASE_DIR/vuln.state"
VULN_DIR="$BASE_DIR/vulnerable"
SEC_DIR="$BASE_DIR/secure"

# ── Vulnerability definitions ──────────────────────────────────
# Format: ID|Name|Difficulty|Files (pipe-separated)
VULNS=(
  "1|JWT none algorithm|Easy|middleware.go"
  "2|IDOR on profile|Easy|handler_user.go"
  "3|Negative amount transfer|Easy|handler_points.go"
  "4|Mass assignment|Medium|handler_auth.go,repository_user.go"
  "5|User enumeration|Easy|handler_auth.go,repository_user.go"
  "6|Reset token leak|Medium|handler_auth.go,repository_user.go"
  "7|Stored XSS|Medium|handler_auth.go,repository_user.go"
  "8|SSL pinning bypass|Hard|api_service.dart"
  "9|Root detection bypass|Hard|main.dart"
  "10|Race condition|Hard|repository_points.go"
)

# Auth vulns share the same files — enabling any one enables all
AUTH_GROUP="4 5 6 7"

# ── File destinations ──────────────────────────────────────────
declare -A FILE_DEST
FILE_DEST["middleware.go"]="backend/middleware/middleware.go"
FILE_DEST["handler_auth.go"]="backend/handler/auth.go"
FILE_DEST["handler_user.go"]="backend/handler/user.go"
FILE_DEST["handler_points.go"]="backend/handler/points.go"
FILE_DEST["repository_user.go"]="backend/repository/user.go"
FILE_DEST["repository_points.go"]="backend/repository/points.go"
FILE_DEST["main.dart"]="app/lib/main.dart"
FILE_DEST["api_service.dart"]="app/lib/services/api_service.dart"

# ── State helpers ──────────────────────────────────────────────
load_state() {
  ENABLED=()
  if [ -f "$STATE_FILE" ]; then
    while IFS= read -r id; do
      ENABLED+=("$id")
    done < "$STATE_FILE"
  fi
}

save_state() {
  printf "%s\n" "${ENABLED[@]}" | sort -n | uniq > "$STATE_FILE"
}

is_enabled() {
  local id="$1"
  for e in "${ENABLED[@]}"; do
    [ "$e" = "$id" ] && return 0
  done
  return 1
}

# ── Core functions ─────────────────────────────────────────────
get_vuln_field() {
  local id="$1" field="$2"
  for entry in "${VULNS[@]}"; do
    IFS='|' read -r vid name diff files <<< "$entry"
    if [ "$vid" = "$id" ]; then
      case "$field" in
        name)  echo "$name" ;;
        diff)  echo "$diff" ;;
        files) echo "$files" ;;
      esac
      return
    fi
  done
}

get_files_for_vuln() {
  get_vuln_field "$1" files
}

swap_file() {
  local file="$1" mode="$2"
  local src dest
  dest="${FILE_DEST[$file]}"
  if [ -z "$dest" ]; then return; fi
  if [ "$mode" = "vulnerable" ]; then
    src="$VULN_DIR/$file"
  else
    src="$SEC_DIR/$file"
  fi
  cp "$src" "$BASE_DIR/$dest"
}

# Determine state per-file: if ANY vuln using that file is enabled → vulnerable
resolve_file_state() {
  local file="$1"
  for entry in "${VULNS[@]}"; do
    IFS='|' read -r vid name diff files <<< "$entry"
    IFS=',' read -ra flist <<< "$files"
    for f in "${flist[@]}"; do
      if [ "$f" = "$file" ] && is_enabled "$vid"; then
        echo "vulnerable"
        return
      fi
    done
  done
  echo "secure"
}

apply_all_files() {
  local changed_backend=false
  local changed_flutter=false

  for file in "${!FILE_DEST[@]}"; do
    local state
    state=$(resolve_file_state "$file")
    swap_file "$file" "$state"

    case "$file" in
      *.go) changed_backend=true ;;
      *.dart) changed_flutter=true ;;
    esac
  done

  if $changed_backend; then
    echo "    ✓ Backend files swapped"
    echo "==> Rebuilding backend containers..."
    cd "$BASE_DIR"
    docker compose up -d --build 2>&1 | tail -5
  fi

  if $changed_flutter; then
    echo "    ✓ Flutter files swapped"
    echo "    Note: Hot-reload the Flutter app to apply client-side changes"
  fi
}

# ── Commands ───────────────────────────────────────────────────
cmd_list() {
  load_state
  echo ""
  echo "  #   Status   Difficulty   Vulnerability"
  echo "  ──  ──────   ──────────   ─────────────────────────────"
  for entry in "${VULNS[@]}"; do
    IFS='|' read -r vid name diff files <<< "$entry"
    if is_enabled "$vid"; then
      status="⚠️  ON "
    else
      status="   OFF"
    fi
    case "$diff" in
      Easy)   color="🟢" ;;
      Medium) color="🟡" ;;
      Hard)   color="🔴" ;;
    esac
    printf "  %-3s %s   %s %-8s  %s\n" "$vid" "$status" "$color" "$diff" "$name"
  done
  echo ""

  # Count enabled
  local count=0
  for e in "${ENABLED[@]}"; do ((count++)) || true; done
  if [ "$count" -gt 0 ]; then
    echo "  ⚠️  $count/10 vulnerabilities enabled"
  else
    echo "  🔒 All secure"
  fi
  echo ""
}

cmd_enable() {
  load_state
  local requested=("$@")
  local added=()

  for id in "${requested[@]}"; do
    # Validate
    if ! get_vuln_field "$id" name > /dev/null 2>&1 || [ -z "$(get_vuln_field "$id" name)" ]; then
      echo "  ✗ Unknown vulnerability: #$id"
      continue
    fi

    # If it's an auth vuln (4-7), enable the whole group
    local to_enable=("$id")
    for g in $AUTH_GROUP; do
      if [ "$id" = "$g" ]; then
        to_enable=($AUTH_GROUP)
        break
      fi
    done

    for eid in "${to_enable[@]}"; do
      if ! is_enabled "$eid"; then
        ENABLED+=("$eid")
        added+=("$eid")
      fi
    done
  done

  if [ ${#added[@]} -eq 0 ]; then
    echo "  No changes — already enabled"
    return
  fi

  save_state
  echo "==> Enabling vulnerabilities..."
  for a in "${added[@]}"; do
    echo "    ⚠️  #$a $(get_vuln_field "$a" name)"
  done
  echo ""

  apply_all_files
  echo ""
  cmd_list
}

cmd_enable_all() {
  load_state
  ENABLED=()
  for entry in "${VULNS[@]}"; do
    IFS='|' read -r vid _ _ _ <<< "$entry"
    ENABLED+=("$vid")
  done
  save_state

  echo "==> Enabling ALL vulnerabilities..."
  apply_all_files
  echo ""
  cmd_list
}

cmd_secure() {
  ENABLED=()
  save_state

  echo "==> Disabling all vulnerabilities..."
  apply_all_files
  echo ""
  echo "  🔒 SECURE MODE — All protections applied"
  echo ""
}

cmd_disable() {
  load_state
  local requested=("$@")
  local removed=()

  for id in "${requested[@]}"; do
    if [ -z "$(get_vuln_field "$id" name)" ]; then
      echo "  ✗ Unknown vulnerability: #$id"
      continue
    fi

    # If it's an auth vuln (4-7), disable the whole group
    local to_disable=("$id")
    for g in $AUTH_GROUP; do
      if [ "$id" = "$g" ]; then
        to_disable=($AUTH_GROUP)
        break
      fi
    done

    for did in "${to_disable[@]}"; do
      if is_enabled "$did"; then
        local new_enabled=()
        for e in "${ENABLED[@]}"; do
          [ "$e" != "$did" ] && new_enabled+=("$e")
        done
        ENABLED=("${new_enabled[@]}")
        removed+=("$did")
      fi
    done
  done

  if [ ${#removed[@]} -eq 0 ]; then
    echo "  No changes — already disabled"
    return
  fi

  save_state
  echo "==> Disabling vulnerabilities..."
  for r in "${removed[@]}"; do
    echo "    🔒 #$r $(get_vuln_field "$r" name)"
  done
  echo ""

  apply_all_files
  echo ""
  cmd_list
}

# ── Main ───────────────────────────────────────────────────────
ACTION="$1"
shift 2>/dev/null || true

case "$ACTION" in
  list)
    cmd_list
    ;;
  on)
    if [ "$1" = "all" ]; then
      cmd_enable_all
    elif [ $# -gt 0 ]; then
      cmd_enable "$@"
    else
      echo "Usage: $0 on <1 2 3 ...> | all"
    fi
    ;;
  off)
    if [ $# -gt 0 ]; then
      cmd_disable "$@"
    else
      echo "Usage: $0 off <1 2 3 ...>"
    fi
    ;;
  secure)
    cmd_secure
    ;;
  *)
    echo "Usage: toggle-vuln.sh <command>"
    echo ""
    echo "Commands:"
    echo "  list              Show vulnerability status"
    echo "  on <IDs...>       Enable vulnerabilities (e.g. on 1 3 10)"
    echo "  on all            Enable all vulnerabilities"
    echo "  off <IDs...>      Disable vulnerabilities (e.g. off 2 5)"
    echo "  secure            Disable all vulnerabilities"
    echo ""
    echo "Vulns 4,5,6,7 share the same files — they toggle as a group."
    exit 1
    ;;
esac
