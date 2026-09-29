import { Task } from '@lit/task';
import { css, html } from 'lit';
import { Component } from '../../litutil/Component';
import { type BroadcasterInfo, fetchBroadcasters } from '../../services/streams';

export class RootView extends Component {
	static styles = css`
		:host {
			display: flex;
			flex-direction: column;
			align-items: center;
			justify-content: center;
			width: 100%;
			height: 100%;
			padding: 1rem;
		}

		.overview {
			display: flex;
			flex-direction: column;

			width: 100%;
			max-width: 600px;

			h1 {
				width: fit-content
			}
		}

		.broadcasters-list {
			margin-top: 1rem;
			width: 100%;
		}
	`;

	broadcasters = new Task(this, {
		args: () => [],
		task: async () => {
			return fetchBroadcasters();
		},
	});

	renderBroadcasters(broadcasters: BroadcasterInfo[]) {
		return broadcasters.map(
			broadcaster => html`
				<lb-broadcaster-tile .broadcaster=${broadcaster}></lb-broadcaster-tile>
			`,
		);
	}

	render() {
		return html`
			<div class="overview">
				<h1>Welcome to LiveBuffer</h1>
				<div class="broadcasters-list">
					${this.broadcasters.render({
						pending: () => html`<p>Loading broadcasters...</p>`,
						complete: (broadcasters: BroadcasterInfo[]) => {
							if (broadcasters.length === 0) {
								return html`<p>No broadcasters found.</p>`;
							}
							return this.renderBroadcasters(broadcasters);
						},
						error: e =>
							html`<p>Error loading broadcasters: ${e instanceof Error ? e.message : 'Unknown error'}</p>`,
					})}
					</div>
				</div>
		`;
	}
}

customElements.define('lb-root-view', RootView);
