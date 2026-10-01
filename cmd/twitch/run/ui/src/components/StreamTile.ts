import { css, html } from 'lit';
import { property } from 'lit/decorators.js';
import { Component } from '../litutil/Component';
import type { StreamInfo } from '../services/streams';

const formatDuration = (milliseconds: number): string => {
	const totalSeconds = Math.floor(milliseconds / 1000);
	const hours = Math.floor(totalSeconds / 3600);
	const minutes = Math.floor((totalSeconds % 3600) / 60);
	const seconds = totalSeconds % 60;

	const parts = [];
	if (hours > 0) {
		parts.push(`${hours}h`);
	}
	if (minutes > 0 || hours > 0) {
		parts.push(`${minutes}m`);
	}
	parts.push(`${seconds}s`);

	return parts.join(' ');
};

export class StreamTile extends Component {
	static styles = css`

		.stream-item {
			display: flex;
			flex-direction: row;
			align-items: center;
			justify-content: space-between;
			width: 100%;
			gap: 0.5rem;

			.info {
				display: flex;
				flex-direction: column;
				gap: 0.25rem;
				flex: 1 1 auto;
				min-width: 0;

				h4 {
					margin: 0;
					font-size: 1rem;
					white-space: nowrap;
					overflow: hidden;
					text-overflow: ellipsis;
				}

				p {
					margin: 0;
					font-size: 0.875rem;
					color: #aaa;
				}
			}

			.actions {
				a {
					--color: var(--primary-reduced-color, #9146ff);
				}
			}
		}
	`;

	@property({ attribute: false })
	stream: StreamInfo | null = null;

	get mustStream(): StreamInfo {
		if (!this.stream) {
			throw new Error('Stream is not set');
		}
		return this.stream;
	}

	renderWatchLink() {
		return html`<a class="watch-link" href="/video/${this.mustStream.id}">Watch Now</a>`;
	}

	render() {
		if (!this.stream) {
			return '';
		}
		return html`
			<div class="stream-item">
				<div class="info">
					<h4 title="${this.mustStream.title}">${this.mustStream.title}</h4>
					<p>${this.mustStream.state} · ${new Date(this.mustStream.started_at).toLocaleString()} · ${formatDuration(this.mustStream.duration_milliseconds)} ·  ${this.mustStream.size}</p>
				</div>
				<div class="actions">
					${this.renderWatchLink()}
				</div>
			</div>
		`;
	}
}

customElements.define('lb-stream-tile', StreamTile);
