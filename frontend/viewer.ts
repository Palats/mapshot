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

    :root {
    --first: #000;
    --second: #fff;
    --third: #292929;
    --link-cursor: pointer;
    }

    * {
        color: var(--second);
        font-family: DejaVu Sans, sans-serif;
        user-select: none;
    }

    body {
        background: var(--first);
    }

    #content {
        height: 100%;
    }

    .leaflet-control-boxzoom,
    .leaflet-control-zoomslider {
        display: none;
    }

    #menu {
        position: fixed;
        top: 0;
        left: 0;
        background: #252425;
        z-index: 1000;
    }

    #menu div {
        background: #161516;
    }

    .leaflet-control.leaflet-control-layers {
        background: #252425;
        border-radius: 0;
        border: 1px solid var(--third);
        padding: 0;
    }

    input[type="radio"] {
        display: none;
    }

    span {
        text-transform: capitalize;
    }

    .leaflet-control-layers-separator {
        display: none;
    }

    .leaflet-control-layers-list::before {
        content: "Surfaces";
        font-size: 15px;
    }

    #planets::before {
        content: "Planets";
    }

    #platforms::before {
        content: "Platforms";
    }

    label,
    .leaflet-control-layers-list::before,
    #planets::before,
    #platforms::before {
        border-bottom: 1px solid #161516;
        padding: 2px 7px;
        display: flex !important;
        align-items: center;
    }

    label div,
    label span {
        display: flex;
        width: 100%;
        gap: 5px;
    }

    .leaflet-control-layers-overlays label,
    .leaflet-control-layers-overlays div {
        width: 30px;
        height: 30px;
        align-items: center;
        justify-content: center;
        padding: 0;
        margin: 0;
    }

    .leaflet-control-layers-overlays img,
    .leaflet-control-layers-overlays span {
        align-items: center;
        justify-content: center;
        display: flex;
    }

    label:hover {
        background: #ffc75a;
        cursor: var(--link-cursor);
    }

    label:hover span {
        color: var(--first);
    }

    .leaflet-right {
        left: 0 !important;
        right: unset !important;
    }

    .leaflet-control-layers-overlays {
        position: fixed;
        right: 0;
        top: 0;
        background: #252425;
        box-shadow: 0 0 10px var(--first);
        /* border: 2px solid var(--first); */
    }

    /* .leaflet-control-layers-base label:nth-child(6)::before {
        content: "Platforms";
    } */

    .leaflet-control-layers-list::before,
    #planets::before,
    #platforms::before {
        color: #f8ecc1;
    }

    .leaflet-touch .leaflet-control-layers,
    .leaflet-touch .leaflet-bar {
        border: none;
    }

    .leaflet-control-layers-list {
        box-shadow: 0 0 10px var(--first);
        /* border: 2px solid var(--first); */
        outline: none;
        width: 200px;
    }

    .leaflet-left .leaflet-control,
    .leaflet-top .leaflet-control {
        margin: 0 !important;
    }

    label.active {
        background: #ffc75a;
    }

    label.active span {
        color: var(--first);
    }

    .icon {
        width: 14px;
        height: 14px;
    }

    .leaflet-tooltip {
        display: flex;
        align-items: center;
        gap: 5px;
        background: none;
        border: none;
        color: var(--second);
        box-shadow: none;
        text-shadow:
            -2px -2px 0 black,
            2px -2px 0 black,
            -2px 2px 0 black,
            2px 2px 0 black;
    }

    .leaflet-tooltip:before {
        display: none;
    }

    .layer-toggle {
        display: flex;
        align-items: center;
        background: #333;
        color: white;
        border: none;
        padding: 5px 10px;
        margin: 5px;
        cursor: pointer;
    }

    .layer-toggle.active {
        background: #ffc75a;
    }

    .layer-icon {
        width: 20px;
        object-fit: contain;
        height: 20px;
    }

    input[type="checkbox"] {
        display: none;
    }

    .leaflet-bottom.leaflet-right {
        display: none;
    }

    .active span div {
        background: #ffc75a;
    }

    label:has(span.active) {
        background: #ffc75a;
    }

    label:has(input[type="checkbox"]:checked) {
        background: #ffc75a;
    }

    span.active {
        color: var(--first);
    }

    #download {
        position: fixed;
        bottom: 20px;
        font-size: 40px;
        right: 20px;
        z-index: 1000;
        border-radius: 100px;
        cursor: var(--link-cursor);
        background: var(--third);
        transition: 0.3s;
    }

    #download:hover {
        filter: brightness(0.9);
    }

    @media screen and (max-width: 768px) {
        .leaflet-control-layers-list{
            width: 120px;
        }
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

        // .fallback comes from leaflet.tilelayer.fallback, which does not have types.
        this.baseLayer = (L.tileLayer as any).fallback(config.encoded_path + si.file_prefix + "{z}/tile_{x}_{y}.jpg", {
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
            for (const station of si.stations as Iterable<common.FactorioStation>) {
                let tooltipContent = replaceItemAndFluidTags(station.backer_name);
                stationsLayers.push(
                    L.marker(
                        this.midPointToLatLng(station.bounding_box),
                        {
                            title: station.backer_name,
                            icon: createCustomIcon("https://raw.githubusercontent.com/ungaul/factorio-resources/main/bonus/station.png")
                        }
                    ).bindTooltip(tooltipContent, { permanent: true })
                );
            }
        }
        this.trainLayer = L.layerGroup(stationsLayers);

        // Layer: tags
        let tagsLayers = [];
        if (common.isIterable(si.tags)) {
            for (const tag of si.tags as Iterable<common.FactorioTag>) {
                let tooltipContent = replaceItemAndFluidTags(tag.text);
                tagsLayers.push(
                    L.marker(
                        this.worldToLatLng(tag.position.x, tag.position.y),
                        {
                            title: `${tag.force_name}: ${tag.text}`,
                            icon: createCustomIcon("https://raw.githubusercontent.com/ungaul/factorio-resources/main/bonus/tag.png")
                        }
                    ).bindTooltip(tooltipContent, { permanent: true })
                );
            }
        }
        this.tagsLayer = L.layerGroup(tagsLayers);

        // Layer: debug
        const debugLayers = [
            L.marker([0, 0], { title: "Start", icon: createCustomIcon("https://raw.githubusercontent.com/ungaul/factorio-resources/main/bonus/debug.png") }).bindPopup("Starting point"),
        ];

        if (common.isIterable(si.players)) {
            for (const player of si.players as Iterable<common.FactorioPlayer>) {
                debugLayers.push(
                    L.marker(
                        this.worldToLatLng(player.position.x, player.position.y),
                        {
                            title: player.name,
                            alt: `Player: ${player.name}`,
                            icon: createCustomIcon("https://raw.githubusercontent.com/ungaul/factorio-resources/main/bonus/debug.png")
                        }
                    ).bindTooltip(player.name, { permanent: true })
                );
            }
        }
        if (si.player) {
            debugLayers.push(
                L.marker(this.worldToLatLng(si.player.x, si.player.y), { title: "Player", icon: createCustomIcon("https://raw.githubusercontent.com/ungaul/factorio-resources/main/bonus/debug.png") }).bindPopup("Player")
            );
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

function createCustomIcon(url: string, size = [16, 16]): L.Icon {
    return L.icon({
        iconUrl: url,
        iconSize: size,
        iconAnchor: [0, 5],
        popupAnchor: [0, 0]
    });
}

function replaceItemAndFluidTags(text: string): string {
    return text.replace(/\[(item|fluid|recipe)=(.*?)\]/g,
        (match, type, name) =>
            `<img class="icon" src="https://raw.githubusercontent.com/ungaul/factorio-resources/main/graphics/icons/64x64/${name}.png" alt="${name}" title="${name}">`
    );
}

function run(config: common.MapshotConfig, info: common.MapshotJSON) {
    const layerControl = L.control.layers();

    interface SurfaceCategory {
        layer: L.TileLayer;
        name: string;
        className: string;
        category: string;
    }
    const surfaces: Surface[] = [];
    const surfaceByKey: Map<string, Surface> = new Map();
    // Collect categorized surfaces for planet/platform UI
    const categories: { planets: SurfaceCategory[]; platforms: SurfaceCategory[]; layers: SurfaceCategory[] } = {
        planets: [],
        platforms: [],
        layers: []
    };

    // populate categories and surfaces
    for (const si of info.surfaces) {
        const s = new Surface(config, si);
        surfaces.push(s);

        // Categorize by surface name
        let cat: "planets" | "platforms" | "layers" = "planets";
        let className = "planet";
        let category = "#planets";
        if (si.surface_name.startsWith("platform-")) {
            cat = "platforms";
            className = "platform";
            category = "#platforms";
        }

        categories[cat].push({
            layer: s.baseLayer,
            name: si.surface_name,
            className,
            category
        });

        surfaceByKey.set(s.surfaceInfo.surface_idx.toString(), s);
        surfaceByKey.set(s.surfaceInfo.surface_name, s);
    }

    // Sorting helper
    function customSort(array: SurfaceCategory[]) {
        return array.sort((a, b) => {
            const [nameA, nameB] = [a.name.toLowerCase(), b.name.toLowerCase()];
            const [matchA, matchB] = [nameA.match(/^platform-(\d+)$/), nameB.match(/^platform-(\d+)$/)];

            if (!matchA && matchB) return -1;
            if (matchA && !matchB) return 1;
            if (nameA === "vulcanus") return 1;
            if (nameB === "vulcanus") return -1;
            return matchA && matchB ? Number(matchA[1]) - Number(matchB[1]) : nameA.localeCompare(nameB);
        });
    }

    function addCategory(categoryId: keyof typeof categories, maps: SurfaceCategory[]) {
        customSort(maps).forEach(item => {
            let icon = categoryId === "planets"
                ? item.name
                : "cargo-pod";
            let displayName =
                `<span class="${item.className}">
                <img src="https://raw.githubusercontent.com/ungaul/factorio-resources/main/graphics/icons/64x64/${icon}.png"
                     width="16" height="16" style="vertical-align: middle; margin-right: 5px;"
                     onerror="this.onerror=null;this.src='https://raw.githubusercontent.com/ungaul/factorio-resources/main/graphics/icons/64x64/cargo-pod.png';">
                ${item.name}
            </span>`;
            layerControl.addBaseLayer(item.layer, displayName);
        });
    }
    addCategory("planets", categories.planets);
    addCategory("platforms", categories.platforms);

    const trainLayer = L.layerGroup();
    const tagsLayer = L.layerGroup();
    const debugLayer = L.layerGroup();
    layerControl.addOverlay(
        trainLayer,
        "<img class='layer-icon' src='https://raw.githubusercontent.com/ungaul/factorio-resources/main/bonus/station.png'>"
    );
    layerControl.addOverlay(
        tagsLayer,
        "<img class='layer-icon' src='https://raw.githubusercontent.com/ungaul/factorio-resources/main/bonus/tag.png'>"
    );
    layerControl.addOverlay(
        debugLayer,
        "<img class='layer-icon' src='https://raw.githubusercontent.com/ungaul/factorio-resources/main/bonus/debug.png'>"
    );
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

    setTimeout(() => {
        document.querySelectorAll(".planet, .platform").forEach(item => {
            item.addEventListener("click", (ev) => {
                document.querySelectorAll(".planet, .platform").forEach(el => el.classList.remove("active"));
                (ev.currentTarget as HTMLElement).classList.add("active");
            });
        });
    }, 500);

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

    let minBounds = currentSurface.worldToLatLng(currentSurface.surfaceInfo.world_min.x, currentSurface.surfaceInfo.world_min.y);
    let maxBounds = currentSurface.worldToLatLng(currentSurface.surfaceInfo.world_max.x, currentSurface.surfaceInfo.world_max.y);
    let bounds = L.latLngBounds(minBounds, maxBounds);
    mymap.setMaxBounds(bounds);
    (mymap as any).options.maxBoundsViscosity = 1.0;

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

    fetch(config.encoded_path + 'mapshot.json')
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

    config.encoded_path = params.get("path") ?? config.encoded_path ?? "";
    if (!!config.encoded_path && config.encoded_path[config.encoded_path.length - 1] != "/") {
        config.encoded_path = config.encoded_path + "/";
    }
    load(config);
}