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

		.control-bar {
			position: absolute;
			top: 0;
			left: 0;
			width: 100%;
			display: flex;
			justify-content: space-between;
			padding: 0.5rem;
			z-index: 10;
			opacity: 1;
			transition: opacity 0.3s ease, display 0.3s ease;

			&.hidden:not(:hover) {
				opacity: 0;
			}

			button {
				background-color: var(--background, #1e1e1e);
				border: 1px solid var(--border-color, #333);

				color: var(--text-color, #e3e3e3);
				font-weight: 500;
				font-size: 1rem;
				cursor: pointer;
				padding: 0.25rem 1rem;
				border-radius: 4px;
				transition: background-color 0.3s, transform 0.1s;


				&:hover {
					background-color: rgba(255, 255, 255, 0.2);
				}

				&.go-back {
					&:hover {
						transform: translateX(-2px);
					}
					background-color: transparent;
					padding: 0.25rem ;
					border: none;
				}
			}
		}
	`;

	@property({ type: Boolean })
	autoplay: boolean = false;

	@property({ type: Boolean })
	live = false;

	@property({ attribute: false })
	hlsSource: string | null = null;

	@property({ type: String })
	title = '';

	@state()
	enabled = true;

	@state()
	error: string | null = null;

	@state()
	loaded = false;

	player: Hls | null = null;
	videoElementRef: Ref<HTMLVideoElement> = createRef();

	get hlsConfig(): Partial<HlsConfig> {
		if (this.live) {
			return {
				autoStartLoad: true,
				startPosition: -1,
				liveBackBufferLength: 0, // Keep memory usage low
			};
		}
		return {
			autoStartLoad: true,
			startPosition: 0,
		};
	}

	private setupMediaSession() {
		if (!('navigator' in window) || !('mediaSession' in navigator)) {
			return;
		}
		if (!this.videoElement) {
			return;
		}
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

	goToLive() {
		if (!this.videoElement) {
			return;
		}
		this.videoElement.currentTime = this.videoElement.duration - 5;
	}

	goBack() {
		window.history.back();
	}

	render() {
		const showStatus = !this.loaded || this.error;
		return html`
			<div class="status-wrapper ${showStatus ? 'visible' : 'hidden'}">
				${this.error ? html`<div class="status-wrapper"><p>${this.error}</p></div>` : ''}
				${!this.loaded && !this.error ? html`<div class="status-wrapper"><p>Loading stream...</p></div>` : ''}
			</div>
			<video ${ref(this.videoElementRef)} autopictureinpicture ?autoplay="${this.autoplay}" muted controls playsinline class="${this.loaded ? 'loaded' : ''}"></video>
		`;
	}
}

customElements.define('lb-video', Video);
