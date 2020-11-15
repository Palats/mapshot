import * as common from "./common";
import * as viewer from "./viewer";

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

runViewer();