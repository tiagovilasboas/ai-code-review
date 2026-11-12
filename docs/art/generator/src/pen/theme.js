// Engraving palette: one ink on paper, plus a single accent ink reserved for BLOCK / security.
// `weight` thins every line in dark mode, where light lines on black read heavier.

export const THEMES = {
  light: {
    name: 'light',
    paper: '#f4efe4',
    ink: '#1d1a16',
    accent: '#b3412c',
    weight: 1,
  },
  dark: {
    name: 'dark',
    paper: '#14130f',
    ink: '#e9e2d0',
    accent: '#e0735a',
    weight: 0.9,
  },
};

// Borderless vignettes for the README: no background rect, so the drawing sits on the page.
// `paper` is still painted where a shape must hide what is behind it, so it matches GitHub's
// page colour in each scheme.
export const BARE = {
  light: { ...THEMES.light, name: 'light', paper: '#ffffff', bare: true },
  dark: { ...THEMES.dark, name: 'dark', paper: '#0d1117', bare: true },
};
