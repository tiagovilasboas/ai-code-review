// Engraving palette: one ink on paper, plus a single accent ink reserved for BLOCK / security.
// The accent is a clay orange. In the light themes it is darkened until small accent text keeps
// 4.5:1 against the paper it sits on (cream for the framed plates, white for the bare ones).
// `weight` thins every line in dark mode, where light lines on black read heavier.

export const THEMES = {
  light: {
    name: 'light',
    paper: '#f4efe4',
    ink: '#1d1a16',
    accent: '#af4928',
    weight: 1,
  },
  dark: {
    name: 'dark',
    paper: '#14130f',
    ink: '#e9e2d0',
    accent: '#d97757',
    weight: 0.9,
  },
};

// Borderless vignettes for the README: no background rect, so the drawing sits on the page.
// `paper` is still painted where a shape must hide what is behind it, so it matches GitHub's
// page colour in each scheme.
export const BARE = {
  light: { ...THEMES.light, name: 'light', paper: '#ffffff', accent: '#bf502b', bare: true },
  dark: { ...THEMES.dark, name: 'dark', paper: '#0d1117', bare: true },
};
