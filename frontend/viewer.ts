import * as L from "leaflet";
import "./boxzoom/leaflet-control-boxzoom-src.js";
import "./zoomslider/L.Control.Zoomslider.js";

import boxzoom_svg from "./boxzoom/leaflet-control-boxzoom.svg";
import "./boxzoom/leaflet-control-boxzoom.css";
import "./zoomslider/L.Control.Zoomslider.css";

import * as common from "./common";

common.globalCSS(`
    html,body {
        margin: 0;
    }

    .with-background-image {
        background-image:url(${boxzoom_svg});
        background-size:22px 22px;
        background-position:4px 4px;
    }
    .leaflet-touch .leaflet-control-zoomslider {
        border: none;
    }
    .leaflet-control-boxzoom {
        border:none;
        width:30px;
        height:30px;
    }
`);

/*
 * Workaround for 1px lines appearing in some browsers due to fractional transforms
 * and resulting anti-aliasing.
 * https://github.com/Leaflet/Leaflet/issues/3575#issuecomment-150544739
 */
function leafletHack() {
    const originalInitTile = (L.GridLayer.prototype as any)._initTile;
    L.GridLayer.include({
        _initTile: function (tile: any) {
            originalInitTile.call(this, tile);
            var tileSize = this.getTileSize();
            tile.style.width = tileSize.x + 1 + 'px';
            tile.style.height = tileSize.y + 1 + 'px';
        }
    });
}
leafletHack();

class Surface {
    surfaceInfo: common.MapshotSurfaceJSON;
    baseLayer: L.TileLayer;
    trainLayer: L.LayerGroup;
    tagsLayer: L.LayerGroup;
    debugLayer: L.LayerGroup;

    constructor(config: common.MapshotConfig, si: common.MapshotSurfaceJSON) {
        this.surfaceInfo = si;

        this.baseLayer = L.tileLayer(config.path + si.file_prefix + "{z}/tile_{x}_{y}.jpg", {
            tileSize: si.render_size,
            bounds: L.latLngBounds(
                this.worldToLatLng(si.world_min.x, si.world_min.y),
                this.worldToLatLng(si.world_max.x, si.world_max.y),
            ),
            noWrap: true,
            maxNativeZoom: si.zoom_max,
            minNativeZoom: si.zoom_min,
            minZoom: si.zoom_min - 4,
            maxZoom: si.zoom_max + 4,
        });

        // Layer: train stations
        let stationsLayers = [];
        if (common.isIterable(si.stations)) {
            for (const station of si.stations) {
                stationsLayers.push(L.marker(
                    this.midPointToLatLng(station.bounding_box),
                    { title: station.backer_name },
                ).bindTooltip(station.backer_name, { permanent: true }))
            }
        }
        this.trainLayer = L.layerGroup(stationsLayers);

        // Layer: tags
        let tagsLayers = [];
        if (common.isIterable(si.tags)) {
            for (const tag of si.tags) {
                tagsLayers.push(L.marker(
                    this.worldToLatLng(tag.position.x, tag.position.y),
                    { title: `${tag.force_name}: ${tag.text}` },
                ).bindTooltip(tag.text, { permanent: true }))
            }
        }
        this.tagsLayer = L.layerGroup(tagsLayers);

        // Layer: debug
        const debugLayers = [
            L.marker([0, 0], { title: "Start" }).bindPopup("Starting point"),
        ]
        if (common.isIterable(si.players)) {
            for (const player of si.players) {
                debugLayers.push(
                    L.marker(
                        this.worldToLatLng(player.position.x, player.position.y),
                        {
                            title: player.name,
                            alt: `Player: ${player.name}`
                        },
                    ).bindTooltip(player.name, {
                        permanent: true
                    })
                )
            }
        }
        if (si.player) {
            debugLayers.push(L.marker(this.worldToLatLng(si.player.x, si.player.y), { title: "Player" }).bindPopup("Player"));
        }
        debugLayers.push(
            L.marker(this.worldToLatLng(si.world_min.x, si.world_min.y), { title: `${si.world_min.x}, ${si.world_min.y}` }),
            L.marker(this.worldToLatLng(si.world_min.x, si.world_max.y), { title: `${si.world_min.x}, ${si.world_max.y}` }),
            L.marker(this.worldToLatLng(si.world_max.x, si.world_min.y), { title: `${si.world_max.x}, ${si.world_min.y}` }),
            L.marker(this.worldToLatLng(si.world_max.x, si.world_max.y), { title: `${si.world_max.x}, ${si.world_max.y}` }),
        );
        this.debugLayer = L.layerGroup(debugLayers);
    }

    worldToLatLng(x: number, y: number) {
        const ratio = this.surfaceInfo.render_size / this.surfaceInfo.tile_size;
        return L.latLng(
            -y * ratio,
            x * ratio
        );
    };

    latLngToWorld(l: L.LatLng) {
        const ratio = this.surfaceInfo.tile_size / this.surfaceInfo.render_size;
        return {
            x: l.lng * ratio,
            y: -l.lat * ratio,
        }
    }

    midPointToLatLng(bbox: common.FactorioBoundingBox) {
        return this.worldToLatLng(
            (bbox.left_top.x + bbox.right_bottom.x) / 2,
            (bbox.left_top.y + bbox.right_bottom.y) / 2,
        );
    }

}

