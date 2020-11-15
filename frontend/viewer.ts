import * as L from "leaflet";
import "./boxzoom/leaflet-control-boxzoom-src.js";
import "./zoomslider/L.Control.Zoomslider.js";

import boxzoom_svg from "./boxzoom/leaflet-control-boxzoom.svg";
import "./boxzoom/leaflet-control-boxzoom.css";
import "./zoomslider/L.Control.Zoomslider.css";

import * as common from "./common";

common.globalCSS(`
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

export function run(config: common.MapshotConfig, info: common.MapshotJSON) {
    const worldToLatLng = function (x: number, y: number) {
        const ratio = info.render_size / info.tile_size;
        return L.latLng(
            -y * ratio,
            x * ratio
        );
    };

    const latLngToWorld = function (l: L.LatLng) {
        const ratio = info.tile_size / info.render_size;
        return {
            x: l.lng * ratio,
            y: -l.lat * ratio,
        }
    }

    const midPointToLatLng = function (bbox: common.FactorioBoundingBox) {
        return worldToLatLng(
            (bbox.left_top.x + bbox.right_bottom.x) / 2,
            (bbox.left_top.y + bbox.right_bottom.y) / 2,
        );
    }

    const baseLayer = L.tileLayer(config.path + "zoom_{z}/tile_{x}_{y}.jpg", {
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

    const mymap = L.map('map', {
        crs: L.CRS.Simple,
        layers: [baseLayer],
        zoomSnap: 0.1,
        zoomsliderControl: true,
        zoomControl: false,
        zoomDelta: 1.0,
    });
    const layerControl = L.control.layers().addTo(mymap);

    const layerKeys = new Map<L.Layer, string>();
    const registerLayer = function (key: string, name: string, layer: L.Layer) {
        layerControl.addOverlay(layer, name);
        layerKeys.set(layer, key);
    }

    // Layer: train stations
    let stationsLayers = [];
    if (common.isIterable(info.stations)) {
        for (const station of info.stations) {
            stationsLayers.push(L.marker(
                midPointToLatLng(station.bounding_box),
                { title: station.backer_name },
            ).bindTooltip(station.backer_name, { permanent: true }))
        }
    }
    registerLayer("lt", "Train stations", L.layerGroup(stationsLayers));

    // Layer: tags
    let tagsLayers = [];
    if (common.isIterable(info.tags)) {
        for (const tag of info.tags) {
            tagsLayers.push(L.marker(
                worldToLatLng(tag.position.x, tag.position.y),
                { title: `${tag.force_name}: ${tag.text}` },
            ).bindTooltip(tag.text, { permanent: true }))
        }
    }
    registerLayer("lg", "Tags", L.layerGroup(tagsLayers));

    // Layer: debug
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
    registerLayer("ld", "Debug", L.layerGroup(debugLayers));

    // Add a control to zoom to a region.
    L.Control.boxzoom({
        position: 'topleft',
    }).addTo(mymap);

    // Set original view (position/zoom/layers).
    const queryParams = new URLSearchParams(window.location.search);
    let x = common.parseNumber(queryParams.get("x"), 0);
    let y = common.parseNumber(queryParams.get("y"), 0);
    let z = common.parseNumber(queryParams.get("z"), 0);
    mymap.setView(worldToLatLng(x, y), z);
    layerKeys.forEach((key, layer) => {
        const p = queryParams.get(key);
        if (p == "0") {
            mymap.removeLayer(layer);
        }
        if (p == "1") {
            mymap.addLayer(layer);
        }
    });

    // Update URL when position/view changes.
    const onViewChange = (e: L.LeafletEvent) => {
        const z = mymap.getZoom();
        const { x, y } = latLngToWorld(mymap.getCenter());
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
        const key = layerKeys.get(e.layer);
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
