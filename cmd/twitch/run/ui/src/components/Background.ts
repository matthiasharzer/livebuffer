import { css, html } from 'lit';
import { Component } from '../litutil/Component.ts';

export class Background extends Component {
	static styles = css`
		:host {
			position: absolute;
			background-color: var(--colors-variants-canvas-surface);
			top: 0;
			left: 0;
			right: 0;
			bottom: 0;

			overflow: hidden;
		}

		.dotted-overlay {
			position: absolute;
			top: -50%;
  		left: -50%;
			width: 200%;
  		height: 200%;

			transform: rotate(-45deg);

			 --circle-diameter: 2px;
			 --circle-spacing: 40px;
			 --circle-color: color-mix(in srgb, var(--colors-variants-canvas-ink) 30%, black);
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
