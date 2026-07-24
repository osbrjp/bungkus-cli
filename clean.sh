#!/usr/bin/env sh
# Remove the scaffolded my-app folder (dev cleanup). Pass a name to remove a
# different folder: ./clean.sh some-app
rm -rf "${1:-my-app}"
