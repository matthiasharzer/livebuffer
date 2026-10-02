import { css, html } from 'lit';
import { Component } from '../litutil/Component.ts';

export class Background extends Component {
	static styles = css`
		:host {
			position: fixed;
			--background: 	#121212;
			--dots: #f0f0f0;

			background-color: var(--background);
			top: 0;
			left: 0;
			right: 0;
			bottom: 0;

			overflow: hidden;
		}

		.dotted-overlay {
			position: absolute;
			top: -100%;
  		left: -100%;
			width: 300%;
  		height: 300%;

			transform: rotate(-45deg);

			 --circle-diameter: 2px;
			 --circle-spacing: 40px;
			 --circle-color: color-mix(in srgb, var(--dots) 30%, black);
			background : radial-gradient(
				circle at
						var(--circle-diameter)
						var(--circle-diameter),
				var(--circle-color) calc(var(--circle-diameter) - 1px),
				transparent var(--circle-diameter)
			)
			0 0 / var(--circle-spacing) var(--circle-spacing);
		}
	`;

	render() {
		return html`
		<div class="dotted-overlay"></div>
		`;
	}
}

customElements.define('lb-background', Background);
