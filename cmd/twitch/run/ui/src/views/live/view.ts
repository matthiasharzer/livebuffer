import { Task } from '@lit/task';
import type { HlsConfig } from 'hls.js';
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

		lb-video {
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

	get hlsConfig(): Partial<HlsConfig> {
		return {
			autoStartLoad: true,
			startPosition: -1,
			liveDurationInfinity: true,
			liveBackBufferLength: 0, // Keep memory usage low
		};
	}

	render() {
		if (!this.username) {
			return html`<div class="status-wrapper"><p>Missing username in the URL.</p></div>`;
		}
		const url = `/api/v1/live/${this.username}/index.m3u8`;

		return html`
			${this._liveStreamTask.render({
				pending: () =>
					html`<div class="status-wrapper"><p>Loading live stream information...</p></div>`,
				complete: (stream: StreamInfo | null) => {
					if (!stream) {
						return html`<div class="status-wrapper"><p>${this.username} is not live</p></div>`;
					}
					return html`<lb-video .title=${stream.title} .hlsSource=${url} live autoplay></lb-video>`;
				},
				error: e =>
					html`<div class="status-wrapper"><p>Error loading live stream information: ${e instanceof Error ? e.message : 'Unknown error'}</p></div>`,
			})}
		`;
	}
}

customElements.define('lb-live-view', LiveView);
