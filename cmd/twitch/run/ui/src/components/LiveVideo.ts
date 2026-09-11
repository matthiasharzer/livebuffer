import { css, html, type PropertyValues } from 'lit';
import { property, state } from 'lit/decorators.js';
import { createRef, type Ref } from 'lit/directives/ref.js';
import mpegts from 'mpegts.js';
import { Component } from '../litutil/Component';

export class LiveVideo extends Component {
	static styles = css`
		:host {
			display: contents;
		}
	`;

	@property()
	url: string = '';

	@state()
	enabled = true;
	player: mpegts.Player | null = null;
	videoElementRef: Ref<HTMLVideoElement> = createRef();

	get mediaDataSource(): mpegts.MediaDataSource {
		return {
			type: 'mpegts',
			isLive: true,
			url: this.url,
		}
	}
	get config(): mpegts.Config {
		return {
			enableStashBuffer: true,
			stashInitialSize: 128 * 1024,
			autoCleanupSourceBuffer: true,
			autoCleanupMaxBackwardDuration: 2 * 60,
			autoCleanupMinBackwardDuration: 1 * 60,
			fixAudioTimestampGap: true,
		}
	}

	get videoElement(): HTMLVideoElement | null {
		return this.videoElementRef.value || null;
	}

	protected firstUpdated(_changedProperties: PropertyValues): void {
		super.firstUpdated(_changedProperties);
		if (!this.videoElement) {
			return;
		}
		if (!mpegts.getFeatureList().mseLivePlayback) {
			this.enabled = false;
			return;
		}
	}

	disconnectedCallback(): void {
		this.player?.destroy();
	}

	render() {
		if (!this.mediaDataSource) {
			return '';
		}
		if (!this.enabled) {
			return html`<p>Video is not supported in this browser.</p>`;
		}

		return html`
			<lb-video .mediaDataSource=${this.mediaDataSource} .config=${this.config} autoplay></lb-video>
		`;
	}
}

customElements.define('lb-live-video', LiveVideo);
