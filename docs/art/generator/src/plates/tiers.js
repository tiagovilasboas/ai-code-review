// Plate IV: risk tiers as three pedestals of rising height. The seals on each are the human
// approvals that tier needs (0, 1, 2); the rules that set the tier stand beside them.

import { Plate } from '../pen/plate.js';
import { polyD } from '../pen/geometry.js';
import { block, medallion } from '../figures/solids.js';
import { sectionRef } from '../source.js';
import { sheet, figure, meta } from './sheet.js';

const GROUND = 420;
const XS = [110, 350, 590];
const W = 180;
const D = 26;
const HEIGHTS = [70, 140, 210];
const RULES_X = 880;

// Engraved horizon: the ground line, then fainter, shorter strokes beneath it.
function ground(p) {
  p.swell([[80, GROUND], [840, GROUND]], 1.8, 'ink', 0.05);
  [[6, 0.9, 20], [11, 0.7, 60], [16, 0.5, 110]].forEach(([dy, w, inset]) => p.swell([[80 + inset, GROUND + dy], [840 - inset, GROUND + dy]], w, 'ink', 0.05));
}

function pedestal(p, i) {
  const x = XS[i];
  const h = HEIGHTS[i];
  const y = GROUND - h;
  block(p, { x: x - 10, y: GROUND - 16, w: W + 20, h: 16, d: D, front: [2, 0] });
  block(p, { x, y, w: W, h: h - 16, d: D, front: [1, 90] });
  return y;
}

// The seals sit on the top face; a zero is drawn as an empty dotted ring.
function seals(p, i, top) {
  const cx = XS[i] + W / 2 + D / 2;
  const cy = top - D * 0.3 - 26;
  if (i === 0) {
    p.add(`<circle cx="${cx}" cy="${cy}" r="22" fill="none" stroke="${p.color()}" stroke-width="${p.w(1.3)}" stroke-dasharray="1 5" stroke-linecap="round"/>`);
    return;
  }
  const offsets = i === 1 ? [0] : [-28, 28];
  for (const dx of offsets) medallion(p, cx + dx, cy, 22);
  offsets.forEach((dx, k) => p.text(String(k + 1), cx + dx, cy + 7, { size: 19, anchor: 'middle' }));
}

function labels(p, t, i) {
  const [name, changes, approvals] = t.tiers[i];
  const cx = XS[i] + W / 2;
  p.caps(name, cx, GROUND + 46, { size: 21, anchor: 'middle', track: 0.1 });
  const end = p.para(approvals, cx, GROUND + 72, { size: 17, width: 220, lead: 21, anchor: 'middle' });
  p.para(changes, cx, end + 24, { size: 15, width: 220, lead: 19, anchor: 'middle', font: 'italic' });
}

function rules(p, t) {
  let y = 196;
  for (const rule of t.rules) {
    p.fill(polyD([[RULES_X - 16, y - 6], [RULES_X - 11, y - 10], [RULES_X - 6, y - 6], [RULES_X - 11, y - 2]]), p.color());
    y = p.para(rule, RULES_X, y, { size: 18, width: 330, lead: 23 }) + 38;
  }
}

export function tiers({ theme, fonts, s }) {
  const t = s.tiers;
  const p = new Plate({ id: 'rt', width: 1280, height: 780, theme, fonts, seed: 1404, lang: s.lang, ...meta(s, t.svgTitle, t.desc) });
  const { bottom } = sheet(p, s, { n: 4, caption: t.caption, refs: [sectionRef(5)] });
  p.text(t.lead, 640, 112, { size: p.fit(t.lead, 1000, { size: 28, font: 'italic' }), font: 'italic', anchor: 'middle' });
  ground(p);
  HEIGHTS.forEach((_, i) => {
    const top = pedestal(p, i);
    seals(p, i, top);
    labels(p, t, i);
  });
  p.swell([[RULES_X - 30, 160], [RULES_X - 30, 520]], 1.2, 'ink', 0.05);
  rules(p, t);
  figure(p, s, 4, t.figure, 640, bottom - 4, { max: 1000, size: 17 });
  return p;
}
