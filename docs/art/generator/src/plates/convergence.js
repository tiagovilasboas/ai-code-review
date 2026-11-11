// Plate V: convergence. A spiral closes in over three rounds on "reconsider the approach"; beside
// it, the results the originating team reported, before and after, as engraved bars.

import { Plate } from '../pen/plate.js';
import { arrowAlong } from '../pen/ornaments.js';
import { block, medallion } from '../figures/solids.js';
import { sectionRef } from '../source.js';
import { sheet, figure, meta } from './sheet.js';

const CX = 330;
const CY = 380;
const TURNS = 3;
const R_OUT = 190;
const R_IN = 78;
const BAR_X = 700;
const BAR_MAX = 360;

// Archimedean spiral from the outside in, starting at the top; one turn per round.
function spiralPoint(t) {
  const r = R_OUT - (R_OUT - R_IN) * (t / TURNS);
  const a = -Math.PI / 2 + t * 2 * Math.PI;
  return [CX + r * Math.cos(a), CY + r * Math.sin(a) * 0.92];
}

function spiral(p, t) {
  const steps = 180;
  const points = [];
  for (let k = 0; k <= steps; k++) points.push(spiralPoint((k / steps) * TURNS * 0.985));
  for (let k = 0; k < steps; k += 15) p.swell(points.slice(k, k + 16), 2.2 - (k / steps) * 1.2, 'ink', 0.6, 1);
  arrowAlong(p, points.slice(-3), { width: 1.1 });
  // Each round is marked where its turn crosses the upper right, with a lettered leader.
  t.rounds.forEach((label, i) => {
    const [x, y] = spiralPoint(i + 0.125);
    p.stroke(`M${x + 14} ${y}H${CX + R_OUT + 34}`, 0.6, 'ink', ' stroke-dasharray="1 3.5" stroke-linecap="round"');
    medallion(p, x, y, 12, { rim: 0.3, shadow: false });
    p.text(String(i + 1), x, y + 5, { size: 14, anchor: 'middle' });
    p.caps(label, CX + R_OUT + 40, y + 5, { size: 14, track: 0.08 });
  });
  const lines = p.wrap(t.centre, 110, { size: 18, font: 'italic' });
  lines.forEach((line, i) => p.text(line, CX, CY + 6 - ((lines.length - 1) * 22) / 2 + i * 22, { size: 18, font: 'italic', anchor: 'middle' }));
}

function bars(p, t) {
  p.caps(t.head, BAR_X, 170, { size: 15, track: 0.12 });
  const swatch = (x, tone, label) => {
    block(p, { x, y: 190, w: 26, h: 12, d: 5, front: tone });
    p.text(label, x + 38, 202, { size: 15, font: 'italic' });
  };
  swatch(BAR_X, [1, 45], t.before);
  swatch(BAR_X + 140, [5, 45], t.after);
  t.metrics.forEach(([label, before, after, b, a], i) => {
    const y = 262 + i * 108;
    p.text(label, BAR_X, y, { size: 19 });
    const max = Math.max(a, b);
    [[b, before, [1, 45]], [a, after, [5, 45]]].forEach(([v, text, tone], k) => {
      const w = (BAR_MAX * v) / max;
      block(p, { x: BAR_X, y: y + 14 + k * 30, w, h: 18, d: 6, front: tone });
      p.text(text, BAR_X + w + 18, y + 29 + k * 30, { size: 17 });
    });
  });
  p.para(t.note, BAR_X, 612, { size: 15.5, width: 470, lead: 20, font: 'italic' });
}

export function convergence({ theme, fonts, s }) {
  const t = s.convergence;
  const p = new Plate({ id: 'cv', width: 1280, height: 840, theme, fonts, seed: 1505, lang: s.lang, ...meta(s, t.svgTitle, t.desc) });
  const { bottom } = sheet(p, s, { n: 5, caption: t.caption, refs: [sectionRef(2)] });
  p.para(t.problem, 90, 112, { size: 18, width: 520, lead: 23, font: 'italic' });
  spiral(p, t);
  p.para(t.cap, CX, 612, { size: 17, width: 480, lead: 21, anchor: 'middle' });
  bars(p, t);
  figure(p, s, 5, t.figure, 640, bottom - 4, { max: 1000, size: 17 });
  return p;
}
