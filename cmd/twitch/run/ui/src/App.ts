import { css, html, type PropertyValues } from 'lit';
import { Component } from './litutil/Component.ts';
import { router } from './services/router.ts';

const routes = [{ path: '/:username/watch', component: 'lb-watch-view' }];

export class App extends Component {
	static styles = css`
		:host {
			display: flex;
			flex-direction: column;
			align-items: center;
			width: 100%;
			height: 100%;
		}

		output {
			width: 100%;
			height: 100%;
		}
	`;

	protected firstUpdated(_changedProperties: PropertyValues): void {
		super.firstUpdated(_changedProperties);

		const routerOutlet = this.renderRoot.querySelector('#router-outlet') as HTMLElement;
		if (!routerOutlet) {
			console.error('Router outlet not found');
			return;
		}

		router.setOutlet(routerOutlet);
		router.setRoutes(routes);
	}

	render() {
		return html`
			<output id="router-outlet"></output>
		`;
	}
}

customElements.define('lb-app', App);
