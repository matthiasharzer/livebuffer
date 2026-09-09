import { css, html, type PropertyValues } from 'lit';
import { property, state } from 'lit/decorators.js';
import { createRef, type Ref, ref } from 'lit/directives/ref.js';
import mpegts from 'mpegts.js';
import { Component } from '../litutil/Component';

export class LiveVideo extends Component {
	static styles = css`
		video {
			width: 100%;
			height: 100%;
			background-color: black;
			max-width: 100%;
			max-height: 100%;
		}
	`;

	@property()
	url: string = '';

	@state()
	enabled = true;
	player: mpegts.Player | null = null;
	videoElementRef: Ref<HTMLVideoElement> = createRef();

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
		this.player = mpegts.createPlayer({
			type: 'mpegts',
			isLive: true,
			url: this.url,
		});
		this.player.attachMediaElement(this.videoElement);
		this.player.load();
		this.player.play();
		this.player.on(mpegts.Events.ERROR, (errorType, errorDetail) => {
			this.dispatch('live-video-error', { errorType, errorDetail });
		});
		this.player.on(mpegts.Events.MEDIA_INFO, () => {
			this.dispatch('live-video-loading-complete');
		});
	}

	disconnectedCallback(): void {
		this.player?.destroy();
	}

	render() {
		if (!this.url) {
			return '';
		}
		if (!this.enabled) {
			return html`<p>Live video is not supported in this browser.</p>`;
		}

		return html`
			<video ${ref(this.videoElementRef)} autoplay muted controls playsinline></video>
		`;
	}
}

customElements.define('lb-live-video', LiveVideo);
