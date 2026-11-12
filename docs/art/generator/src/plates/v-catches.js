// Vignette for "O que ela pega": a magnifying glass over diff lines, one finding blocked,
// beside the risk classes the rules cover. Borderless; copy from strings.vCatches.

import { Plate } from '../pen/plate.js';
import { polyD, num } from '../pen/geometry.js';
import { burst } from '../pen/ornaments.js';

const PANEL = { x: 50, y: 56, w: 560, h: 270 };
const ROW0 = 104;
const ROW = 38;
const HIT = 3;
const LENS = { x: 330, r: 96 };

function panel(p) {
  const { x, y, w, h } = PANEL;
  p.swell([[x, y], [x + w, y]], 1.4, 'ink', 0.1);
  p.swell([[x, y + h], [x + w, y + h]], 1.4, 'ink', 0.1);
  p.swell([[x + 38, y + 6], [x + 38, y + h - 6]], 0.8, 'ink', 0.2);
}

function rows(p, t) {
  const { x, w } = PANEL;
  t.diff.forEach(([sign, code], i) => {
    const y = ROW0 + i * ROW;
    if (sign === '+') p.fill(polyD([[x + 40, y - 22], [x + w, y - 22], [x + w, y + 10], [x + 40, y + 10]]), p.tone(1, 45));
    p.text(sign === ' ' ? String(i + 1) : sign, x + 30, y, { size: 16, anchor: 'end', font: sign === ' ' ? 'italic' : 'bold' });
    const indent = (code.length - code.trimStart().length) * 9;
    p.text(code.trim(), x + 52 + indent, y, { size: 18, font: sign === '-' ? 'italic' : 'roman', halo: sign === '+' ? 3 : 0 });
  });
}

// The glass hides what is behind it and shows the finding larger, in the accent.
function lens(p, t) {
  const cy = ROW0 + HIT * ROW - 6;
  const { x, r } = LENS;
  const disc = (rr) => `M${num(x - rr)} ${num(cy)}a${num(rr)} ${num(rr)} 0 1 0 ${num(2 * rr)} 0a${num(rr)} ${num(rr)} 0 1 0 ${num(-2 * rr)} 0Z`;
  p.fill(polyD([[x + r * 0.62, cy + r * 0.74], [x + r * 0.82, cy + r * 0.56], [x + 190, cy + 128], [x + 168, cy + 146]]), p.tone(4, 45));
  p.swell([[x + r * 0.66, cy + r * 0.72], [x + 178, cy + 138]], 1.2, 'ink', 0.1);
  p.fill(disc(r), p.color('paper'));
  const code = t.diff[HIT][1].trim();
  const size = p.fit(code, 2 * r - 26, { size: 32, font: 'bold' });
  p.text(code, x, cy + 7, { size, anchor: 'middle', ink: 'accent', font: 'bold' });
  p.swell([[x - r + 18, cy + 20], [x + r - 18, cy + 20]], 1.2, 'accent', 0.1);
  p.fill(disc(r + 9) + disc(r), p.tone(4, 45), ' fill-rule="evenodd"');
  p.stroke(disc(r + 9), 1.5);
  p.stroke(disc(r), 1);
  return cy;
}

function blockBadge(p, t, cy) {
  const x = LENS.x + LENS.r + 42;
  const y = cy - LENS.r - 6;
  p.text(t.block, x, y, { size: 20, font: 'bold', ink: 'accent', track: 0.14 });
  p.swell([[x, y + 8], [x + p.measure(t.block, { size: 20, font: 'bold', track: 0.14 }), y + 8]], 1.4, 'accent', 0.1);
  burst(p, x - 18, y - 6, { inner: 5, outer: 11, rays: 8 });
}

function tags(p, t) {
  t.tags.forEach(([name, hint], i) => {
    const x = 690 + (i % 2) * 300;
    const y = 96 + Math.floor(i / 2) * 66;
    p.fill(polyD([[x - 12, y - 6], [x - 7, y - 11], [x - 2, y - 6], [x - 7, y - 1]]), p.color());
    p.caps(name, x + 6, y, { size: 19, track: 0.08 });
    p.text(hint, x + 6, y + 24, { size: p.fit(hint, 270, { size: 16, font: 'italic' }), font: 'italic' });
  });
  p.swell([[650, 70], [650, 316]], 1, 'ink', 0.05);
}

export function vCatches({ theme, fonts, s }) {
  const t = s.vCatches;
  const p = new Plate({ id: 'vc', width: 1280, height: 400, theme, fonts, seed: 3319, lang: s.lang, title: t.svgTitle, desc: t.desc });
  panel(p);
  rows(p, t);
  blockBadge(p, t, lens(p, t));
  tags(p, t);
  p.text(t.legend, 940, 372, { size: 18, font: 'italic', anchor: 'middle' });
  return p;
}
