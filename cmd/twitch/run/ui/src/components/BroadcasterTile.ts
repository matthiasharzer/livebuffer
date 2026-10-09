import './StreamTile';
import { css, html } from 'lit';
import { property } from 'lit/decorators.js';
import { Component } from '../litutil/Component';
import type { BroadcasterInfo } from '../services/streams';

export class BroadcasterTile extends Component {
	static styles = css`
		:host {
			display: block;
			width: 100%;
		}

		@keyframes pulse {
			0% {
				transform: scale(1);
				opacity: 1;
			}
			50% {
				transform: scale(1.5);
				opacity: 0.5;
			}
			100% {
				transform: scale(1);
				opacity: 1;
			}
		}

		details {
			background-color: var(--background, #1e1e1e);
			border: 1px solid var(--border-color, #333);
			border-radius: 0.5rem;
			padding: 0.5rem;
			margin-bottom: 1rem;
			width: 100%;

			&::details-content {
				height: 0;
				overflow: clip;
				opacity: 0;
				transition:
					height 0.3s ease,
					opacity 0.3s ease,
					content-visibility 0.3s ease allow-discrete;
			}
			&[open]::details-content {
				height: auto;
				opacity: 1;
			}
		}

		summary {
			display: flex;
			justify-content: flex-start;
			align-items: center;
			cursor: pointer;

			.content {
				display: flex;
				justify-content: space-between;
				align-items: center;
				width: 100%;
				padding: 0.25rem 0.5rem;
				gap: 1rem;
			}
		}

		.angle-icon {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 1rem;
			height: 1rem;
			transition: transform 0.3s;
			transform: rotate(-90deg);
		}

		details[open] .angle-icon {
			transform: rotate(0deg);
		}

		.info {
			display: flex;
			flex-direction: column;
			gap: 0.5rem;

			.header {
				display: flex;
				align-items: center;
				gap: 0.5rem;
			}

			.description {
				font-size: 0.875rem;
				color: #aaa;
			}
		}

		.live-bubble {
			background-color: #fa7878;
			color: white;
			font-weight: bold;
			padding: 0.2rem 0.5rem;
			border-radius: 0.25rem;
			font-size: 0.75rem;
			white-space: nowrap;

			.dot {
				display: inline-block;
				width: 0.5rem;
				height: 0.5rem;
				background-color: #fc3c3c;
				border-radius: 50%;
				margin-left: 0.25rem;

				&:before {
					content: '';
					display: block;
					background-color: #fc3c3c;
					width: 100%;
					height: 100%;
					border-radius: 50%;
					animation: pulse 5s infinite;
				}
			}
		}

		.actions {
			a {
				--color: var(--primary-color, #9146ff);
			}
		}

		.stream-list {
			padding: 1rem 0.5rem 0.25rem 1.5rem;
			display: flex;
			flex-direction: column;
			gap: 1rem;

			.no-streams {
				font-size: 0.875rem;
				color: #aaa;
			}
		}
	`;

	@property({ attribute: false })
	broadcaster: BroadcasterInfo | null = null;

	get mustBroadcaster() {
		if (!this.broadcaster) {
			throw new Error('Broadcaster is not set');
		}
		return this.broadcaster;
	}

	get streamsText() {
		const streamCount = this.mustBroadcaster.streams.length;
		return `${streamCount} stream${streamCount !== 1 ? 's' : ''}`;
	}

	get latestStream() {
		const streams = this.mustBroadcaster.streams;
		if (streams.length === 0) {
			return null;
		}
		const lastStream = streams.reduce((latest, current) => {
			return new Date(current.started_at) > new Date(latest.started_at) ? current : latest;
		});
		return lastStream;
	}

	get lastLiveText() {
		const latestStream = this.latestStream;
		if (!latestStream) {
			return 'never live';
		}
		if (latestStream.state === 'live') {
			return 'live right now';
		}
		const startedAtMs = new Date(latestStream.started_at).getTime();
		const lastLiveMs = startedAtMs + latestStream.duration_milliseconds;

		const now = new Date();
		const diffMs = now.getTime() - lastLiveMs;
		const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
		if (diffDays > 0) {
			return `live ${diffDays} day${diffDays !== 1 ? 's' : ''} ago`;
		}
		const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
		if (diffHours > 0) {
			return `live ${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`;
		}
		const diffMinutes = Math.floor(diffMs / (1000 * 60));
		if (diffMinutes > 0) {
			return `live ${diffMinutes} minute${diffMinutes !== 1 ? 's' : ''} ago`;
		}
		return 'live just now';
	}

	get liveBubble() {
		const latestStream = this.latestStream;
		if (!latestStream) {
			return null;
		}
		if (latestStream.state === 'live') {
			return html`
				<span class="live-bubble">
					LIVE
					<span class="dot"></span>
				</span>
				`;
		}
		return null;
	}

	get sortedStreams() {
		return this.mustBroadcaster.streams.toSorted(
			(a, b) => new Date(b.started_at).getTime() - new Date(a.started_at).getTime(),
		);
	}

	render() {
		if (!this.broadcaster) {
			return '';
		}

		return html`
			<details>
				<summary>
					<div class="angle-icon">
						<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-chevron-down" viewBox="0 0 16 16">
							<path fill-rule="evenodd" d="M1.646 4.646a.5.5 0 0 1 .708 0L8 10.293l5.646-5.647a.5.5 0 0 1 .708.708l-6 6a.5.5 0 0 1-.708 0l-6-6a.5.5 0 0 1 0-.708z"/>
						</svg>
					</div>
					<div class="content">
						<div class="info">
							<div class="header">
								<h3>${this.broadcaster.username}</h3>
									${this.liveBubble}
							</div>
							<p class="description">${this.streamsText} · ${this.lastLiveText}</p>
						</div>
						<div class="actions">
								${
									this.latestStream && this.latestStream.state === 'live'
										? html`<a class="simple-button watch-link" href="/live/${this.broadcaster.username}">Watch Live</a>`
										: ''
								}
						</div>
					</div>
				</summary>
				<div class="stream-list">
					${
						this.mustBroadcaster.streams.length === 0
							? html`<p class="no-streams">No streams available.</p>`
							: this.sortedStreams.map(
									stream => html`<lb-stream-tile .stream=${stream}></lb-stream-tile>`,
								)
					}

				</div>
			</details>
		`;
	}
}

customElements.define('lb-broadcaster-tile', BroadcasterTile);
