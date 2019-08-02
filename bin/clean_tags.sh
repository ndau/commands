#!/usr/bin/env bash

datefmt="%Y-%m-%d"

# clean up tags which are more than 30 days old

unixify() {
    date -ju -v0H -v0M -v0S "-f$datefmt" "$1" +%s
}

today=$(date -ju "+$datefmt")
today_unix=$(unixify "$today")

for tag in $(git tag | grep -v '^v' | grep -v '^mainnet'); do
    tag_date=$(git log "$tag" -1 --pretty=format:"%ad" --date=short)
    tag_unix=$(unixify "$tag_date")
    age_days=$(((today_unix-tag_unix)/(24*60*60)))
    echo "$tag: $age_days days old"
    if [[ "$age_days" -gt 30 ]]; then
        git push origin ":$tag" >/dev/null 2>&1
        git tag -d "$tag" >/dev/null
    fi
done
