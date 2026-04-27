#!/usr/bin/env bash
# Re-runs install.sh against the latest GitHub Release, replacing the binary in place.
# Once the binary embeds its own version (--version), this can short-circuit when
# already current; for now, always fetches the latest.
set -euo pipefail

REPO="spencer-osbrjp/bungkus-cli"
exec curl -fsSL "https://raw.githubusercontent.com/${REPO}/main/install.sh" | bash
