import { Task } from '@lit/task';
import { css, html } from 'lit';
import { property } from 'lit/decorators.js';
import { Component } from '../../litutil/Component';
import { fetchLiveStream, type StreamInfo } from '../../services/streams';

export class LiveView extends Component {
	static styles = css`
		:host {
			display: block;
			width: 100dvw;
			height: 100dvh;
		}

		.status-wrapper {
			display: flex;
			justify-content: center;
			align-items: center;
			width: 100%;
			height: 100%;
			font-size: 1.5rem;
			text-align: center;
		}

		lb-live-video {
			display: block;
			height: 100%;
		}
	`;

	@property({ attribute: false })
	username: string | null = null;

	private _liveStreamTask = new Task(this, {
		args: () => [this.username],
		task: async ([username]) => {
			return fetchLiveStream(username || '');
		},
	});

	render() {
		if (!this.username) {
			return html`<div class="status-wrapper"><p>Missing username in the URL.</p></div>`;
		}
		const url = `/api/v1/${this.username}/live/index.m3u8`;

		return html`
			${this._liveStreamTask.render({
				pending: () =>
					html`<div class="status-wrapper"><p>Loading live stream information...</p></div>`,
				complete: (stream: StreamInfo | null) => {
					if (!stream) {
						return html`<div class="status-wrapper"><p>${this.username} is not live.</p></div>`;
					}
					return html`<lb-live-video url="${url}"></lb-live-video>`;
				},
				error: e =>
					html`<div class="status-wrapper"><p>Error loading live stream information: ${e instanceof Error ? e.message : 'Unknown error'}</p></div>`,
			})}
		`;
	}
}

customElements.define('lb-live-view', LiveView);
