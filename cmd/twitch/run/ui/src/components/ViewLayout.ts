import { css, html } from 'lit';
import { property } from 'lit/decorators.js';
import { Component } from '../litutil/Component';

export class ViewLayout extends Component {
	static styles = css`
		:host {
			display: flex;
			flex-direction: column;
			align-items: center;
			justify-content: safe center;
			width: 100%;
			height: 100%;
			padding: 1rem;
			overflow-y: auto;
			scrollbar-gutter: stable;
		}

		.view {
			display: flex;
			flex-direction: column;
			gap: 1rem;

			width: 100%;
			max-width: 700px;
		}

		h1 {
			width: fit-content;
		}

		.body {
			display: flex;
			flex-direction: column;
			gap: 1rem;
			width: 100%;
		}
	`;

	@property({ type: Boolean, attribute: 'show-back-button' })
	showBackButton: boolean = false;

	render() {
		return html`
			<div class="view" part="view">
				${this.showBackButton ? html`<lb-back-button href="/"></lb-back-button>` : ''}
				<slot name="header">
					<h1>
						<slot name="title"></slot>
					</h1>
				</slot>
				<div class="body" part="view-body">
					<slot></slot>
				</div>
			</div>
		`;
	}
}

customElements.define('lb-view-layout', ViewLayout);
