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
