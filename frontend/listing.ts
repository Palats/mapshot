import * as common from "./common";
import { LitElement, html, customElement, property } from 'lit-element';

@customElement('mapshot-listing')
export class MapshotListing extends LitElement {
    @property({ type: Object })
    info: common.ShotsJSON | undefined;

    render() {
        if (!this.info || !this.info.all) {
            return html`No mapshots have been found. Create some and re-start mapshot server.`;
        }
        return html`
            <ul>
                ${this.info.all.map((si) => html`<li><a href="?path=${si.path}">${si.name}</a></li>`)}
            </ul>
        `;
    }
}