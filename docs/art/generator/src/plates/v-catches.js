// Vignette for "O que ela pega": a magnifying glass over diff lines, one finding blocked,
// beside the risk classes the rules cover. Borderless; copy from strings.vCatches.

import { Plate } from '../pen/plate.js';
import { polyD, num } from '../pen/geometry.js';
import { burst } from '../pen/ornaments.js';
import { PHONE_W, PHONE_TEXT, phoneLegend, closePhone } from './phone.js';

// Everything in the diff is sized from the code text, and the glass from its radius, so the phone
// cut is the same drawing at reading size.
const DESK = { panel: { x: 50, y: 56, w: 560, h: 270 }, row0: 104, row: 38, text: 18, sign: 16, lens: { x: 330, r: 96, max: 32, handle: [190, 128, 168, 146] } };
const HIT = 3;

function panel(p, g) {
  const { x, y, w, h } = g.panel;
  const k = g.text / 18;
  p.swell([[x, y], [x + w, y]], 1.4, 'ink', 0.1);
  p.swell([[x, y + h], [x + w, y + h]], 1.4, 'ink', 0.1);
  p.swell([[x + 38 * k, y + 6], [x + 38 * k, y + h - 6]], 0.8, 'ink', 0.2);
}

function rows(p, t, g) {
  const { x, w } = g.panel;
  const k = g.text / 18;
  t.diff.forEach(([sign, code], i) => {
    const y = g.row0 + i * g.row;
    if (sign === '+') p.fill(polyD([[x + 40 * k, y - 22 * k], [x + w, y - 22 * k], [x + w, y + 10 * k], [x + 40 * k, y + 10 * k]]), p.tone(1, 45));
    p.text(sign === ' ' ? String(i + 1) : sign, x + 30 * k, y, { size: g.sign, anchor: 'end', font: sign === ' ' ? 'italic' : 'bold' });
    const indent = (code.length - code.trimStart().length) * 9 * k;
    p.text(code.trim(), x + 52 * k + indent, y, { size: g.text, font: sign === '-' ? 'italic' : 'roman', halo: sign === '+' ? 3 * k : 0 });
  });
}

// The glass hides what is behind it and shows the finding larger, in the accent.
function lens(p, t, g) {
  const { x, r, max, handle: [hx, hy, ix, iy] } = g.lens;
  const q = r / 96;
  const cy = g.row0 + HIT * g.row - 6 * (g.text / 18);
  const disc = (rr) => `M${num(x - rr)} ${num(cy)}a${num(rr)} ${num(rr)} 0 1 0 ${num(2 * rr)} 0a${num(rr)} ${num(rr)} 0 1 0 ${num(-2 * rr)} 0Z`;
  p.fill(polyD([[x + r * 0.62, cy + r * 0.74], [x + r * 0.82, cy + r * 0.56], [x + hx * q, cy + hy * q], [x + ix * q, cy + iy * q]]), p.tone(4, 45));
  p.swell([[x + r * 0.66, cy + r * 0.72], [x + ((hx + ix) / 2 - 1) * q, cy + ((hy + iy) / 2 + 1) * q]], 1.2, 'ink', 0.1);
  p.fill(disc(r), p.color('paper'));
  const code = t.diff[HIT][1].trim();
  const size = p.fit(code, 2 * r - 26, { size: max, font: 'bold' });
  p.text(code, x, cy + 7 * q, { size, anchor: 'middle', ink: 'accent', font: 'bold' });
  p.swell([[x - r + 18, cy + 20 * q], [x + r - 18, cy + 20 * q]], 1.2, 'accent', 0.1);
  p.fill(disc(r + 9 * q) + disc(r), p.tone(4, 45), ' fill-rule="evenodd"');
  p.stroke(disc(r + 9 * q), 1.5);
  p.stroke(disc(r), 1);
  return cy;
}

function blockBadge(p, t, cy) {
  const { x: lx, r } = DESK.lens;
  const x = lx + r + 42;
  const y = cy - r - 6;
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
  panel(p, DESK);
  rows(p, t, DESK);
  blockBadge(p, t, lens(p, t, DESK));
  tags(p, t);
  p.text(t.legend, 940, 372, { size: 18, font: 'italic', anchor: 'middle' });
  return p;
}

// Phone cut: the diff at reading size with the verdict inside the glass, then the risk names as
// one wrapped line. The hints stay in the README text.
const PHONE = { panel: { x: 24, y: 20, w: 672, h: 364 }, row0: 86, row: 54, text: PHONE_TEXT, sign: PHONE_TEXT, lens: { x: 482, r: 156, max: 40, handle: [130, 88, 112, 102] } };

// Lines of whole names joined by a middle dot, so a name never breaks in two. The lines are evened
// out: the narrowest measure that still takes no more lines than the full width does.
function nameLines(p, names, max, o) {
  const fill = (width) => names.reduce((lines, name) => {
    const last = lines[lines.length - 1];
    if (last !== undefined && p.measure(`${last} · ${name}`, o) <= width) lines[lines.length - 1] = `${last} · ${name}`;
    else lines.push(name);
    return lines;
  }, []);
  let best = fill(max);
  for (let width = max - 20; width > max / 2; width -= 20) {
    const lines = fill(width);
    if (lines.length > best.length) break;
    best = lines;
  }
  return best;
}

export function vCatchesPhone({ theme, fonts, s }) {
  const t = s.vCatches;
  const p = new Plate({ id: 'vcm', width: PHONE_W, height: 0, theme, fonts, seed: 3319, lang: s.lang, title: t.svgTitle, desc: t.desc });
  panel(p, PHONE);
  rows(p, t, PHONE);
  const cy = lens(p, t, PHONE);
  const { x } = PHONE.lens;
  const half = p.caps(t.block, x + 14, cy + 84, { size: 34, anchor: 'middle', ink: 'accent', track: 0.14 }) / 2;
  burst(p, x - half - 14, cy + 72, { inner: 5, outer: 11, rays: 8 });
  const lines = nameLines(p, t.tags.map(([name]) => name), PHONE_W - 60, { size: 34 });
  lines.forEach((line, i) => p.text(line, PHONE_W / 2, 484 + i * 48, { size: 34, anchor: 'middle' }));
  return closePhone(p, phoneLegend(p, t.legend, 484 + (lines.length - 1) * 48 + 68));
}
