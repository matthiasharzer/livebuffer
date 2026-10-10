import './Button';
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
				gap: 1rem;
				flex-shrink: 0;

				a {
					--color: var(--primary-reduced-color, #9146ff);
				}

				.download-button,
				.clip-button {
					display: flex;
					align-items: center;
					justify-content: center;
					background-color: transparent;
					border: none;
					cursor: pointer;
					height: 1.4rem;
					flex-shrink: 0;
					transition: color 0.2s;
					color: var(--text-color, #e2e2e2);

					svg {
						width: 1.4rem;
						height: 1.4rem;

					}

					&:hover {
						color: var(--primary-reduced-color, #9146ff);
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
					<a class="clip-button" title="Clip" href="/clip/${this.mustStream.id}">
						<svg xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M160-240v-320 320Zm0 80q-33 0-56.5-23.5T80-240v-480q0-33 23.5-56.5T160-800l80 160h120l-80-160h80l80 160h120l-80-160h80l80 160h120l-80-160h120q33 0 56.5 23.5T880-720v160H160v320h320v80H160Zm400 40v-123l221-220q9-9 20-13t22-4q12 0 23 4.5t20 13.5l37 37q8 9 12.5 20t4.5 22q0 11-4 22.5T903-340L683-120H560Zm300-263-37-37 37 37ZM620-180h38l121-122-18-19-19-18-122 121v38Zm141-141-19-18 37 37-18-19Z"/></svg>
					</a>
					<a class="download-button" title="Download" href="/api/v1/download/${this.mustStream.id}" target="_blank">
						<svg xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M480-320 280-520l56-58 104 104v-326h80v326l104-104 56 58-200 200ZM240-160q-33 0-56.5-23.5T160-240v-120h80v120h480v-120h80v120q0 33-23.5 56.5T720-160H240Z"/></svg>
					</a>
					<lb-button variant="secondary" href="/video/${this.mustStream.id}">Watch Now</lb-button>
				</div>
			</div>
		`;
	}
}

customElements.define('lb-stream-tile', StreamTile);
