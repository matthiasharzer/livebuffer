import '../../components/Video';
import '../../components/StreamInfoBox';
import '../../components/BackButton';
import '../../components/ViewLayout';
import { Task } from '@lit/task';
import { css, html } from 'lit';
import { property } from 'lit/decorators.js';
import { Component } from '../../litutil/Component';
import { fetchLiveStream, type StreamInfo } from '../../services/streams';

export class LiveView extends Component {
	static styles = css`
		.status-wrapper {
			display: flex;
			justify-content: center;
			align-items: center;
			width: 100%;
			height: 100%;
			font-size: 1.5rem;
			text-align: center;
		}

		::part(view) {
			max-width: 1200px;
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

	renderView(stream: StreamInfo) {
		const url = `/api/v1/video/${stream.id}/index.m3u8`;

		return html`
			<span slot="title">Watch ${stream.username} live</span>
			<div class="container video-container">
				<lb-video .title=${stream.title} .hlsSource=${url} live autoplay></lb-video>
			</div>
			<lb-stream-info-box .stream=${stream}></lb-stream-info-box>
		`;
	}

	render() {
		if (!this.username) {
			return '';
		}

		return html`
			<lb-view-layout show-back-button>
				${this._liveStreamTask.render({
					pending: () =>
						html`<div class="status-wrapper"><p>Loading live stream information...</p></div>`,
					complete: (stream: StreamInfo | null) => {
						if (!stream) {
							return html`<div class="status-wrapper"><p>${this.username} is not live</p></div>`;
						}
						return this.renderView(stream);
					},
					error: e =>
						html`<div class="status-wrapper"><p>Error loading live stream information: ${e instanceof Error ? e.message : 'Unknown error'}</p></div>`,
				})}
			</lb-view-layout>
		`;
	}
}

customElements.define('lb-live-view', LiveView);
