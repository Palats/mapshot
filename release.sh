#!/bin/bash
./regen.sh
version=$(cat info.json | python3 -c "import sys, json; print(json.load(sys.stdin)['version'])")
name="mapshot_${version?}"
dest="tmp/${name?}"

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

cd tmp
zip -r "${name?}.zip" "${name?}"