function run(config: common.MapshotConfig, info: common.MapshotJSON) {
    const layerControl = L.control.layers();

    const surfaces: Surface[] = [];
    const surfaceByKey: Map<string, Surface> = new Map();
    for (const si of info.surfaces) {
        const s = new Surface(config, si);
        surfaces.push(s);
        layerControl.addBaseLayer(s.baseLayer, si.surface_name);
        surfaceByKey.set(s.surfaceInfo.surface_idx.toString(), s);
        surfaceByKey.set(s.surfaceInfo.surface_name, s);
    }

    const trainLayer = L.layerGroup();
    const tagsLayer = L.layerGroup();
    const debugLayer = L.layerGroup();
    layerControl.addOverlay(trainLayer, "Train stations");
    layerControl.addOverlay(tagsLayer, "Tags");
    layerControl.addOverlay(debugLayer, "Debug");
    const overlayKeys = new Map<L.Layer, string>();
    overlayKeys.set(trainLayer, "lt");
    overlayKeys.set(tagsLayer, "lg");
    overlayKeys.set(debugLayer, "ld");

    const updateOverlays = (s: Surface) => {
        trainLayer.clearLayers();
        trainLayer.addLayer(s.trainLayer);
        tagsLayer.clearLayers();
        tagsLayer.addLayer(s.tagsLayer);
        debugLayer.clearLayers();
        debugLayer.addLayer(s.debugLayer);
    }

    const mymap = L.map('content', {
        crs: L.CRS.Simple,
        layers: [],
        zoomSnap: 0.1,
        zoomsliderControl: true,
        zoomControl: false,
        zoomDelta: 1.0,
    });
    layerControl.addTo(mymap);

    // Add a control to zoom to a region.
    L.Control.boxzoom({
        position: 'topleft',
    }).addTo(mymap);

    // Set original view (position/zoom/layers).
    const queryParams = new URLSearchParams(window.location.search);
    let currentSurface = surfaceByKey.get(queryParams.get("s") ?? "1") ?? surfaces[0];
    mymap.addLayer(currentSurface.baseLayer);
    updateOverlays(currentSurface);

    let x = common.parseNumber(queryParams.get("x"), 0);
    let y = common.parseNumber(queryParams.get("y"), 0);
    let z = common.parseNumber(queryParams.get("z"), 0);
    mymap.setView(currentSurface.worldToLatLng(x, y), z);
    overlayKeys.forEach((key, layer) => {
        const p = queryParams.get(key);
        if (p == "0") {
            mymap.removeLayer(layer);
        }
        if (p == "1") {
            mymap.addLayer(layer);
        }
    });

    // Update URL & current surface when base layer changed
    mymap.on('baselayerchange', (e: L.LayersControlEvent) => {
        const s = surfaceByKey.get(e.name);
        if (!s) {
            console.log("unknown layer", e.name);
            return;
        }
        currentSurface = s;
        updateOverlays(currentSurface);
        const queryParams = new URLSearchParams(window.location.search);
        queryParams.set("s", currentSurface.surfaceInfo.surface_idx.toString());
        history.replaceState(null, "", "?" + queryParams.toString());
    });

    // Update URL when position/view changes.
    const onViewChange = (e: L.LeafletEvent) => {
        const z = mymap.getZoom();
        const { x, y } = currentSurface.latLngToWorld(mymap.getCenter());
        const queryParams = new URLSearchParams(window.location.search);
        queryParams.set("x", x.toFixed(1));
        queryParams.set("y", y.toFixed(1));
        queryParams.set("z", z.toFixed(1));
        history.replaceState(null, "", "?" + queryParams.toString());
    }
    mymap.on('zoomend', onViewChange);
    mymap.on('moveend', onViewChange);
    mymap.on('resize', onViewChange);

    // Update URL when overlays are added/removed.
    const onLayerChange = (e: L.LayersControlEvent) => {
        const key = overlayKeys.get(e.layer);
        if (!key) {
            console.log("unknown layer", e.name);
            return;
        }
        const queryParams = new URLSearchParams(window.location.search);
        queryParams.set(key, e.type == "overlayadd" ? "1" : "0");
        history.replaceState(null, "", "?" + queryParams.toString());
    }
    mymap.on('overlayadd', onLayerChange);
    mymap.on('overlayremove', onLayerChange);
};


// ------ Bootstrap ------

function load(config: common.MapshotConfig) {
    console.log("Config", config);

    fetch(config.path + 'mapshot.json')
        .then(resp => resp.json())
        .then((info: common.MapshotJSON) => {
            // Backward compatibility - try to load mapshot.json from data before
            // support for multiple surfaces.

            if (info.surfaces === undefined) {
                // No surfaces defined, that's an old format.
                const raw = info as any;
                info.surfaces = [{
                    surface_name: "nauvis",
                    surface_idx: 1,
                    file_prefix: "zoom_",

                    tile_size: raw.tile_size,
                    render_size: raw.render_size,
                    world_min: raw.world_min,
                    world_max: raw.world_max,
                    zoom_min: raw.zoom_min,
                    zoom_max: raw.zoom_max,
                    player: raw.player,
                    players: raw.players,
                    stations: raw.stations,
                    tags: raw.tags,
                }];
            }

            console.log("Map info", info);
            run(config, info);
        });
}


// The name of the path to use by default is injected in the HTML.
declare var MAPSHOT_CONFIG: common.MapshotConfig;

const params = new URLSearchParams(window.location.search);
if (params.get("l")) {
    fetch("/latest/" + params.get("l"))
        .then(resp => resp.json())
        .then((config: common.MapshotConfig) => {
            load(config);
        });
} else {
    const config = JSON.parse(JSON.stringify(MAPSHOT_CONFIG ?? {}));

    config.path = params.get("path") ?? config.path ?? "";
    if (!!config.path && config.path[config.path.length - 1] != "/") {
        config.path = config.path + "/";
    }
    load(config);
}