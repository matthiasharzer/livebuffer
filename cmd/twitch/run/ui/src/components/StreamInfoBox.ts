import { css, html } from 'lit';
import { property } from 'lit/decorators.js';
import { Component } from '../litutil/Component';
import { formatDuration } from '../services/formatDuration';
import type { StreamInfo } from '../services/streams';

export class StreamInfoBox extends Component {
	static styles = css`
		.username, .stream-details {
			color: var(--text-reduced-color, #aaa);
		}

		.stream-info-box {
			display: flex;
			flex-direction: row;
			justify-content: space-between;
			flex-wrap: wrap;

			gap: 0.5rem;
			padding: 1rem;
		}

		.content {
			display: flex;
			flex-direction: column;
			gap: 0.5rem;
			min-width: 0;
		}

		.actions {
			flex: 0 0 auto;
			display: flex;
			flex-direction: row;
			gap: 0.5rem;
			align-items: center;
		}

		.actions > a {
			display: flex;
			align-items: center;
			justify-content: center;
			gap: 0.25rem;
		}

		a {
			--color: var(--primary-reduced-color, #9146ff);
		}

		h2 {
			min-width: 0;
			width: 100%;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}
	`;

	@property({ attribute: false })
	stream: StreamInfo | null = null;

	@property({ attribute: false })
	getCurrentTime: (() => number) | null = null;

	injectClipURL(e: MouseEvent) {
		if (!this.stream) {
			return;
		}
		const currentTime = this.getCurrentTime?.() ?? 0;
		const clipURL = `/clip/${this.stream.id}?t=${Math.floor(currentTime)}`;
		const target = e.currentTarget as HTMLAnchorElement;
		target.href = clipURL;
	}

	downloadStream() {
		if (!this.stream) {
			return;
		}
		const downloadURL = `/api/v1/download/${this.stream.id}`;
		window.open(downloadURL, '_blank');
	}

	renderDate(dateString: string) {
		const date = new Date(dateString);
		const formattedDate = date.toLocaleString(undefined, {
			year: 'numeric',
			month: 'long',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
		});
		return html`
			<time datetime="${date.toISOString()}">${formattedDate}</time>
		`;
	}

	renderStreamDetails(stream: StreamInfo) {
		return html`
			<div class="stream-details">
				${this.renderDate(stream.started_at)} · ${stream.size} · ${formatDuration(stream.duration_milliseconds)}
			</div>
		`;
	}

	render() {
		if (!this.stream) {
			return '';
		}

		return html`
			<div class="container stream-info-box">
				<div class="content">
					<p class="username">${this.stream.username}</p>
					<h2 title="${this.stream.title}">${this.stream.title}</h2>
					${this.renderStreamDetails(this.stream)}
				</div>
				<div class="actions">
					<lb-button variant="secondary" href="/clip/${this.stream.id}" @mousedown=${this.injectClipURL} show-icon>
						<svg slot="icon" xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M160-240v-320 320Zm0 80q-33 0-56.5-23.5T80-240v-480q0-33 23.5-56.5T160-800l80 160h120l-80-160h80l80 160h120l-80-160h80l80 160h120l-80-160h120q33 0 56.5 23.5T880-720v160H160v320h320v80H160Zm400 40v-123l221-220q9-9 20-13t22-4q12 0 23 4.5t20 13.5l37 37q8 9 12.5 20t4.5 22q0 11-4 22.5T903-340L683-120H560Zm300-263-37-37 37 37ZM620-180h38l121-122-18-19-19-18-122 121v38Zm141-141-19-18 37 37-18-19Z"/></svg>
						Clip
					</lb-button>
					<lb-button variant="secondary" title="Download" @click=${this.downloadStream} download show-icon>
						<svg slot="icon" xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M480-320 280-520l56-58 104 104v-326h80v326l104-104 56 58-200 200ZM240-160q-33 0-56.5-23.5T160-240v-120h80v120h480v-120h80v120q0 33-23.5 56.5T720-160H240Z"/></svg>
						Download
					</lb-button>
				</div>
			</div>
		`;
	}
}

customElements.define('lb-stream-info-box', StreamInfoBox);
