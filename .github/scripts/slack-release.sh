#!/usr/bin/env bash
# Builds the Slack Block Kit payload for a stable release and prints it to
# stdout. Inputs (env): NAME, TAG, URL, BODY (GitHub release notes markdown),
# MASCOT_URL (public image; PNG/JPG/GIF). The workflow pipes the output to the
# Slack webhook; run it locally to preview the JSON.
set -euo pipefail

# GitHub markdown → Slack mrkdwn: drop the mascot <img> and the
# "# [x.y.z](compare) (date)" title,
# "### Features" → "*Features*", "* item" → "• item", "**b**" → "*b*",
# "[text](url)" → "<url|text>", then squeeze blank lines.
notes=$(printf '%s\n' "$BODY" | sed -E \
  -e '/^<img /d' \
  -e '/^# /d' \
  -e 's/^### (.*)$/*\1*/' \
  -e 's/^\* /• /' \
  -e 's/\*\*([^*]+)\*\*/*\1*/g' \
  -e 's/\[([^]]+)\]\(([^)]+)\)/<\2|\1>/g' | cat -s | sed -e '/./,$!d')

# Section text is capped at 3000 characters by Slack.
if [ "${#notes}" -gt 2900 ]; then
  notes="${notes:0:2900}"$'\n…'
fi

puns=(
  "Your code, bungkus'd to go."
  "Bungkus satu! One release, to go."
  "Fresh from the kitchen, wrapped in banana leaf."
  "Extra rice, extra tests."
  "Wrap it up, ship it out. Bungkus!"
)
pun=${puns[RANDOM % ${#puns[@]}]}

jq -n \
  --arg name "$NAME" --arg tag "$TAG" --arg url "$URL" \
  --arg notes "$notes" --arg pun "$pun" --arg mascot "$MASCOT_URL" '
{
  text: "\($name) \($tag) is out",
  blocks: [
    {
      type: "section",
      text: { type: "mrkdwn", text: "*\($name) \($tag) is out* 🍃\n_\($pun)_" },
      accessory: { type: "image", image_url: $mascot, alt_text: "bungkus mascot" }
    },
    { type: "divider" },
    { type: "section", text: { type: "mrkdwn", text: $notes } },
    {
      type: "context",
      elements: [
        { type: "mrkdwn", text: "<\($url)|View release> · update with `\($name) update`" }
      ]
    }
  ]
}'
