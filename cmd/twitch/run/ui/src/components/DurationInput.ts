import { css, html } from 'lit';
import { property, state } from 'lit/decorators.js';
import { Component } from '../litutil/Component';
import { formatDuration } from '../services/formatDuration';

const DURATION_REGEX = /^(\d+h)?\s*([0-5]?[0-9]m)?\s*([0-5]?[0-9]s)?$/;
const DURATION_REGEX_HMS = /^(\d+):([0-5]?[0-9]):([0-5]?[0-9])$/;

export class DurationInput extends Component {
	static styles = css`
		:host {
			--bg-color: var(--surface-color, #2a2a2a);
			--border-color: var(--surface-border-color, #444);
			--focus-border-color: var(--primary-reduced-color, #9146ff);
			--error-border-color: var(--error-color, #ff4d4d);
		}
		input {
			background-color: var(--bg-color);
			border: 1px solid var(--border-color);
			border-radius: 4px;
			padding: 0.25rem 0.5rem;
			color: white;
			width: 150px;

			&:disabled {
				opacity: 0.5;
				cursor: not-allowed;
			}

			&:focus {
				outline: none;
				border-color: var(--focus-border-color);
				background-color: color-mix(in srgb, var(--focus-border-color) 10%, var(--bg-color));
			}

			&.invalid {
				border-color: var(--error-border-color);
				background-color: color-mix(in srgb, var(--error-border-color) 10%, var(--bg-color));
			}
		}
	`;

	@property({ type: Number, attribute: 'value-ms' })
	valueMs: number | null = 0;

	@property({ type: Boolean })
	disabled: boolean = false;

	@property({ type: Number })
	maxValueMs: number | null = null;

	@property()
	name = 'duration';

	@state()
	valid: boolean = true;

	valueToSubmit: number = 0;

	get formattedValue(): string {
		if (this.valueMs === null) {
			return '—';
		}
		return formatDuration(this.valueMs);
	}

	isValid(value: number): boolean {
		if (value < 0) {
			return false;
		}
		if (this.maxValueMs !== null && value > this.maxValueMs) {
			return false;
		}
		return true;
	}

	parseDuration(match: RegExpExecArray): number | null {
		let totalMs = 0;
		const hours = match[1] ? parseInt(match[1], 10) : 0;
		const minutes = match[2] ? parseInt(match[2], 10) : 0;
		const seconds = match[3] ? parseInt(match[3], 10) : 0;

		totalMs += hours * 60 * 60 * 1000;
		totalMs += minutes * 60 * 1000;
		totalMs += seconds * 1000;
		return totalMs;
	}

	parseDurationHMS(match: RegExpExecArray): number | null {
		const hours = parseInt(match[1], 10);
		const minutes = parseInt(match[2], 10);
		const seconds = parseInt(match[3], 10);

		if (minutes > 59 || seconds > 59) {
			return null;
		}

		let totalMs = 0;
		totalMs += hours * 60 * 60 * 1000;
		totalMs += minutes * 60 * 1000;
		totalMs += seconds * 1000;
		return totalMs;
	}

	onInput(event: InputEvent) {
		const input = event.target as HTMLInputElement;
		const value = input.value.trim();

		if (value === '') {
			this.valueToSubmit = 0;
			this.setValid(true);
			return;
		}

		const matchDuration = DURATION_REGEX.exec(value);
		const matchDurationHMS = DURATION_REGEX_HMS.exec(value);

		let totalMs: number | null = null;
		if (matchDuration) {
			totalMs = this.parseDuration(matchDuration);
		} else if (matchDurationHMS) {
			totalMs = this.parseDurationHMS(matchDurationHMS);
		} else {
			this.setValid(false);
			return;
		}

		if (totalMs === null) {
			this.setValid(false);
			return;
		}

		if (!this.isValid(totalMs)) {
			this.setValid(false);
			return;
		}

		this.valueToSubmit = totalMs;
		this.setValid(true);
	}

	submit() {
		if (!this.valid) {
			return;
		}
		this.dispatch(
			'duration-change',
			{ value: this.valueToSubmit },
			{ bubbles: true, composed: true },
		);
	}

	setValid(isValid: boolean) {
		this.valid = isValid;
		this.dispatch(
			'duration-validity-change',
			{ valid: isValid },
			{ bubbles: true, composed: true },
		);
	}

	render() {
		return html`
			<input
				.name=${this.name}
				type="text"
				part="input"
				.value=${this.formattedValue}
				@input=${this.onInput}
				@change=${this.submit}
				class=${this.valid ? 'valid' : 'invalid'}
				?disabled=${this.disabled}
			/>
		`;
	}
}

customElements.define('lb-duration-input', DurationInput);
