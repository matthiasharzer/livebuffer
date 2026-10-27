import { css, html } from 'lit';
import { Component } from '../litutil/Component';

export class Footer extends Component {
	static styles = css`
		:host {
			display: flex;
			flex-direction: column;
			align-items: center;
			justify-content: safe center;
		}

		p {
			padding: 1rem;
			max-width: 700px;
			color: var(--text-reduced-color);
		}

		a {
			color: color-mix(in srgb, var(--primary-reduced-color) 70%, var(--text-reduced-color));
			text-decoration: none;

			transition: color 0.2s ease-in-out;

			&:hover {
				color: var(--primary-color);
			}
		}
	`;

	render() {
		return html`
			<p>this project is open source at <a href="https://github.com/matthiasharzer/livebuffer" target="_blank">github.com/matthiasharzer/livebuffer</a></p>
		`;
	}
}

customElements.define('lb-footer', Footer);
