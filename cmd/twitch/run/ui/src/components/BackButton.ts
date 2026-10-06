import { css, html } from 'lit';
import { property } from 'lit/decorators.js';
import { Component } from '../litutil/Component';

export class BackButton extends Component {
	static styles = css`
		:host {
			align-self: flex-start;
		}

		.back-button {
			display: flex;
			align-items: center;
			gap: 0.5rem;
			color: var(--text-reduced-color, #aaa);
			text-decoration: none;
			transition: color 0.3s;

			svg {
				transition: transform 0.2s ease;
			}

			&:hover {
				color: var(--primary-reduced-color, #cd79fd);
				svg {
					transform: translateX(-4px);
				}
			}
		}
	`;

	@property({ type: String })
	href: string = '/';

	render() {
		return html`
			<a href="${this.href}" class="back-button">
				<svg xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M400-240 160-480l240-240 56 58-142 142h486v80H314l142 142-56 58Z"/></svg>
				Go back
			</a>
		`;
	}
}

customElements.define('lb-back-button', BackButton);
