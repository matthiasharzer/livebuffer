import './Button';
import { css, html } from 'lit';
import { property } from 'lit/decorators.js';
import { Component } from '../litutil/Component';
import { formatDuration } from '../services/formatDuration';
import type { MutedSegment } from '../services/streams';

export class MutedSegmentsBox extends Component {
	static styles = css`
		.muted-segments-box {
			display: flex;
			flex-direction: column;
			justify-content: space-between;
			flex-wrap: wrap;

			gap: 0.5rem;
			padding: 1rem;

			p {
				color: var(--text-reduced-color, #aaa);
			}
		}

		.muted-segments {
			display: flex;
			flex-direction: row;
			justify-content: flex-start;
			flex-wrap: wrap;

			gap: 0.5rem;
		}
	`;

	@property({ attribute: false })
	mutedSegments: MutedSegment[] | null = null;

	jumpToSegment(segment: MutedSegment) {
		const startSeconds = segment.start_time_milliseconds / 1000;
		this.dispatch('jump-to-segment', { startSeconds }, { bubbles: true, composed: true });
	}

	renderMutedSegment(segment: MutedSegment) {
		const start = formatDuration(segment.start_time_milliseconds);
		const end = formatDuration(segment.end_time_milliseconds);

		return html`
			<lb-button variant="primary" @click=${() => this.jumpToSegment(segment)}>
				${start} – ${end}
			</lb-button>
		`;
	}

	render() {
		if (!this.mutedSegments) {
			return '';
		}
		return html`
			<div class="container muted-segments-box">
				<h2>Muted Segments</h2>
				<div class="muted-segments">
					${this.mutedSegments.map(segment => this.renderMutedSegment(segment))}
				</div>
				<p class="hint">The now available Twitch VOD contains one or more muted segments. Click on a segment to jump to it.</p>
			</div>
		`;
	}
}

customElements.define('lb-muted-segments-box', MutedSegmentsBox);
