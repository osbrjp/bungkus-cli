#!/usr/bin/env sh
# Remove the scaffolded my-app folder (dev cleanup). Pass a name to remove a
# different folder: ./clean.sh some-app
target="${1:-my-app}"
if [ -e "$target" ]; then
  rm -rf "$target"
  echo "Removed $target"
else
  echo "Nothing to remove: $target does not exist"
fi
