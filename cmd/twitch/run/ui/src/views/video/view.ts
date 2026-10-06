import { Task } from '@lit/task';
import { css, html } from 'lit';
import { property } from 'lit/decorators.js';
import { Component } from '../../litutil/Component';
import { fetchStream, type StreamInfo } from '../../services/streams';

export class VideoView extends Component {
	static styles = css`
		:host {
			display: flex;
			flex-direction: column;
			align-items: center;
			justify-content: safe center;
			width: 100%;
			height: 100%;
			padding: 1rem;
			overflow-y: auto;
			scrollbar-gutter: stable;
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

		.video-view {
			display: flex;
			flex-direction: column;
			gap: 1rem;

			width: 100%;
			max-width: 1200px;
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
		const parsedTime = timeParam ? parseInt(timeParam, 10) : NaN;
		this.startAt = Number.isFinite(parsedTime) ? parsedTime : null;
	}

	renderView(stream: StreamInfo) {
		const url = `/api/v1/video/${stream.id}/index.m3u8`;

		return html`
			<div class="container video-container">
				<lb-video .title=${stream.title} .hlsSource="${url}" .startAt=${this.startAt} autoplay></lb-video>
			</div>
			<lb-stream-info-box .stream=${stream}></lb-stream-info-box>
		`;
	}

	render() {
		return html`
			<div class="video-view">
				<lb-back-button href="/"></lb-back-button>
				${this._streamTask.render({
					pending: () =>
						html`<div class="status-wrapper"><p>Loading stream information...</p></div>`,
					complete: (stream: StreamInfo | null) => {
						if (!stream) {
							return html`<div class="status-wrapper"><p>Stream not found.</p></div>`;
						}
						return this.renderView(stream);
					},
					error: e =>
						html`<div class="status-wrapper"><p>Error loading stream information: ${e instanceof Error ? e.message : 'Unknown error'}</p></div>`,
				})}
			</div>
		`;
	}
}

customElements.define('lb-video-view', VideoView);
