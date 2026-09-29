import type { Color, Theme, ThemeBlueprint, Variant } from './theme.ts';

export const buildTheme = (blueprint: ThemeBlueprint): Theme => {
	const resolveColor = (color: Color, seen: Set<Color> | null = null): string => {
		if (typeof color === 'function') {
			seen = seen || new Set();
			if (seen.has(color)) {
				throw new Error('Circular reference detected');
			}
			seen.add(color);
			const result = resolveColor(color(blueprint), seen);
			seen.delete(color);
			return result;
		}
		return color;
	};

	const resolveSurface = (surface: Variant<Color>): Variant<string> => {
		if (surface.ink) {
			return {
				surface: resolveColor(surface.surface),
				ink: resolveColor(surface.ink),
			};
		}
		return {
			surface: resolveColor(surface.surface),
			ink: resolveColor(blueprint.colors.defaults.ink),
		};
	};

	const resolveSurfaces = (
		surfaces: ThemeBlueprint['colors']['variants'],
	): Theme['colors']['variants'] => {
		const resolvedSurfaces: Theme['colors']['variants'] = {} as Theme['colors']['variants'];
		for (const [key, surface] of Object.entries(surfaces)) {
			resolvedSurfaces[key as keyof Theme['colors']['variants']] = resolveSurface(surface);
		}
		return resolvedSurfaces;
	};

	return {
		name: blueprint.name,
		colors: {
			defaults: {
				surface: blueprint.colors.defaults.surface,
				ink: blueprint.colors.defaults.ink,
			},
			variants: resolveSurfaces(blueprint.colors.variants),
			stats: {
				attacK: resolveColor(blueprint.colors.stats.attacK),
				defense: resolveColor(blueprint.colors.stats.defense),
				stamina: resolveColor(blueprint.colors.stats.stamina),
			},
		},
		border: {
			color: resolveColor(blueprint.border.color),
			width: blueprint.border.width,
		},
		shadow: {
			color: resolveColor(blueprint.shadow.color),
			offset: blueprint.shadow.offset,
		},
		radius: blueprint.radius,
	};
};

type ThemeValue = string | { [key: string]: ThemeValue };
export function flattenTheme(
	obj: Theme | Record<string, ThemeValue>,
	prefix: string = '-',
): Record<string, string> {
	const result: Record<string, string> = {};

	for (const [key, value] of Object.entries(obj)) {
		const newKey = prefix === '-' ? `-${prefix}${key}` : `${prefix}-${key}`;

		if (typeof value === 'object' && value !== null) {
			Object.assign(result, flattenTheme(value as Record<string, ThemeValue>, newKey));
		} else {
			result[newKey] = String(value);
		}
	}

	return result;
}
