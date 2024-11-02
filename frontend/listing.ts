import * as common from "./common";
import { LitElement, html, css, customElement, property } from 'lit-element';
import { render } from 'lit-html';

@customElement('factorio-ticks')
class FactorioTicks extends LitElement {
    @property({ type: Number })
    ticks: number = 0;

    render() {
        if (!this.ticks) {
            return html`Played: unknown`;
        }
        let gameDays = Math.trunc(this.ticks / 25000);
        let playedHours = Math.trunc(this.ticks / (60 * 3600));
        return html`<span>Played: ${playedHours} hours; Game: ${gameDays} days; Ticks: ${this.ticks}</span>`;
    }
}

@customElement('factorio-relticks')
class FactorioRelTicks extends LitElement {
    @property({ type: Number })
    ticks: number = 0;

    @property({ type: Number })
    refticks: number = 0;

    render() {
        if (!this.ticks || !this.refticks) {
            return html`unknown`;
        }
        let gameDays = Math.trunc(this.ticks / 25000);
        let playedHours = Math.trunc(this.ticks / (60 * 3600));
        let diff = this.refticks - this.ticks;

        if (diff == 0) {
            return html`latest`;
        }
        let secs = Math.trunc(diff / 60);
        if (secs < 3600) {
            return html`<span>${secs}s ago</span>`;
        }
        let hours = Math.trunc(10 * secs / 3600) / 10;
        return html`<span>${hours}h ago</span>`;
    }
}

@customElement('mapshot-listing')
class MapshotListing extends LitElement {
    @property({ type: Object })
    shots: common.ShotsJSON | undefined;

    static get styles() {
        return css`
            div.savename {
                background-color: #efefef;
                padding: 0.1ex 1ex 0.1ex 1ex;
                margin: 1ex 0.1ex 0 0.1ex;
                border-radius: 1ex;
            }
        `;
    }

    render() {
        if (!this.shots || !this.shots.all) {
            return html`No mapshots have been found. Create some and re-start mapshot server.`;
        }

        // The encoded_path gets itself encoded - i.e., double encoding. Reason
        // is that we really want to get the encoded path on the receiving page.
        // So, without encodeURI, the encoded path would get decoded, and
        // provide a raw path. That in turn would potentiall fail to get data
        // from the Go server, as it would then do a partial encoding (spaces
        // but not square bracket) which Go mux routing seems to have trouble
        // with.
        return html`
                ${this.shots.all.map((save) => html`
                    <div class="savename">
                        <h2>${save.savename} <a href="map?l=${save.savename}">[permalink]</a></h2>
                        <factorio-ticks .ticks=${save.versions[0].ticks_played}></factorio-ticks>
                        <p>
                        Available versions:
                        <ul>
                            ${save.versions.map((si) => html`
                                <li>
                                    <a href="map?path=${encodeURI(si.encoded_path)}"><factorio-relticks .ticks=${si.ticks_played} .refticks=${save.versions[0].ticks_played}></factorio-relticks></a>
                                    (<factorio-ticks .ticks=${si.ticks_played}></factorio-ticks>)
                                </li>`)}
                        </ul>
                    </div>
                `)}
        `;
    }
}

// ------ Bootstrap ------

fetch('shots.json')
    .then(resp => resp.json())
    .then((shots: common.ShotsJSON) => {
        render(html`<mapshot-listing .shots=${shots}>foo</mapshot-listing>`, document.body);
    });