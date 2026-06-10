#!/usr/bin/env bash
#
# AI.sh — prepare a local Ollama + Gemma 4 for chklst's AI features
# (daily activity summary, analytics narratives).
#
# It will:
#   1. detect the OS / distro
#   2. ensure Ollama is installed and new enough for Gemma 4 (>= 0.30.6)
#   3. ensure the Ollama service is running (bound for Docker)
#   4. pull the model (default gemma4:e4b; override with MODEL=... ./AI.sh)
#   5. verify with a test generation
#
# Safe to re-run. No system files are touched without asking; the Ollama upgrade
# is a user-local install under ~/.local.
#
set -euo pipefail

MODEL="${MODEL:-gemma4:e4b}"  # override with: MODEL=gemma4:e4b ./AI.sh
MIN_VERSION="0.30.6"          # first Ollama release that can pull Gemma 4
OLLAMA_HOST="${OLLAMA_HOST:-http://localhost:11434}"
LOCAL_PREFIX="$HOME/.local"   # where a user-local Ollama lives

# ---- pretty output -------------------------------------------------------
c_reset=$'\033[0m'; c_blue=$'\033[34m'; c_green=$'\033[32m'; c_yellow=$'\033[33m'; c_red=$'\033[31m'
info()  { printf "%s==>%s %s\n" "$c_blue" "$c_reset" "$*"; }
ok()    { printf "%s✓%s %s\n"  "$c_green" "$c_reset" "$*"; }
warn()  { printf "%s!%s %s\n"  "$c_yellow" "$c_reset" "$*"; }
err()   { printf "%s✗%s %s\n"  "$c_red" "$c_reset" "$*" >&2; }
ask()   { # ask "question" -> returns 0 for yes
  printf "%s?%s %s [y/N] " "$c_yellow" "$c_reset" "$*"
  read -r ans </dev/tty || ans=""
  [ "$ans" = "y" ] || [ "$ans" = "Y" ]
}

# ---- 1. detect OS --------------------------------------------------------
detect_os() {
  OS="$(uname -s)"
  DISTRO="unknown"
  if [ "$OS" = "Linux" ] && [ -r /etc/os-release ]; then
    # shellcheck disable=SC1091
    . /etc/os-release
    DISTRO="${ID:-unknown}"
    info "Detected: ${PRETTY_NAME:-Linux} ($(uname -m))"
  elif [ "$OS" = "Darwin" ]; then
    DISTRO="macos"
    info "Detected: macOS ($(uname -m))"
  else
    info "Detected: $OS ($(uname -m))"
  fi
}

# ---- version helpers -----------------------------------------------------
# returns 0 if $1 >= $2 (dotted versions)
version_ge() {
  [ "$(printf '%s\n%s\n' "$2" "$1" | sort -V | head -n1)" = "$2" ]
}

ollama_version() {
  # prints the binary's own version, even with no server running
  ollama --version 2>&1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -n1
}

# ---- 2. ensure Ollama ----------------------------------------------------
install_ollama_linux_local() {
  info "Installing/upgrading Ollama (user-local) under $LOCAL_PREFIX ..."
  tmp="$(mktemp -d)"
  curl -fsSL https://ollama.com/download/ollama-linux-amd64.tgz -o "$tmp/ollama.tgz"
  mkdir -p "$LOCAL_PREFIX"
  tar -xzf "$tmp/ollama.tgz" -C "$LOCAL_PREFIX"
  rm -rf "$tmp"
  case ":$PATH:" in
    *":$LOCAL_PREFIX/bin:"*) : ;;
    *) warn "$LOCAL_PREFIX/bin is not on PATH — add it to use 'ollama' directly." ;;
  esac
}

