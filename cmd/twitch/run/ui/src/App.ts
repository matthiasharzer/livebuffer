import { css, html, type PropertyValues } from 'lit';
import { Component } from './litutil/Component.ts';
import { router, setRoute } from './services/router.ts';

const routes = [{ path: '/watch', component: 'lb-watch-view' }];

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

		const currentPath = window.location.pathname;
		const matchedRoute = routes.find(route => route.path === currentPath);
		if (matchedRoute) {
			setRoute(currentPath);
		} else {
			setRoute('/watch');
		}
	}

	render() {
		return html`
			<output id="router-outlet"></output>
		`;
	}
}

customElements.define('lb-app', App);
