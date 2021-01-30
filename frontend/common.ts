// Config for mapshot UI.
export interface MapshotConfig {
    // Where to find the mapshot to load (not including `mapshot.json`).
    path?: string;
}

export interface FactorioPosition {
    x: number,
    y: number,
}

export interface FactorioBoundingBox {
    left_top: FactorioPosition,
    right_bottom: FactorioPosition,
}

export interface FactorioIcon {
    name: string,
    type: string,
}

export interface FactorioStation {
    backer_name: string,
    bounding_box: FactorioBoundingBox,
}

export interface FactorioTag {
    force_name: string,
    force_index: string,
    icon: FactorioIcon,
    tag_number: number,
    position: FactorioPosition,
    text: string,
}

export interface MapshotJSON {
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

export function parseNumber(v: any, defvalue: number): number {
    const c = Number(v);
    return isNaN(c) ? defvalue : c;
}

export function isIterable<T>(obj: Iterable<T> | any): obj is Iterable<T> {
    // falsy value is javascript includes empty string, which is iterable,
    // so we cannot just check if the value is truthy.
    if (obj === null || obj === undefined) {
        return false;
    }
    return typeof obj[Symbol.iterator] === "function";
}

export function globalCSS(css: string) {
    var style = document.createElement('style');
    style.innerHTML = css;
    document.head.appendChild(style);
}