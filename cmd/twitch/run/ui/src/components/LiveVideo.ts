import type { HlsConfig } from 'hls.js';
import { css, html } from 'lit';
import { property } from 'lit/decorators.js';
import { Component } from '../litutil/Component';

export class LiveVideo extends Component {
	static styles = css`
		:host {
			display: contents;
		}
	`;

	@property()
	url: string = '';

	get hlsConfig(): Partial<HlsConfig> {
		return {
			autoStartLoad: true,
			startPosition: -1,
			debug: false,
			liveDurationInfinity: true,
			liveBackBufferLength: 0, // Keep memory usage low
		};
	}

	render() {
		if (!this.url) {
			return '';
		}
		return html`
			<lb-video .hlsSource=${this.url} .hlsConfig=${this.hlsConfig} autoplay></lb-video>
		`;
	}
}

customElements.define('lb-live-video', LiveVideo);
