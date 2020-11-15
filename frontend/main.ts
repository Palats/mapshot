import * as common from "./common";
import * as viewer from "./viewer";

const main_css = `
    html,body {
        margin: 0;
    }
    .with-background-image {
        background-image:url(${viewer.svg});
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
`;

var style = document.createElement('style');
style.innerHTML = main_css;
document.head.appendChild(style);

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

runViewer();