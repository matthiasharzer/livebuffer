import Hls, { type HlsConfig } from 'hls.js';
import { css, html, type PropertyValues } from 'lit';
import { property, state } from 'lit/decorators.js';
import { createRef, type Ref, ref } from 'lit/directives/ref.js';
import { Component } from '../litutil/Component';

export class Video extends Component {
	static styles = css`
		:host {
			display: contents;
		}

		video {
			width: 100%;
			height: 100%;
			background-color: black;
			max-width: 100%;
			max-height: 100%;
			display: none;

			&.loaded {
				display: block;
			}
		}

		.status-wrapper {
			display: flex;
			justify-content: center;
			align-items: center;
			width: 100%;
			height: 100%;
			font-size: 1.5rem;
			text-align: center;
			padding: 1rem;

			&.hidden {
				display: none;
			}
		}
	`;

	@property({ type: Boolean })
	autoplay: boolean = false;

	@property({ attribute: false })
	hlsConfig?: Partial<HlsConfig>;

	@property({ attribute: false })
	hlsSource: string | null = null;

	@property({ type: String })
	title = '';

	@state()
	enabled = true;

	player: Hls | null = null;
	videoElementRef: Ref<HTMLVideoElement> = createRef();

	@state()
	error: string | null = null;

	@state()
	loaded = false;

	private setupMediaSession() {
		if (!('navigator' in window) || !('mediaSession' in navigator)) {
			return;
		}
		if (!this.videoElement) {
			return;
		}
		console.log(this.title);
		const video = this.videoElement;
		navigator.mediaSession.metadata = new MediaMetadata({
			title: this.title || 'LiveBuffer Video',
		});

		navigator.mediaSession.setActionHandler('play', async () => {
			try {
				await video.play();
				navigator.mediaSession.playbackState = 'playing';
			} catch (err) {
				console.error('Play failed:', err);
			}
		});

		navigator.mediaSession.setActionHandler('pause', () => {
			video.pause();
			navigator.mediaSession.playbackState = 'paused';
		});

		// Optional: Seeking (if not a live stream)
		navigator.mediaSession.setActionHandler('seekto', details => {
			if (details.fastSeek && 'fastSeek' in video) {
				video.fastSeek(details.seekTime || 0);
			} else {
				video.currentTime = details.seekTime || 0;
			}
		});
	}

	get videoElement(): HTMLVideoElement | null {
		return this.videoElementRef.value || null;
	}

	protected firstUpdated(_changedProperties: PropertyValues): void {
		super.firstUpdated(_changedProperties);
		if (!this.videoElement) {
			return;
		}
		if (!this.hlsSource) {
			this.error = 'No HLS source provided.';
			return;
		}
		if (!Hls.isSupported()) {
			this.error = 'HLS playback is not supported in this browser.';
			return;
		}
		this.player = new Hls(this.hlsConfig);
		this.player.loadSource(this.hlsSource);
		this.player.attachMedia(this.videoElement);
		this.player.on(Hls.Events.ERROR, (_, data) => {
			if (data.fatal) {
				this.error = `Error loading video: ${data.type} - ${data.details}`;
				return;
			}
			if (this.loaded) {
				// ignored, since this is likely a network error that occurred after seeking
				return;
			}
			this.error = `Error loading video: ${data.type} - ${data.details}`;
		});
		this.player.on(Hls.Events.MANIFEST_PARSED, () => {
			this.loaded = true;
			this.dispatch('video-loaded', null, { bubbles: true, composed: true });
			if (this.autoplay) {
				this.videoElement?.play();
			}
		});

		this.setupMediaSession();
	}

	disconnectedCallback(): void {
		this.player?.destroy();
	}

	render() {
		const showStatus = !this.loaded || this.error;
		return html`
			<div class="status-wrapper ${showStatus ? 'visible' : 'hidden'}">
				${this.error ? html`<div class="status-wrapper"><p>${this.error}</p></div>` : ''}
				${!this.loaded && !this.error ? html`<div class="status-wrapper"><p>Loading stream...</p></div>` : ''}
			</div>
			<video ${ref(this.videoElementRef)} ?autoplay="${this.autoplay}" muted controls playsinline class="${this.loaded ? 'loaded' : ''}"></video>
		`;
	}
}

customElements.define('lb-video', Video);
