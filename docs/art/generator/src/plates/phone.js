// Shared by the phone cuts of the README vignettes. A phone shows a 720-wide plate at about half
// size, so 32 px here reads like the README's own 16 px text. Nothing meant to be read is set
// smaller than that.

export const PHONE_W = 720;
export const PHONE_TEXT = 32;

// The legend under a phone cut: centred italic, wrapped. Returns the baseline of its last line.
export function phoneLegend(p, str, y) {
  return p.para(str, PHONE_W / 2, y, { size: PHONE_TEXT, font: 'italic', width: PHONE_W - 60, lead: 42, anchor: 'middle' });
}

// A phone cut is as tall as what was drawn: call last, with the lowest baseline or edge.
export function closePhone(p, bottom) {
  p.height = Math.round(bottom + 28);
  return p;
}
