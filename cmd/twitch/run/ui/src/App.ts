import { Router } from '@lit-labs/router';
import { css, html } from 'lit';
import { Component } from './litutil/Component.ts';
import 'urlpattern-polyfill';

export class App extends Component {
	static styles = css`
		:host {
			display: flex;
			flex-direction: column;
			align-items: center;
			width: 100%;
			height: 100%;
		}

		main {
			width: 100%;
			height: 100%;
		}
	`;

	private router = new Router(this, [
		{ path: '/:username/live', render: ({ username }) => html`<lb-live-view .username=${username ?? null}></lb-live-view>` },
		{ path: '/*', render: () => html`<lb-not-found-view></lb-not-found-view>` },
	])

	render() {
		return html`
			<main>
				${this.router.outlet()}
			</main>
		`;
	}
}

customElements.define('lb-app', App);
