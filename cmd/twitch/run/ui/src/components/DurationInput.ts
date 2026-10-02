import { css, html } from 'lit';
import { property, state } from 'lit/decorators.js';
import { Component } from '../litutil/Component';
import { formatDuration } from '../services/formatDuration';

const DURATION_REGEX = /^(\d+h)*\s*([0-5]?[0-9]m)?\s*([0-5]?[0-9]s)?$/;

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
	valueMs: number = 0;

	@property({ type: Boolean })
	disabled: boolean = false;

	@state()
	valid: boolean = true;

	valueToSubmit: number = 0;

	get formattedValue(): string {
		return formatDuration(this.valueMs);
	}

	onInput(event: InputEvent) {
		const input = event.target as HTMLInputElement;
		const value = input.value.trim();

		if (value === '') {
			this.valueToSubmit = 0;
			this.valid = true;
			return;
		}

		const match = DURATION_REGEX.exec(value);
		if (!match) {
			this.valid = false;
			return;
		}

		let totalMs = 0;
		const hours = match[1] ? parseInt(match[1], 10) : 0;
		const minutes = match[2] ? parseInt(match[2], 10) : 0;
		const seconds = match[3] ? parseInt(match[3], 10) : 0;

		totalMs += hours * 60 * 60 * 1000;
		totalMs += minutes * 60 * 1000;
		totalMs += seconds * 1000;

		if (totalMs < 0) {
			this.valid = false;
			return;
		}

		this.valueToSubmit = totalMs;
		this.valid = true;
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

	render() {
		return html`
			<input
				name="duration"
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
