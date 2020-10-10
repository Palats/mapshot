// The name of the path to use by default is injected in the HTML.
declare var MAPSHOT_DEFAULT_PATH: string;

const params = new URLSearchParams(window.location.search);
let path = params.get("path") ?? MAPSHOT_DEFAULT_PATH;
if (!!path && path[path.length - 1] != "/") {
    path = path + "/";
}
console.log("Path", path);

fetch(path + 'mapshot.json')
    .then(resp => resp.json())
    .then(info => {
        console.log("Map info", info);

        const iterable = function (obj: any): Boolean {
            // falsy value is javascript includes empty string, which is iterable,
            // so we cannot just check if the value is truthy.
            if (obj === null || obj === undefined) {
                return false;
            }
            return typeof obj[Symbol.iterator] === "function";
        }

        const worldToLatLng = function (x: number, y: number) {
            const ratio = info.render_size / info.tile_size;
            return L.latLng(
                -y * ratio,
                x * ratio
            );
        };

        const midPointToLatLng = function (bbox: any) {
            return worldToLatLng(
                (bbox.left_top.x + bbox.right_bottom.x) / 2,
                (bbox.left_top.y + bbox.right_bottom.y) / 2,
            );
        }

        const baseLayer = L.tileLayer(path + "zoom_{z}/tile_{x}_{y}.jpg", {
            tileSize: info.render_size,
            bounds: L.latLngBounds(
                worldToLatLng(info.world_min.x, info.world_min.y),
                worldToLatLng(info.world_max.x, info.world_max.y),
            ),
            noWrap: true,
            maxNativeZoom: info.zoom_max,
            minNativeZoom: info.zoom_min,
            minZoom: info.zoom_min - 4,
            maxZoom: info.zoom_max + 4,
        });

        const debugLayer = L.layerGroup([
            L.marker([0, 0], { title: "Start" }).bindPopup("Starting point"),
            L.marker(worldToLatLng(info.player.x, info.player.y), { title: "Player" }).bindPopup("Player"),
            L.marker(worldToLatLng(info.world_min.x, info.world_min.y), { title: `${info.world_min.x}, ${info.world_min.y}` }),
            L.marker(worldToLatLng(info.world_min.x, info.world_max.y), { title: `${info.world_min.x}, ${info.world_max.y}` }),
            L.marker(worldToLatLng(info.world_max.x, info.world_min.y), { title: `${info.world_max.x}, ${info.world_min.y}` }),
            L.marker(worldToLatLng(info.world_max.x, info.world_max.y), { title: `${info.world_max.x}, ${info.world_max.y}` }),
        ]);

        let stations = [];
        if (iterable(info.stations)) {
            for (const station of info.stations) {
                stations.push(L.marker(
                    midPointToLatLng(station.bounding_box),
                    { title: station.backer_name },
                ).bindTooltip(station.backer_name, { permanent: true }))
            }
        }
        const stationsLayer = L.layerGroup(stations);

        let tags = [];
        if (iterable(info.tags)) {
            for (const tag of info.tags) {
                tags.push(L.marker(
                    worldToLatLng(tag.position.x, tag.position.y),
                    { title: `${tag.force_name}: ${tag.text}` },
                ).bindTooltip(tag.text, { permanent: true }))
            }
        }
        const tagsLayer = L.layerGroup(tags);

        const mymap = L.map('map', {
            crs: L.CRS.Simple,
            layers: [baseLayer],
        });

        L.control.layers({/* Only one default base layer */ }, {
            "Train stations": stationsLayer,
            "Tags": tagsLayer,
            "Debug": debugLayer,
        }).addTo(mymap);

        mymap.setView([0, 0], 0);
    });
