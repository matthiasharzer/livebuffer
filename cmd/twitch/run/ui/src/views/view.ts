import { css, html } from 'lit';
import { state } from 'lit/decorators.js';
import { Component } from '../litutil/Component';
import { type BroadcasterInfo, fetchBroadcasters } from '../services/streams';

export class RootView extends Component {
	static styles = css`
		:host {
		}
	`;

	@state()
	broadcasters: Promise<BroadcasterInfo[]> | null = null;

	connectedCallback(): void {
		super.connectedCallback();
		this.broadcasters = fetchBroadcasters();
	}



	render() {
		return html`

		`;
	}
}

customElements.define('lb-root-view', RootView);
