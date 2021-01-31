import * as common from "./common";

export function run(config: common.MapshotConfig, info: common.ShotsJSON) {
    let root = document.getElementById("content");
    if (!root) {
        console.log("failed to find root element");
        return;
    }
    root.textContent = '';

    if (!info.all.length) {
        root.appendChild(document.createTextNode("No mapshots have been found. Create some and re-start mapshot server."));
        return;
    }

    let ul = root.appendChild(document.createElement("ul"));
    for (let si of info.all) {
        let li = ul.appendChild(document.createElement("li"));
        let a = li.appendChild(document.createElement("a"));
        a.href = `?path=${si.path}`;
        a.appendChild(document.createTextNode(si.name));
    }
}