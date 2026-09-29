import { Observable } from '../reactive.ts';
import { buildTheme, flattenTheme } from './builder.ts';
import type { Theme } from './theme.ts';

export const defaultTheme: Theme = buildTheme({
	name: 'Default',
	colors: {
		defaults: {
			surface: '#121212',
			ink: '#f0f0f0',
		},
		variants: {
			accent: {
				surface: '#ffe275',
				ink: '#000000',
			},
			default: {
				surface: t => t.colors.defaults.surface,
				ink: t => t.colors.defaults.ink,
			},
			canvas: {
				surface: '#1c1c1c',
			},
			// Alerts kept brutally neon, adding black ink for contrast
			info: {
				surface: '#00E5FF',
				ink: '#000000',
			},
			success: {
				surface: '#00FF66',
				ink: '#000000',
			},
			warning: {
				surface: '#FFC900',
				ink: '#000000',
			},
			error: {
				surface: '#FF4911',
				ink: '#000000',
			},
			'search-box': {
				surface: '#1e3828', // Deep forest green
			},
			'search-result': {
				surface: '#4a3a00', // Deep gold
			},
			'pokemon-page': {
				surface: '#331d5a', // Deep violet
			},
			'pokemon-attributes': {
				surface: '#453903', // Deep mustard
			},
			'pokemon-cp-stats': {
				surface: '#00424a', // Deep teal
			},
			'pokemon-image': {
				surface: '#262626', // Neutral dark gray
			},
		},
		stats: {
			attacK: '#5eb0e5',
			defense: '#f2f25e',
			stamina: '#ff6b9e',
		},
	},
	border: { color: '#ffffff', width: { thick: '3px', thin: '1px' } },
	shadow: { color: '#ffffff', offset: { x: '4px', y: '4px' } },
	radius: { sharp: '0px', soft: '4px' },
});

export const themes: Record<string, Theme> = {
	default: defaultTheme,
};

const applyTheme = (theme: Theme) => {
	const flattenedTheme = flattenTheme(theme);

	const styleElementId = 'theme-styles';
	let styleElement = document.getElementById(styleElementId) as HTMLStyleElement | null;

	if (!styleElement) {
		styleElement = document.createElement('style');
		styleElement.id = styleElementId;
		document.head.appendChild(styleElement);
	}

	let cssVariables = '';
	for (const [key, value] of Object.entries(flattenedTheme)) {
		cssVariables += `${key}: ${value};\n`;
	}

	styleElement.textContent = `:root {\n${cssVariables}}`;
};

const savedThemeId = localStorage.getItem('themeId');
const initialThemeId = savedThemeId && themes[savedThemeId] ? savedThemeId : 'default';

export const currentThemeId = new Observable<string>(initialThemeId);
currentThemeId.subscribe(themeId => {
	if (themes[themeId]) {
		applyTheme(themes[themeId]);
		localStorage.setItem('themeId', themeId);
	} else {
		console.warn(`Theme "${themeId}" not found. Falling back to default theme.`);
		currentThemeId.value = 'default';
	}
}, true);
