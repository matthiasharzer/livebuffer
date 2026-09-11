import { Task } from '@lit/task';
import { css, html } from 'lit';
import { property } from 'lit/decorators.js';
import { Component } from '../../litutil/Component';

interface StreamInfo {
	id: string;
	stream_state: 'live' | 'archived';
	size_bytes: number;
	duration_milliseconds: number;
}

interface StreamListResponse {
	streams: StreamInfo[];
}

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

	private _streamTask = new Task(this, {
		args: () => [this.username],
		task: async ([username]) => {
			if (!username) {
				return null;
			}
			const response = await fetch(`/api/v1/${username}/list`);
			if (!response.ok) {
				throw new Error(`Failed to fetch stream list: ${response.status} ${response.statusText}`);
			}
			const data: StreamListResponse = await response.json();
			return data.streams.find(stream => stream.stream_state === 'live') || null;
		},
	});

	render() {
		if (!this.username) {
			return html`<div class="status-wrapper"><p>Missing username in the URL.</p></div>`;
		}
		const url = `/api/v1/${this.username}/live/index.m3u8`;

		return html`
			${this._streamTask.render({
				pending: () => html`<div class="status-wrapper"><p>Loading stream information...</p></div>`,
				complete: (stream: StreamInfo | null) => {
					if (!stream) {
						return html`<div class="status-wrapper"><p>User ${this.username} is not live.</p></div>`;
					}
					return html`<lb-live-video url="${url}"></lb-live-video>`;
				},
				error: e =>
					html`<div class="status-wrapper"><p>Error loading stream information: ${e instanceof Error ? e.message : 'Unknown error'}</p></div>`,
			})}
		`;
	}
}

customElements.define('lb-live-view', LiveView);
