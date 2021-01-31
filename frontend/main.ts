import * as common from "./common";
import * as viewer from "./viewer";
import * as listing from "./listing";
import { html, render } from 'lit-html';

common.globalCSS(`
    html,body {
        margin: 0;
    }
`);

// The name of the path to use by default is injected in the HTML.
declare var MAPSHOT_CONFIG: common.MapshotConfig;

const config = JSON.parse(JSON.stringify(MAPSHOT_CONFIG ?? {}));

const params = new URLSearchParams(window.location.search);
config.path = params.get("path") ?? config.path ?? "";
if (!!config.path && config.path[config.path.length - 1] != "/") {
    config.path = config.path + "/";
}
console.log("Config", config);

function runViewer() {
    fetch(config.path + 'mapshot.json')
        .then(resp => resp.json())
        .then((info: common.MapshotJSON) => {
            console.log("Map info", info);
            viewer.run(config, info);
        });
}

function runListing() {
    fetch('shots.json')
        .then(resp => resp.json())
        .then((shots: common.ShotsJSON) => {
            render(html`<mapshot-listing .info=${shots}>foo</mapshot-listing>`, document.body);
        });
}

listing.MapshotListing;

if (config.path) {
    runViewer();
} else {
    runListing();
}