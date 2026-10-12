import { css, html } from 'lit';
import { Component } from '../litutil/Component';

export class Spinner extends Component {
	static styles = css`
		:host {
			display: inline-block;
		}

		svg {
			animation: spin 1s linear infinite;
		}

		@keyframes spin {
			from {
				transform: rotate(0deg);
			}
			to {
				transform: rotate(360deg);
			}
		}
	`;

	render() {
		return html`
			<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" width="24px" height="24px">
				<path d="M12,4c-4.418,0-8,3.582-8,8s3.582,8,8,8s8-3.582,8-8h-2c0,3.309-2.691,6-6,6s-6-2.691-6-6s2.691-6,6-6V4z"/>
			</svg>
		`;
	}
}

customElements.define('lb-spinner', Spinner);
