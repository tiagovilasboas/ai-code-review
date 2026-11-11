// Plate I, the repository banner: title, tagline, the basis of the process, the seven layers as a
// legend, and the engraved swiss-cheese. One defect is stopped at III (security), another at V.

import { Plate } from '../pen/plate.js';
import { rule } from '../pen/ornaments.js';
import { cheese } from '../figures/cheese.js';
import { sheet, figure, meta, ROMAN } from './sheet.js';

const SECURITY = 2;
const COLUMN = 600;

function drawText(p, s, t) {
  const titleWidth = p.text(t.title, 72, 136, { size: 68 });
  rule(p, 74, 74 + titleWidth, 156);
  p.text(t.tagline, 74, 194, { size: p.fit(t.tagline, COLUMN, { size: 28, font: 'italic' }), font: 'italic' });
  p.text(s.basis, 75, 228, { size: p.fit(s.basis, COLUMN, { size: 20, font: 'italic' }), font: 'italic' });
  p.text(s.independent, 76, 254, { size: p.fit(s.independent, COLUMN, { size: 15 }) });
  s.layers.forEach((label, i) => {
    const y = 300 + i * 26;
    const ink = i === SECURITY ? 'accent' : 'ink';
    p.caps(ROMAN[i], 104, y, { size: 15, anchor: 'end', ink, track: 0.04 });
    p.text(label, 116, y, { size: 19, ink });
  });
}

export function hero({ theme, fonts, s }) {
  const t = s.hero;
  const p = new Plate({ id: 'hr', width: 1280, height: 600, theme, fonts, seed: 1101, lang: s.lang, ...meta(s, t.svgTitle, t.desc) });
  sheet(p, s, { n: 1, caption: t.caption });
  drawText(p, s, t);
  cheese(p, { yb: 400 });
  figure(p, s, 1, t.figure, 960, 476, { max: 560, size: 17 });
  return p;
}
