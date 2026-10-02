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

		lb-background {
			z-index: -1;
		}
	`;

	private router = new Router(this, [
		{
			path: '/',
			render: () => html`<lb-root-view></lb-root-view>`,
		},
		{
			path: '/live/:username',
			render: ({ username }) => html`<lb-live-view .username=${username ?? null}></lb-live-view>`,
		},
		{
			path: '/video/:stream_id',
			render: ({ stream_id }) =>
				html`<lb-video-view .streamId=${stream_id ?? null}></lb-video-view>`,
		},
		{
			path: '/clip/:stream_id',
			render: ({ stream_id }) => html`<lb-clip-view .streamId=${stream_id ?? null}></lb-clip-view>`,
		},
		{ path: '/*', render: () => html`<lb-not-found-view></lb-not-found-view>` },
	]);

	render() {
		return html`
			<lb-background></lb-background>
			<main>
				${this.router.outlet()}
			</main>
		`;
	}
}

customElements.define('lb-app', App);
