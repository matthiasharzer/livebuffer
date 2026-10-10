import '../../components/Video';
import '../../components/Button';
import '../../components/DurationInput';
import '../../components/BackButton';
import '../../components/ViewLayout';
import { Task } from '@lit/task';
import { css, html } from 'lit';
import { property, query, state } from 'lit/decorators.js';
import { ifDefined } from 'lit/directives/if-defined.js';
import type { Video } from '../../components/Video';
import { Component } from '../../litutil/Component';
import { formatDurationParts } from '../../services/formatDuration';
import { fetchStream, type StreamInfo } from '../../services/streams';

export class ClipView extends Component {
	static styles = css`
		::part(view) {
			max-width: 800px;
		}
		h1, h2 {
			width: 100%;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}

		.clip-info-controls {
			display: flex;
			flex-direction: column;
			gap: 1rem;
			padding: 0.5rem;
		}

		.clip-controls {
			display: flex;
			flex-direction: row;
			justify-content: space-between;
			flex-wrap: wrap;
			gap: 0.5rem;

			.clip-range {
				display: flex;
				flex-direction: row;
				flex-wrap: wrap;
				gap: 0.5rem;
			}
		}

		.chip {
			display: flex;
			flex-direction: column;
			justify-content: flex-start;
			gap: 0.5rem;
			padding: 0.5rem;
			background-color: var(--surface-color, #2a2a2a);
			border: 1px solid var(--surface-border-color, #444);
			border-radius: 4px;
			position: relative;


			.input-wrapper {
				display: flex;
				flex-direction: column;
				align-items: flex-start;
				background-color: transparent;

				button {
					margin-top: 0.1rem;
					padding: 0;
					background-color: transparent;
					color: var(--primary-reduced-color, #cd79fd);
					border: none;
					font-size: 0.875rem;
					cursor: pointer;
				}
			}

			h3 {
				width: 100%;
				text-align: left;
			}
		}

		.back-button {
			align-self: flex-start;

			svg {
				transition: transform 0.2s ease;
			}
			&:hover {
				svg {
					transform: translateX(-4px);
				}
			}
		}

		lb-duration-input {
			--bg-color: var(--background, #1e1e1e);
		}

		.actions {
			display: flex;
			flex-direction: column;
			align-items: center;
			gap: 0.5rem;

			button,
			a {
				width: 100%;
				display: flex;
				align-items: center;
				justify-content: center;
				cursor: pointer;
				gap: 0.5rem;
			}
		}

		.format-options {
			display: flex;
			flex-direction: column;
			gap: 0.5rem;

			p {
				color: #aaa;
				font-size: 0.875rem;
				font: monospace;
			}
		}
	`;

	@property({ attribute: false })
	streamId: string | null = null;

	@state()
	startMs: number = 0;

	@state()
	endMs: number = 0;

	@state()
	isClipPreview: boolean = false;

	@state()
	isDownloading: boolean = false;

	@query('lb-video')
	videoElement!: Video;

	@state()
	durationValid = {
		start: true,
		end: true,
	};

	startAtSeconds: number | null = null;

	get durationMs(): number | null {
		if (!this.isDurationInputsValid) {
			return null;
		}

		return this.endMs - this.startMs;
	}

	get isDurationInputsValid(): boolean {
		return this.durationValid.start && this.durationValid.end;
	}

	get isValid(): boolean {
		return this.startMs >= 0 && this.endMs > this.startMs && this.isDurationInputsValid;
	}

	get clipUrl(): string | null {
		if (!this.streamId) {
			return null;
		}

		const start = formatDurationParts(this.startMs).join('');
		const end = formatDurationParts(this.endMs).join('');

		const params = new URLSearchParams();
		params.set('start', start);
		params.set('end', end);
		return `/api/v1/clip/${this.streamId}?${params.toString()}`;
	}

	private _streamTask = new Task(this, {
		args: () => [this.streamId],
		task: async ([streamId]) => {
			if (!streamId) {
				return null;
			}
			const stream = await fetchStream(streamId);
			if (!stream) {
				return null;
			}
			const startAtMs = this.startAtSeconds !== null ? this.startAtSeconds * 1000 : 0;
			if (
				this.startAtSeconds !== null &&
				startAtMs >= 0 &&
				startAtMs <= stream.duration_milliseconds
			) {
				this.startMs = startAtMs;
			} else {
				this.startMs = 0;
			}
			this.endMs = stream.duration_milliseconds;
			return stream;
		},
	});

	connectedCallback(): void {
		super.connectedCallback();
		const urlParams = new URLSearchParams(window.location.search);
		const timeParam = urlParams.get('t');
		const parsedTime = timeParam ? parseInt(timeParam, 10) : NaN;
		this.startAtSeconds = Number.isFinite(parsedTime) ? parsedTime : null;
	}

	onDurationChangeStart(event: CustomEvent<{ value: number }>) {
		this.startMs = event.detail.value;
	}

	onDurationChangeEnd(event: CustomEvent<{ value: number }>) {
		this.endMs = event.detail.value;
	}

	setStartToNow() {
		this.startMs = this.videoElement.currentTime * 1000;
	}

	setEndToNow() {
		this.endMs = this.videoElement.currentTime * 1000;
	}

