import { css, html } from 'lit';
import { property } from 'lit/decorators.js';
import { Component } from '../litutil/Component';
import { formatDuration } from '../services/formatDuration';
import type { StreamInfo } from '../services/streams';

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
				display: flex;
				gap: 0.5rem;
				flex-shrink: 0;

				a {
					--color: var(--primary-reduced-color, #9146ff);
				}

				.download-button {
					display: flex;
					align-items: center;
					justify-content: center;
					padding: 0.25rem 0.5rem;
					background-color: transparent;
					border: none;
					cursor: pointer;
					height: 1.4rem;
					flex-shrink: 0;

					svg {
						width: 1.4rem;
						height: 1.4rem;
						fill: #e3e3e3;
						transition: fill 0.2s;
					}

					&:hover svg {
						fill: var(--primary-reduced-color, #9146ff);
					}
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
					<a class="download-button" title="Download" href="/api/v1/download/${this.mustStream.id}" download>
						<svg xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="#e3e3e3"><path d="M480-320 280-520l56-58 104 104v-326h80v326l104-104 56 58-200 200ZM240-160q-33 0-56.5-23.5T160-240v-120h80v120h480v-120h80v120q0 33-23.5 56.5T720-160H240Z"/></svg>
					</a>
					<a class="simple-button watch-link" href="/video/${this.mustStream.id}">Watch Now</a>
				</div>
			</div>
		`;
	}
}

customElements.define('lb-stream-tile', StreamTile);