ensure_ollama() {
  if ! command -v ollama >/dev/null 2>&1; then
    warn "Ollama is not installed."
    case "$DISTRO" in
      macos)
        err "Install Ollama from https://ollama.com/download (or 'brew install ollama'), then re-run."
        exit 1 ;;
      *)
        if ask "Install Ollama now (user-local, no sudo)?"; then
          install_ollama_linux_local
        else
          err "Ollama is required. Aborting."; exit 1
        fi ;;
    esac
  fi

  local v; v="$(ollama_version || true)"
  if [ -z "$v" ]; then
    warn "Could not determine Ollama version; continuing."
    return
  fi
  if version_ge "$v" "$MIN_VERSION"; then
    ok "Ollama $v (>= $MIN_VERSION, supports Gemma 4)."
  else
    warn "Ollama $v is too old for Gemma 4 (need >= $MIN_VERSION)."
    case "$DISTRO" in
      macos)
        err "Upgrade Ollama from https://ollama.com/download, then re-run."; exit 1 ;;
      *)
        if ask "Upgrade Ollama to the latest (user-local) now?"; then
          # stop any running server so the binary can be replaced cleanly
          pkill -f "ollama serve" 2>/dev/null || true
          sleep 1
          install_ollama_linux_local
          ok "Upgraded to $(ollama_version)."
        else
          err "Gemma 4 needs Ollama >= $MIN_VERSION. Aborting."; exit 1
        fi ;;
    esac
  fi
}

# ---- 3. ensure the service is running ------------------------------------
ensure_running() {
  if curl -fsS "$OLLAMA_HOST/api/version" >/dev/null 2>&1; then
    # Already up — but is it reachable from Docker (bound to 0.0.0.0, not just localhost)?
    if ss -ltn 2>/dev/null | grep -q '0\.0\.0\.0:11434'; then
      ok "Ollama service is running and reachable from Docker."
      return
    fi
    warn "Ollama is running but bound to localhost only — the Docker app can't reach it."
    if ask "Restart Ollama bound to 0.0.0.0 so chklst (in Docker) can reach it?"; then
      pkill -f "ollama serve" 2>/dev/null || true
      sleep 1
    else
      ok "Leaving Ollama as-is (fine if you run chklst outside Docker)."
      return
    fi
  fi
  info "Starting Ollama service (bound to 0.0.0.0 so the Docker app can reach it) ..."
  # Bind to all interfaces: chklst runs in Docker and reaches the host via the
  # bridge IP, which a default 127.0.0.1-only bind would refuse.
  OLLAMA_HOST="0.0.0.0:11434" nohup ollama serve >/tmp/ollama.log 2>&1 &
  for _ in $(seq 1 30); do
    if curl -fsS "$OLLAMA_HOST/api/version" >/dev/null 2>&1; then
      ok "Ollama service started."
      return
    fi
    sleep 1
  done
  err "Ollama service did not come up. Check /tmp/ollama.log"; exit 1
}

# ---- 4. ensure the model -------------------------------------------------
ensure_model() {
  if ollama list 2>/dev/null | awk '{print $1}' | grep -qx "$MODEL"; then
    ok "Model $MODEL already present."
    return
  fi
  info "Pulling $MODEL (several GB, one-time). This can take a while ..."
  ollama pull "$MODEL"
  ok "Model $MODEL pulled."
}

# ---- 5. verify -----------------------------------------------------------
verify() {
  info "Running a test generation ..."
  local out
  out="$(ollama run "$MODEL" "Reply with exactly the word: READY" 2>/dev/null || true)"
  if printf '%s' "$out" | grep -qi "ready"; then
    ok "Gemma is responding. AI is ready."
  else
    warn "Model ran but response was unexpected: ${out:-<empty>}"
    warn "It still may work for chklst; check the app's Settings > Test AI."
  fi
}

main() {
  printf "%s== chklst AI setup (Ollama + %s) ==%s\n" "$c_blue" "$MODEL" "$c_reset"
  detect_os
  ensure_ollama
  ensure_running
  ensure_model
  verify
  echo
  ok "Done. In chklst Settings, set Ollama URL to '$OLLAMA_HOST' and model to '$MODEL', then enable AI."
  warn "Note: with no discrete GPU, generation runs on CPU and may take 1-2 minutes per request."
}

main "$@"
