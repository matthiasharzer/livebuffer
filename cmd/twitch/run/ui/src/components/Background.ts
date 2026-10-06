import { css, html } from 'lit';
import { Component } from '../litutil/Component.ts';

export class Background extends Component {
	static styles = css`
		@keyframes pulse {
			0% {
				opacity: 0;
			}
			40% {
				opacity: 1;
			}
			80% {
				opacity: 0;
			}
			100% {
				opacity: 0;
			}
		}

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

		.background {
			position: absolute;
			top: -100%;
			left: -100%;
			width: 300%;
			height: 300%;
			transform: rotate(-45deg);

			--circle-diameter: 2px;
			--circle-spacing: 40px;

		}

		.dotted-overlay {
			position: absolute;
			width: 100%;
			height: 100%;

			background : radial-gradient(
				circle at
						var(--circle-diameter)
						var(--circle-diameter),
				var(--circle-color) calc(var(--circle-diameter) - 1px),
				transparent var(--circle-diameter)
			)
			0 0 / var(--circle-spacing) var(--circle-spacing);
		}

		.background .level-1 {
			--circle-color: color-mix(in srgb, var(--dots) 30%, black);

			opacity: 1;
			animation: pulse 10s infinite;
			animation-delay: -4s;
		}

		.background .level-2 {
			--circle-color: color-mix(in srgb, var(--primary-color) 70%, black);

			transform: translate(20px, 20px);
			opacity: 0;

			animation: pulse 10s infinite;
			animation-delay: -9s;
		}

		@media (prefers-reduced-motion: reduce) {
			.background .level-1,
			.background .level-2 {
				animation: none;
			}
		}
	`;

	render() {
		return html`
		<div class="background">
			<div class="dotted-overlay level-1"></div>
			<div class="dotted-overlay level-2"></div>
		</div>
		`;
	}
}

customElements.define('lb-background', Background);
