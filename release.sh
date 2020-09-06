#!/bin/bash

version=$(cat info.json | python3 -c "import sys, json; print(json.load(sys.stdin)['version'])")
name="mapshot_${version?}"
dest="tmp/${name?}"

./prep.sh "${dest?}"

cd tmp
zip -r "${name?}.zip" "${name?}"