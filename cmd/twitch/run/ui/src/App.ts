import './components/Background.ts';
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
			enter: async () => {
				await import('./views/root/view.ts');
				return true;
			},
		},
		{
			path: '/live/:username',
			render: ({ username }) => html`<lb-live-view .username=${username ?? null}></lb-live-view>`,
			enter: async () => {
				await import('./views/live/view.ts');
				return true;
			},
		},
		{
			path: '/video/:stream_id',
			render: ({ stream_id }) =>
				html`<lb-video-view .streamId=${stream_id ?? null}></lb-video-view>`,
			enter: async () => {
				await import('./views/video/view.ts');
				return true;
			},
		},
		{
			path: '/clip/:stream_id',
			render: ({ stream_id }) => html`<lb-clip-view .streamId=${stream_id ?? null}></lb-clip-view>`,
			enter: async () => {
				await import('./views/clip/view.ts');
				return true;
			},
		},
		{
			path: '/*',
			render: () => html`<lb-not-found-view></lb-not-found-view>`,
			enter: async () => {
				await import('./views/notfound/view.ts');
				return true;
			},
		},
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
