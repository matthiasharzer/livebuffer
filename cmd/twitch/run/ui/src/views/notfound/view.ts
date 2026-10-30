import { css, html } from 'lit';
import { Component } from '../../litutil/Component';
import '../../components/BackButton';
import '../../components/ViewLayout';

export class NotFound extends Component {
	static styles = css`
		.view {
			display: flex;
			flex-direction: column;
			justify-content: center;
			align-items: center;
			gap: 1rem;
			width: 100%;
			height: 100%;
			font-size: 1.5rem;
		}

		lb-back-button {
			align-self: center;
		}
	`;

	render() {
		return html`
			<div class="view">
				<p>404 - Page not found</p>
				<lb-back-button href="/"></lb-back-button>
			</div>
		`;
	}
}

customElements.define('lb-not-found-view', NotFound);
