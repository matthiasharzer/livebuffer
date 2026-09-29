import { type CSSResultGroup, css, LitElement } from 'lit';

const cssOverwrites = css`

input {
  min-width: 0;
  width: auto;
	font-family: inherit;
}

*, *::before, *::after {
  box-sizing: border-box;
}
* {
  margin: 0;
	interpolate-size: allow-keywords;
	font-variant-ligatures: none;
	font-family: var(--font-body);
}

img, picture, video, canvas, svg {
  display: block;
  max-width: 100%;
}
input, button, textarea, select {
  font: inherit;
}
p, h1, h2, h3, h4, h5, h6 {
  overflow-wrap: break-word;
}
p {
  text-wrap: pretty;
}
h1, h2, h3, h4, h5, h6 {
  text-wrap: balance;
	font-family: var(--font-header);
}
#root, #__next {
  isolation: isolate;
}

a.watch-link {
	text-decoration: none;
	color: var(--color, #007bff);
	font-weight: bold;
	padding: 0.25rem 0.5rem;
	border: 1px solid var(--color, #007bff);
	border-radius: 0.25rem;
	transition: background-color 0.3s, color 0.3s;
	white-space: nowrap;

	&:hover {
		background-color: var(--color, #007bff);
		color: white;
	}
}
`;

export class Component extends LitElement {
	protected static _styles: CSSResultGroup;

	static get styles(): CSSResultGroup {
		const derivedStyles = Component._styles || [];
		return [cssOverwrites, ...(Array.isArray(derivedStyles) ? derivedStyles : [derivedStyles])];
	}

	static set styles(styles: CSSResultGroup) {
		Component._styles = styles;
	}

	get rect() {
		return this.getBoundingClientRect();
	}

	sleep(ms: number) {
		return new Promise(resolve => setTimeout(resolve, ms));
	}

	connectedCallback(): void {
		super.connectedCallback();
	}

	dispatch<T>(
		name: string,
		detail: T | null = null,
		options: { bubbles?: boolean; composed?: boolean } = {},
	) {
		this.dispatchEvent(new CustomEvent(name, { detail, ...options }));
	}

	dispatchBubble<T>(name: string, detail: T | null = null) {
		this.dispatch(name, detail, { bubbles: true, composed: true });
	}

	disconnectedCallback() {
		super.disconnectedCallback();
	}
}
