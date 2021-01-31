import * as common from "./common";
import { LitElement, html, customElement, property } from 'lit-element';

@customElement('factorio-ticks')
export class FactorioTicks extends LitElement {
    @property({ type: Number })
    ticks: number = 0;

    render() {
        if (!this.ticks) {
            return html`unknown age`;
        }
        let days = Math.trunc(this.ticks / 25000);
        return html`<span>${days} game days</span>`;
    }
}

@customElement('mapshot-listing')
export class MapshotListing extends LitElement {
    @property({ type: Object })
    shots: common.ShotsJSON | undefined;

    render() {
        if (!this.shots || !this.shots.all) {
            return html`No mapshots have been found. Create some and re-start mapshot server.`;
        }
        return html`
            <ul>
                ${this.shots.all.map((save) => html`
                    <li>${save.savename}<ul>
                        ${save.versions.map((si) => html`
                        <li><a href="?path=${si.path}">
                            <factorio-ticks .ticks=${si.ticks_played}></factorio-ticks>
                        </a></li>`)}
                    </ul></li>
                `)}
            </ul>
        `;
    }
}