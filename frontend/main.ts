import * as L from "leaflet";
import "./boxzoom/leaflet-control-boxzoom-src.js";

import boxzoom_svg from "./boxzoom/leaflet-control-boxzoom.svg";
import "./boxzoom/leaflet-control-boxzoom.css";

const main_css = `
    html,body {
        margin: 0;
    }
    .with-background-image {
        background-image:url(${boxzoom_svg});
        background-repeat:no-repeat;
        background-size:21px 21px; /* 25px image, 25px box; subtract 2px for padding on every side = 21px rendering height */
        background-position:2px 2px;
    }
`;

var style = document.createElement('style');
style.innerHTML = main_css;
document.head.appendChild(style);

// The name of the path to use by default is injected in the HTML.
declare var MAPSHOT_CONFIG: MapshotConfig;

// Config for mapshot UI.
interface MapshotConfig {
    // Where to find the mapshot to load (not including `mapshot.json`).
    path?: string;
}

interface FactorioPosition {
    x: number,
    y: number,
}

interface FactorioBoundingBox {
    left_top: FactorioPosition,
    right_bottom: FactorioPosition,
}

interface FactorioIcon {
    name: string,
    type: string,
}

interface FactorioStation {
    backer_name: string,
    bounding_box: FactorioBoundingBox,
}

interface FactorioTag {
    force_name: string,
    force_index: string,
    icon: FactorioIcon,
    tag_number: number,
    position: FactorioPosition,
    text: string,
}

interface MapshotJSON {
    // A unique ID generated for this render.
    unique_id: string,
    // The name of the save - not reliable, as it can be customized.
    // This is mostly the subdir that was used.
    savename: string,

    // game.tick
    tick: number,
    // game.ticks_played
    ticks_played: number,
    // Seed of the map.
    seed: number,
    // Factorio map exchange string
    map_exchange?: string,
    // A short ID of the map, derived from map_exchange.
    map_id: string,

    // Size of a tile in in-game units for the least detailed layer.
    tile_size: number,
    // Size of a tile, in pixels.
    render_size: number,
    // Area rendered.
    world_min: FactorioPosition,
    world_max: FactorioPosition,
    // Minimal available zoom level index (least detailed)
    zoom_min: number,
    // Maximal available zoom level index (most detailed)
    zoom_max: number,

    // Current position of the player.
    player?: FactorioPosition,
    // List of train stations.
    stations?: FactorioStation[] | {},
    // List of map tags.
    tags?: FactorioTag[] | {},
}

const params = new URLSearchParams(window.location.search);
let path = params.get("path") ?? MAPSHOT_CONFIG.path ?? "";
if (!!path && path[path.length - 1] != "/") {
    path = path + "/";
}
console.log("Path", path);

fetch(path + 'mapshot.json')
    .then(resp => resp.json())
    .then((info: MapshotJSON) => {
        console.log("Map info", info);

        const isIterable = function <T>(obj: Iterable<T> | any): obj is Iterable<T> {
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

        const midPointToLatLng = function (bbox: FactorioBoundingBox) {
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

        const debugLayers = [
            L.marker([0, 0], { title: "Start" }).bindPopup("Starting point"),
        ]
        if (info.player) {
            debugLayers.push(L.marker(worldToLatLng(info.player.x, info.player.y), { title: "Player" }).bindPopup("Player"))
        }
        debugLayers.push(
            L.marker(worldToLatLng(info.world_min.x, info.world_min.y), { title: `${info.world_min.x}, ${info.world_min.y}` }),
            L.marker(worldToLatLng(info.world_min.x, info.world_max.y), { title: `${info.world_min.x}, ${info.world_max.y}` }),
            L.marker(worldToLatLng(info.world_max.x, info.world_min.y), { title: `${info.world_max.x}, ${info.world_min.y}` }),
            L.marker(worldToLatLng(info.world_max.x, info.world_max.y), { title: `${info.world_max.x}, ${info.world_max.y}` }),
        );

        let stationsLayers = [];
        if (isIterable(info.stations)) {
            for (const station of info.stations) {
                stationsLayers.push(L.marker(
                    midPointToLatLng(station.bounding_box),
                    { title: station.backer_name },
                ).bindTooltip(station.backer_name, { permanent: true }))
            }
        }

        let tagsLayers = [];
        if (isIterable(info.tags)) {
            for (const tag of info.tags) {
                tagsLayers.push(L.marker(
                    worldToLatLng(tag.position.x, tag.position.y),
                    { title: `${tag.force_name}: ${tag.text}` },
                ).bindTooltip(tag.text, { permanent: true }))
            }
        }

        const mymap = L.map('map', {
            crs: L.CRS.Simple,
            layers: [baseLayer],
            zoomSnap: 0.1,
        });
        L.control.layers({/* Only one default base layer */ }, {
            "Train stations": L.layerGroup(stationsLayers),
            "Tags": L.layerGroup(tagsLayers),
            "Debug": L.layerGroup(debugLayers),
        }).addTo(mymap);

        L.Control.boxzoom({
            position: 'topleft',
        }).addTo(mymap);

        mymap.setView([0, 0], 0);
    });
