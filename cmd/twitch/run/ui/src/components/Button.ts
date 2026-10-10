import './Spinner';
import { css, html } from 'lit';
import { property } from 'lit/decorators.js';
import { ifDefined } from 'lit/directives/if-defined.js';
import { Component } from '../litutil/Component';

type Variant = 'primary' | 'secondary' | 'plain';

type Width = 'auto' | 'full';

export class Button extends Component {
	static styles = css`
		:host {
			display: inline-block;
			width: var(--width, auto);
		}
		a, button {
			all: unset;

			display: flex;
			align-items: center;
			justify-content: center;
			gap: 0.5rem;

			padding: 0.25rem 0.5rem;
			border-radius: 0.25rem;
			box-sizing: border-box;

			font-size: 1rem;
			font-weight: bold;
			text-align: center;
			width: var(--width, auto);
			height: 1.7em;

			cursor: pointer;
			white-space: nowrap;

			&.disabled {
				opacity: 0.7;
				cursor: not-allowed;
				filter: grayscale(100%);
			}

			&:focus-visible {
				outline: 2px solid var(--primary-color, #cd79fd);
				outline-offset: 2px;
			}
		}

		.primary, .secondary {
			color: var(--color);
			border: 1px solid var(--color);
			transition: background-color 0.2s, color 0.2s;

			&:hover:not(.disabled) {
				background-color: var(--color);
				color: white;
			}
		}

		.primary {
			--color: var(--primary-color, #cd79fd);
		}

		.secondary {
			--color: var(--primary-reduced-color, #aaa);
		}


		.plain {
			padding: 0.25rem 0;
			color: var(--text-reduced-color, #aaa);
			font-weight: normal;
			transition: color 0.2s;

			&:hover:not(.disabled) {
				color: var(--primary-reduced-color, #cd79fd);
			}
		}

		.icon {
			height: 1.3em;
			width: 1.3em;
			display: inline-flex;
			align-items: center;
			justify-content: center;
		}
	`;

	@property({ type: String })
	variant: Variant = 'primary';

	@property({ type: Boolean })
	disabled: boolean = false;

	@property({ type: String })
	href: string | null = null;

	@property({ type: Boolean, attribute: 'show-icon' })
	showIcon: boolean = false;

	@property({ type: Boolean, attribute: 'is-loading' })
	isLoading: boolean = false;

	@property({ type: String })
	width: Width = 'auto';

	@property({ type: Boolean, attribute: 'download' })
	download: boolean = false;

	@property({ type: String })
	target: string | null = null;

	classes() {
		const classes = ['button'];

		if (this.variant) {
			classes.push(this.variant);
		}
		if (this.disabled) {
			classes.push('disabled');
		}

		return classes.join(' ');
	}

	onClick(event: Event) {
		if (this.disabled) {
			event.preventDefault();
			event.stopPropagation();
		}
	}

	renderIcon() {
		if (this.isLoading) {
			return html`<lb-spinner></lb-spinner>`;
		}

		if (this.showIcon) {
			return html`<div class="icon"><slot name="icon"></slot></div>`;
		}
		return '';
	}

	renderContent() {
		return html`
			${this.renderIcon()}
			<span class="text">
				<slot></slot>
			</span>
		`;
	}

	renderAnchor() {
		if (!this.href) {
			return '';
		}

		return html`
			<a href="${this.href}" class="${this.classes()}" aria-disabled=${this.disabled} ?download=${this.download} target=${ifDefined(this.target)} @click=${this.onClick}>
				${this.renderContent()}
			</a>
		`;
	}

	renderButton() {
		return html`
			<button class="${this.classes()}" ?disabled=${this.disabled} @click=${this.onClick}>
				${this.renderContent()}
			</button>
		`;
	}

	render() {
		return html`
			<style>
				:host {
					--width: ${this.width === 'full' ? '100%' : 'auto'};
				}
			</style>
			${this.href ? this.renderAnchor() : this.renderButton()}
		`;
	}
}

customElements.define('lb-button', Button);
