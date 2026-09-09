import type { RouterLocation } from '@vaadin/router';
import { css, html } from 'lit';
import { property, state } from 'lit/decorators.js';
import { Component } from '../../litutil/Component';

export class WatchView extends Component {
	static styles = css`
		:host {
			display: block;
			width: 100dvw;
			height: 100dvh;
			background-color: #1a1919;
			color: #e2e2e2;
		}

		.status-wrapper {
			display: flex;
			justify-content: center;
			align-items: center;
			width: 100%;
			height: 100%;
			font-size: 1.5rem;
		}

		lb-live-video {
			display: none;

			&.loaded {
				display: unset;
			}
		}
	`;

	@state()
	error: string | null = null;

	@state()
	loaded = false;

	@property({ attribute: false })
	accessor location: RouterLocation | undefined;

	get username(): string | null {
		const username = this.location?.params.username || null;
		if (username) {
			return username as string;
		}
		return null;
	}

	onError(event: CustomEvent<{ errorType: string; errorDetail: string }>) {
		this.error = `Error loading stream: ${event.detail.errorType} - ${event.detail.errorDetail}`;
	}

	onLoad() {
		this.loaded = true;
	}

	render() {
		if (!this.username) {
			return html`<div class="status-wrapper"><p>Missing username in the URL.</p></div>`;
		}
		const url = `/api/v1/${this.username}/live`;
		return html`
			${this.error ? html`<div class="status-wrapper"><p>${this.error}</p></div>` : ''}
			${!this.loaded && !this.error ? html`<div class="status-wrapper"><p>Loading live stream...</p></div>` : ''}
			<lb-live-video url="${url}" @live-video-error="${this.onError}" @live-video-loading-complete="${this.onLoad}" class="${this.loaded ? 'loaded' : ''}"></lb-live-video>
		`;
	}
}

customElements.define('lb-watch-view', WatchView);
