export type Color = string | ((theme: ThemeBuilder<Color>) => Color);

export interface Variant<TColor> {
	surface: TColor;
	ink?: TColor;
}

export interface ThemeBuilder<TColor> {
	name: string;
	colors: {
		defaults: {
			surface: string;
			ink: string;
		};
		variants: {
			accent: Variant<TColor>;
			default: Variant<TColor>;
			canvas: Variant<TColor>;
			info: Variant<TColor>;
			success: Variant<TColor>;
			warning: Variant<TColor>;
			error: Variant<TColor>;
			'search-box': Variant<TColor>;
			'search-result': Variant<TColor>;
			'pokemon-page': Variant<TColor>;
			'pokemon-attributes': Variant<TColor>;
			'pokemon-cp-stats': Variant<TColor>;
			'pokemon-image': Variant<TColor>;
		};
		stats: {
			attacK: TColor;
			defense: TColor;
			stamina: TColor;
		};
	};
	border: {
		color: TColor;
		width: {
			thick: string;
			thin: string;
		};
	};
	shadow: {
		color: TColor;
		offset: {
			x: string;
			y: string;
		};
	};
	radius: {
		sharp: string;
		soft: string;
	};
}

export type ThemeBlueprint = ThemeBuilder<Color>;
export type Theme = ThemeBuilder<string>;

export type VariantName = keyof Theme['colors']['variants'];
export type BorderWidthName = keyof Theme['border']['width'];
export type RadiusName = keyof Theme['radius'];
