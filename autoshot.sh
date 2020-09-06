#!/bin/bash
# Proof of concept automatically running & exiting Factorio to take a
# mapshot without impacting the save or state of mods.

datadir="$HOME/factorio"
factorio="$HOME/factorio/bin/x64/factorio"
name="${1?}"

savefile="${datadir}/saves/${name?}.zip"
echo "Savefile: " "${savefile}"
if [ ! -f "${savefile?}" ]; then
    echo "Unknown file"
    exit 1
fi

workdir="$(mktemp -d mapshot.XXXXXXXXXX --tmpdir)"
echo "Workdir: " "${workdir}"

donefile="${datadir}/script-output/mapshot-done"

# Make a copy of the save game, to avoid overwriting by mistake.
rsync -a "${savefile}" "${workdir}/${name}.zip"

# Copy the mod structure to be able to fiddle with it.
rsync -a "${datadir}/mods/" "${workdir}/mods/"

# Remove existing mapshot, if installed.
rm -rf "${workdir?}/mods/mapshot"
rm -rf "${workdir?}/mods/mapshot_"*

# Setup mapshot plugin.
./prep.sh "${workdir}/mods/mapshot"

# Enable mapshot plugin.
py=$(cat <<EOF
import sys, json
data = json.load(sys.stdin)
mods = []
for mod in data['mods']:
    if mod['name'] != 'mapshot':
        mods.append(mod)

mods.append({'name': 'mapshot', 'enabled': True})
data['mods'] = mods
print(json.dumps(data, indent=2))
EOF
)
mv "${workdir}/mods/mod-list.json" "${workdir}/mods/mod-list.copy.json"
cat "${workdir}/mods/mod-list.copy.json" | python3 -c "$py" > "${workdir}/mods/mod-list.json"

# Create settings
cat > "${workdir}/mods/mapshot/overrides.lua" <<EOF
return {
    onstartup = true,
    shotname = "${name}",
}
EOF

# Remove previously 'done' indicator - otherwise we will get the impression that
# things are done immediately.
rm -f "${donefile?}"

# Run factorio
"$factorio" \
    --disable-audio \
    --disable-prototype-history \
    --load-game "${workdir}/${name}.zip" \
    --mod-directory "${workdir}/mods" \
    --force-graphics-preset very-low &

PID=$!
echo PID: ${PID?}

# Wait for the 'done' file to be creating, indicating end of screenshot.
while [ ! -f "${donefile}" ]; do sleep 1; done

# Kill & Wait for factorio to terminate.
kill ${PID?}
wait ${PID?}