// Plate II: the seven layers drawn as strata in section, each named with what implements it in
// this repository. A defect falls through channels in I and II and is stopped on III.

import { Plate } from '../pen/plate.js';
import { arrowAlong, burst } from '../pen/ornaments.js';
import { polyD } from '../pen/geometry.js';
import { block } from '../figures/solids.js';
import { sheet, figure, meta, ROMAN } from './sheet.js';

const X0 = 90;
const W = 500;
const H = 44;
const D = 22;
const TOP = 168;
const PITCH = 64;
const SLOT = 16;
const DEFECT_X = 330;
// Channels through each stratum, as x offsets from X0. Only I and II have one on the defect's path.
const SLOTS = [[232, 420], [232, 80], [120, 400], [300, 52], [176, 444], [350], [104, 262]];
const TONES = [[1, 0], [1, 8], [2, 0], [1, -6], [2, 4], [2, -4], [3, 0]];
const SECURITY = 2;

function channel(p, x, y) {
  const dx = D;
  const dy = -D * 0.6;
  p.fill(polyD([[x, y], [x + dx, y + dy], [x + SLOT + dx, y + dy], [x + SLOT, y]]), p.tone(5, 90));
  p.fill(polyD([[x, y], [x + SLOT, y], [x + SLOT, y + H], [x, y + H]]), p.color('paper'));
  p.fill(polyD([[x + SLOT - 5, y], [x + SLOT, y], [x + SLOT, y + H], [x + SLOT - 5, y + H]]), p.tone(4, 45));
  p.swell([[x, y], [x, y + H]], 1.2);
  p.swell([[x + SLOT, y], [x + SLOT, y + H]], 1.2);
}

function strata(p) {
  SLOTS.forEach((slots, i) => {
    const y = TOP + i * PITCH;
    const ink = i === SECURITY ? 'accent' : 'ink';
    block(p, { x: X0, y, w: W, h: H, d: D, front: TONES[i], side: [3, 90], ink });
    for (const sx of slots) channel(p, X0 + sx, y);
    p.stroke(`M${X0 + W + D + 6} ${y + H / 2 - 6}H${646}`, 0.6, 'ink', ' stroke-dasharray="1 4" stroke-linecap="round"');
  });
}

function defect(p) {
  const stop = TOP + SECURITY * PITCH;
  const x = DEFECT_X + SLOT / 2 - 8;
  arrowAlong(p, [[x, 112], [x, stop - 2]], { ink: 'accent', width: 1.6 });
  burst(p, x, stop - 2, { ink: 'accent', inner: 6, outer: 14 });
  p.caps('BLOCK', x + 34, stop - 12, { size: 16, anchor: 'start', ink: 'accent', halo: 5, track: 0.12 });
}

function legend(p, s, t) {
  p.caps(t.head, 690, TOP - 26, { size: 14, track: 0.12 });
  s.layers.forEach((name, i) => {
    const y = TOP + i * PITCH;
    const ink = i === SECURITY ? 'accent' : 'ink';
    p.caps(ROMAN[i], 680, y + 14, { size: 15, anchor: 'end', ink, track: 0.04 });
    p.text(name, 690, y + 14, { size: 20, ink });
    p.text(t.impl[i], 690, y + 36, { size: p.fit(t.impl[i], 530, { size: 16, font: 'italic' }), font: 'italic' });
  });
}

export function section({ theme, fonts, s }) {
  const t = s.section;
  const p = new Plate({ id: 'sx', width: 1280, height: 800, theme, fonts, seed: 1202, lang: s.lang, ...meta(s, t.svgTitle, t.desc) });
  sheet(p, s, { n: 2, caption: t.caption });
  p.text(t.question, 640, 100, { size: p.fit(t.question, 1080, { size: 22, font: 'italic' }), font: 'italic', anchor: 'middle' });
  strata(p);
  defect(p);
  legend(p, s, t);
  figure(p, s, 2, t.figure, X0 + (W + D) / 2, 630, { max: 560, size: 17 });
  p.text(t.guard, 640, 670, { size: p.fit(t.guard, 1000, { size: 19 }), anchor: 'middle' });
  return p;
}
