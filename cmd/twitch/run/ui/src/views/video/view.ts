import { Task } from '@lit/task';
import type { HlsConfig } from 'hls.js';
import { css, html } from 'lit';
import { property, state } from 'lit/decorators.js';
import { createRef, ref } from 'lit/directives/ref.js';
import type { Video } from '../../components/Video';
import { Component } from '../../litutil/Component';
import { fetchStream, type StreamInfo } from '../../services/streams';

export class VideoView extends Component {
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
	`;

	@property({ attribute: false })
	username: string | null = null;

	@state()
	streamId: string | null = null;

	videoRef = createRef<Video>();

	private _streamTask = new Task(this, {
		args: () => [this.username, this.streamId],
		task: async ([username, streamId]) => {
			if (!username || !streamId) {
				return null;
			}
			return fetchStream(username, streamId);
		},
	});

	connectedCallback(): void {
		super.connectedCallback();

		const params = new URLSearchParams(window.location.search);
		this.streamId = params.get('stream_id');
	}

	get hlsConfig(): Partial<HlsConfig> {
		return {
			autoStartLoad: true,
			startPosition: 0,
		};
	}

	render() {
		if (!this.username) {
			return html`<div class="status-wrapper"><p>Missing username in the URL.</p></div>`;
		}
		if (!this.streamId) {
			return html`<div class="status-wrapper"><p>Missing stream_id in the URL.</p></div>`;
		}
		const url = `/api/v1/${this.username}/video/${this.streamId}/index.m3u8`;

		return html`
			${this._streamTask.render({
				pending: () => html`<div class="status-wrapper"><p>Loading stream information...</p></div>`,
				complete: (stream: StreamInfo | null) => {
					if (!stream) {
						return html`<div class="status-wrapper"><p>Stream not found.</p></div>`;
					}
					return html`<lb-video ${ref(this.videoRef)} .hlsSource="${url}" autoplay></lb-video>`;
				},
				error: e =>
					html`<div class="status-wrapper"><p>Error loading stream information: ${e instanceof Error ? e.message : 'Unknown error'}</p></div>`,
			})}
		`;
	}
}

customElements.define('lb-video-view', VideoView);
