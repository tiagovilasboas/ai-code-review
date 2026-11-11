// What every plate shares: the double-rule frame, the plate caption at the top and the source
// cartouche at the foot. Returns the band left for the figure.

import { frame } from '../pen/ornaments.js';
import { cartouche, creditLines } from '../pen/cartouche.js';

export const ROMAN = ['I', 'II', 'III', 'IV', 'V', 'VI', 'VII', 'VIII'];

export function sheet(p, s, { n, caption, refs = [] }) {
  frame(p);
  if (n) p.caps(`${s.plate} ${ROMAN[n - 1]} — ${caption}`, p.width / 2, 58, { size: 17, anchor: 'middle', track: 0.16 });
  const top = cartouche(p, [...refs, ...creditLines(s.after)], p.height - 34);
  return { top: 84, bottom: top - 16 };
}

// "Fig. n. text" in italic, centred on x.
export function figure(p, s, n, text, x, y, o = {}) {
  const str = `${s.fig} ${n}. ${text}`;
  const size = p.fit(str, o.max ?? 1100, { size: o.size ?? 18, font: 'italic' });
  p.text(str, x, y, { size, font: 'italic', anchor: 'middle' });
}

// <title> and <desc> for a plate: its own description, then the source with its address.
export function meta(s, title, desc) {
  return { title, desc: `${desc} ${s.descSource}` };
}
