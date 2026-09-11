import { css, html, type PropertyValues } from 'lit';
import { property, state } from 'lit/decorators.js';
import { createRef, type Ref, ref } from 'lit/directives/ref.js';
import mpegts from 'mpegts.js';
import { Component } from '../litutil/Component';

export class LiveVideo extends Component {
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
				display: unset;
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
	mediaDataSource: mpegts.MediaDataSource | null = null;

	@property({ attribute: false })
	config?: mpegts.Config

	@state()
	enabled = true;
	player: mpegts.Player | null = null;
	videoElementRef: Ref<HTMLVideoElement> = createRef();

	@state()
	error: string | null = null;

	@state()
	loaded = false;

	get videoElement(): HTMLVideoElement | null {
		return this.videoElementRef.value || null;
	}

	protected firstUpdated(_changedProperties: PropertyValues): void {
		super.firstUpdated(_changedProperties);
		if (!this.videoElement) {
			return;
		}
		if (!this.mediaDataSource) {
			return;
		}
		this.player = mpegts.createPlayer(this.mediaDataSource, this.config);
		this.player.attachMediaElement(this.videoElement);
		this.player.load();
		this.player.play();
		this.player.on(mpegts.Events.ERROR, (errorType, errorDetail) => {
			this.error = `Error loading video: ${errorType} - ${errorDetail}`;
		});
		this.player.on(mpegts.Events.MEDIA_INFO, () => {
			this.loaded = true;
		});
	}

	disconnectedCallback(): void {
		this.player?.destroy();
	}

	render() {
		const showStatus = !this.loaded || this.error;
		return html`
			<div class="status-wrapper ${showStatus ? 'visible' : 'hidden'}">
				${this.error ? html`<div class="status-wrapper"><p>${this.error}</p></div>` : ''}
				${!this.loaded && !this.error ? html`<div class="status-wrapper"><p>Loading live stream...</p></div>` : ''}
			</div>
			<video ${ref(this.videoElementRef)} ?autoplay=${this.autoplay} muted controls playsinline class="${this.loaded ? 'loaded' : ''}"></video>
		`;
	}
}

customElements.define('lb-video', LiveVideo);