	onTimeUpdate() {
		if (!this.isClipPreview) {
			return;
		}
		const currentTimeMs = this.videoElement.currentTime * 1000;
		if (currentTimeMs >= this.endMs) {
			this.videoElement.pause();
			this.isClipPreview = false;
		}
	}

	togglePreviewClip() {
		if (!this.isValid) {
			return;
		}
		if (this.isClipPreview) {
			this.isClipPreview = false;
			this.videoElement.pause();
			return;
		}
		this.isClipPreview = true;
		this.videoElement.currentTime = this.startMs / 1000;
		this.videoElement.play()?.catch(e => {
			console.error('Error playing video:', e);
			this.isClipPreview = false;
		});
	}

	onDurationValidityChange(type: 'start' | 'end', event: CustomEvent<{ valid: boolean }>) {
		this.durationValid[type] = event.detail.valid;
		this.requestUpdate();
	}

	renderDurationFormatOptions() {
		return html`
			<div class="format-options">
				<p>??h ??m ??s / hh:mm:ss</p>
			</div>
		`;
	}

	renderStartControls(stream: StreamInfo) {
		return html`
			<div class="container chip clip-start">
				<h3>Start</h3>
				<div class="input-wrapper">
					<lb-duration-input
						.valueMs=${this.startMs}
						.maxValueMs=${stream.duration_milliseconds}
						name="start"
						@duration-change=${this.onDurationChangeStart}
						@duration-validity-change=${(e: CustomEvent<{ valid: boolean }>) => this.onDurationValidityChange('start', e)}
					></lb-duration-input>
					<button @click=${this.setStartToNow}>Set to now</button>
				</div>
				${this.renderDurationFormatOptions()}
			</div>
		`;
	}

	renderEndControls(stream: StreamInfo) {
		return html`
			<div class="container chip clip-end">
				<h3>End</h3>
				<div class="input-wrapper">
					<lb-duration-input
						.valueMs=${this.endMs}
						.maxValueMs=${stream.duration_milliseconds}
						name="end"
						@duration-change=${this.onDurationChangeEnd}
						@duration-validity-change=${(e: CustomEvent<{ valid: boolean }>) => this.onDurationValidityChange('end', e)}
					></lb-duration-input>
					<button @click=${this.setEndToNow}>Set to now</button>
				</div>
				${this.renderDurationFormatOptions()}
			</div>
		`;
	}

	renderActions() {
		return html`
			<lb-button variant="secondary" ?disabled=${!this.isValid} @click=${this.togglePreviewClip} show-icon width="full">
				${
					this.isClipPreview
						? html`
					<svg slot="icon" xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M520-200v-560h240v560H520Zm-320 0v-560h240v560H200Zm400-80h80v-400h-80v400Zm-320 0h80v-400h-80v400Zm0-400v400-400Zm320 0v400-400Z"/></svg>
					Previewing
					`
						: html`
					<svg slot="icon" xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M320-200v-560l440 280-440 280Zm80-280Zm0 134 210-134-210-134v268Z"/></svg>
					Preview
					`
				}
			</lb-button>
			<lb-button variant="secondary" ?disabled=${!this.isValid} show-icon href="${ifDefined(this.clipUrl)}" target="_blank" width="full" ?is-loading=${this.isDownloading}>
				<svg slot="icon" xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M480-320 280-520l56-58 104 104v-326h80v326l104-104 56 58-200 200ZM240-160q-33 0-56.5-23.5T160-240v-120h80v120h480v-120h80v120q0 33-23.5 56.5T720-160H240Z"/></svg>
				Download
			</lb-button>
		`;
	}

	renderView(stream: StreamInfo) {
		if (!this.clipUrl) {
			return '';
		}

		const url = `/api/v1/video/${stream.id}/index.m3u8`;

		return html`
			<lb-view-layout show-back-button>
				<span slot="title">Create a clip</span>
				<div class="container video-container">
					<lb-video .title=${stream.title} .hlsSource="${url}" autoplay @timeupdate=${this.onTimeUpdate} .startAt=${this.startMs / 1000}></lb-video>
				</div>
				<div class="container clip-info-controls">
					<h2 title="${stream.title}">${stream.title}</h2>
					<div class="clip-controls">
						<div class="clip-range">
							${this.renderStartControls(stream)}
							${this.renderEndControls(stream)}
							<div class="container chip clip-duration">
								<h3>Duration</h3>
								<lb-duration-input disabled .valueMs=${this.durationMs}></lb-duration-input>
							</div>
						</div>
						<div class="container chip actions">
							${this.renderActions()}
						</div>
					</div>
				</div>
			</lb-view-layout>
		`;
	}

	render() {
		return this._streamTask.render({
			pending: () => html`<div class="status-wrapper"><p>Loading stream information...</p></div>`,
			complete: (stream: StreamInfo | null) => {
				if (!stream) {
					return html`<div class="status-wrapper"><p>Stream not found.</p></div>`;
				}
				return this.renderView(stream);
			},
			error: e =>
				html`<div class="status-wrapper"><p>Error loading stream information: ${e instanceof Error ? e.message : 'Unknown error'}</p></div>`,
		});
	}
}

customElements.define('lb-clip-view', ClipView);
