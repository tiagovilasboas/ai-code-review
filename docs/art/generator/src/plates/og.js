// The social card (1280 × 640): title, tagline, basis, a report slip from the golden sample, and the
// swiss-cheese with the one defect that security stops. Dark only; no plate number.

import { Plate } from '../pen/plate.js';
import { rule } from '../pen/ornaments.js';
import { cheese, DEFECTS } from '../figures/cheese.js';
import { sheet, meta } from './sheet.js';

const COLUMN = 620;
const STOPPED = [0, 3];

// A report slip: paper with a hatched drop shadow and a double rule; the decision and action in accent.
function slip(p, lines, x, y) {
  const w = 430;
  const h = lines.length * 30 + 30;
  const box = (dx, dy, ww, hh) => `M${x + dx} ${y + dy}h${ww}v${hh}h${-ww}Z`;
  p.fill(box(7, 7, w, h), p.tone(4, 45));
  p.fill(box(0, 0, w, h), p.color('paper'));
  p.stroke(box(0, 0, w, h), 1.4);
  p.stroke(box(5, 5, w - 10, h - 10), 0.5);
  lines.forEach((line, i) => p.text(line, x + 26, y + 40 + i * 30, { size: 21, ink: STOPPED.includes(i) ? 'accent' : 'ink' }));
}

export function og({ theme, fonts, s }) {
  const t = s.og;
  const p = new Plate({ id: 'og', width: 1280, height: 640, theme, fonts, seed: 1909, lang: s.lang, ...meta(s, t.svgTitle, t.desc) });
  sheet(p, s, {});
  const titleWidth = p.text(t.title, 72, 140, { size: 84 });
  rule(p, 74, 74 + titleWidth, 164);
  p.text(t.tagline, 74, 206, { size: p.fit(t.tagline, COLUMN, { size: 30, font: 'italic' }), font: 'italic' });
  p.text(s.basis, 75, 240, { size: p.fit(s.basis, COLUMN, { size: 20, font: 'italic' }), font: 'italic' });
  p.text(s.independent, 76, 266, { size: p.fit(s.independent, COLUMN, { size: 15 }) });
  slip(p, t.slip, 76, 300);
  cheese(p, { yb: 452, from: 590, defects: [DEFECTS[0]] });
  return p;
}
