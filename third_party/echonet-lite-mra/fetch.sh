#!/bin/sh
set -eu

release=MRA_v1.4.0
url="https://echonet.jp/wp/wp-content/uploads/pdf/General/Standard/MRA/$release.zip"
sha256=fbb58d12b2676b686e519cb48937022fc8e6476ec01e609553cc61720e54c6a2
base_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
data_dir="$base_dir/$release"

if [ -f "$data_dir/metaData.json" ]; then
    printf '%s\n' "$data_dir"
    exit 0
fi

temp_dir=$(mktemp -d "$base_dir/.fetch.XXXXXXXX")
trap 'rm -rf -- "$temp_dir"' EXIT
trap 'exit 1' HUP INT TERM

archive="$temp_dir/$release.zip"
curl --fail --location --silent --show-error --retry 2 --output "$archive" "$url"
if ! printf '%s  %s\n' "$sha256" "$archive" | shasum -a 256 --check --status; then
    printf 'SHA-256 verification failed for %s\n' "$url" >&2
    exit 1
fi

unzip -q "$archive" -d "$temp_dir"
mv -n -- "$temp_dir/$release" "$base_dir/"
if [ ! -f "$data_dir/metaData.json" ]; then
    printf 'Could not install MRA data at %s\n' "$data_dir" >&2
    exit 1
fi

printf '%s\n' "$data_dir"
