import { Task } from '@lit/task';
import { css, html } from 'lit';
import { property } from 'lit/decorators.js';
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
	streamId: string | null = null;

	startAt: number | null = null;

	private _streamTask = new Task(this, {
		args: () => [this.streamId],
		task: async ([streamId]) => {
			if (!streamId) {
				return null;
			}
			return fetchStream(streamId);
		},
	});

	connectedCallback(): void {
		super.connectedCallback();
		const urlParams = new URLSearchParams(window.location.search);
		const timeParam = urlParams.get('t');
		this.startAt = timeParam ? parseInt(timeParam, 10) : null;
	}

	render() {
		if (!this.streamId) {
			return html`<div class="status-wrapper"><p>Missing stream_id in the URL.</p></div>`;
		}
		const url = `/api/v1/video/${this.streamId}/index.m3u8`;

		return html`
			${this._streamTask.render({
				pending: () => html`<div class="status-wrapper"><p>Loading stream information...</p></div>`,
				complete: (stream: StreamInfo | null) => {
					if (!stream) {
						return html`<div class="status-wrapper"><p>Stream not found.</p></div>`;
					}
					return html`<lb-video .title=${stream.title} .hlsSource="${url}" .startAt=${this.startAt} autoplay></lb-video>`;
				},
				error: e =>
					html`<div class="status-wrapper"><p>Error loading stream information: ${e instanceof Error ? e.message : 'Unknown error'}</p></div>`,
			})}
		`;
	}
}

customElements.define('lb-video-view', VideoView);
