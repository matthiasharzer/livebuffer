import '../../components/BroadcasterTile';
import '../../components/ViewLayout';
import { Task } from '@lit/task';
import { css, html } from 'lit';
import { Component } from '../../litutil/Component';
import { type BroadcasterInfo, fetchBroadcasters } from '../../services/streams';

export class RootView extends Component {
	static styles = css`
		::part(view) {
			max-width: 700px;
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
			<lb-view-layout>
				<span slot="title">Welcome to LiveBuffer</span>
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
				</lb-view-layout>
		`;
	}
}

customElements.define('lb-root-view', RootView);
