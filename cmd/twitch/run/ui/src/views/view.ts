import { css, html } from 'lit';
import { state } from 'lit/decorators.js';
import { Component } from '../litutil/Component';
import { type BroadcasterInfo, fetchBroadcasters } from '../services/streams';

export class RootView extends Component {
	static styles = css`
		:host {
			display: flex;
			flex-direction: column;
			align-items: center;
			justify-content: center;
			width: 100%;
			height: 100%;
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
			Hello World
		`;
	}
}

customElements.define('lb-root-view', RootView);
