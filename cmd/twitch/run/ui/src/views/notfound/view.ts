import { css, html } from 'lit';
import { Component } from '../../litutil/Component';

export class NotFound extends Component {
	static styles = css`
		.status-wrapper {
			display: flex;
			justify-content: center;
			align-items: center;
			width: 100%;
			height: 100%;
			font-size: 1.5rem;
		}
	`;

	render() {
		return html`
			<div class="status-wrapper">
				<p>404 - Page not found</p>
			</div>
		`;
	}
}

customElements.define('lb-not-found-view', NotFound);
