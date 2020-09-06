#!/bin/bash
# Copy the files necessary for the mod.

dest=${1?}

./regen.sh

mkdir -p "${dest?}"
cp \
    changelog.txt \
    control.lua \
    generated.lua \
    info.json \
    LICENSE \
    README.md \
    settings.lua \
    thumbnail.png \
    "${dest?}"