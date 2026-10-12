import './Button';
import { css, html } from 'lit';
import { property } from 'lit/decorators.js';
import { Component } from '../litutil/Component';

export class BackButton extends Component {
	static styles = css`
		:host {
			align-self: flex-start;
		}

		lb-button {
			svg {
				transition: transform 0.2s ease;
			}

			&:hover svg {
				transform: translateX(-4px);
			}
		}
	`;

	@property({ type: String })
	href: string = '/';

	render() {
		return html`
			<lb-button variant="plain" href=${this.href} show-icon>
				<svg slot="icon" xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M400-240 160-480l240-240 56 58-142 142h486v80H314l142 142-56 58Z"/></svg>
				Go back
			</lb-button>
		`;
	}
}

customElements.define('lb-back-button', BackButton);
