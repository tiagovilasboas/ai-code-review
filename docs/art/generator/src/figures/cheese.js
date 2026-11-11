// The engraved swiss-cheese model: seven upright slices on a stippled ground, and defects that
// travel through the holes until one slice stops them. Used by the hero banner and the social card.

import { arrow, burst } from '../pen/ornaments.js';
import { slab, holeAt, drawSlab, drawShadow, maskMarkup } from './slab.js';

const ROMAN = ['I', 'II', 'III', 'IV', 'V', 'VI', 'VII'];
const SECURITY = 2;

// Decorative holes per slice, in face units {u, v, rx, ry}; they never sit on a stopped path.
const EXTRA_HOLES = [
  [{ u: 36, v: 168, rx: 5, ry: 7 }],
  [{ u: 38, v: 30, rx: 6, ry: 10 }],
  [{ u: 40, v: 22, rx: 5, ry: 8 }, { u: 14, v: 168, rx: 5, ry: 7 }],
  [{ u: 16, v: 46, rx: 8, ry: 13 }, { u: 40, v: 104, rx: 5, ry: 8 }],
  [{ u: 24, v: 58, rx: 9, ry: 15 }, { u: 14, v: 166, rx: 5, ry: 8 }],
  [{ u: 30, v: 40, rx: 7, ry: 12 }, { u: 18, v: 150, rx: 9, ry: 15 }],
  [{ u: 20, v: 96, rx: 9, ry: 15 }, { u: 36, v: 30, rx: 5, ry: 8 }],
];

// Defects: height above the ground (rise), depth in the slab (u), the slice that stops them.
export const DEFECTS = [
  { rise: 124, u: 22, stopAt: 2, ink: 'accent', label: 'BLOCK' },
  { rise: 60, u: 30, stopAt: 4, ink: 'ink' },
];

function drawGround(p, x0, yb) {
  p.add(`<ellipse cx="${x0 + 195}" cy="${yb - 12}" rx="282" ry="34" fill="${p.stipple(150)}"/>`);
  p.add(`<ellipse cx="${x0 + 195}" cy="${yb - 12}" rx="246" ry="24" fill="${p.tone(1, 0)}"/>`);
}

function drawDefect(p, slices, holesBy, d, i, from) {
  const maskId = `${p.id}m${i}`;
  const parts = slices.slice(0, d.stopAt).map((s, k) => maskMarkup(s, d.u, holesBy[k].find((h) => h.defect === i)));
  p.def(maskId, `<mask id="${maskId}" maskUnits="userSpaceOnUse" x="0" y="0" width="${p.width}" height="${p.height}"><rect width="${p.width}" height="${p.height}" fill="#fff"/>${parts.join('')}</mask>`);
  const y = slices[0].yb - d.rise;
  const tip = [slices[d.stopAt].xf - d.u, y];
  arrow(p, [from, y], tip, { ink: d.ink, width: 1.5, extra: ` mask="url(#${maskId})"` });
  burst(p, tip[0], tip[1], { ink: d.ink });
  if (d.label) p.caps(d.label, tip[0] + 2, y - 24, { size: 17, anchor: 'middle', ink: d.ink, halo: 5, track: 0.12 });
}

// x0: front edge of slice I; yb: ground line; from: where the defect arrows start.
export function cheese(p, { x0 = 758, yb, from = 656, defects = DEFECTS }) {
  drawGround(p, x0, yb);
  const slices = ROMAN.map((_, i) => slab({ xf: x0 + i * 72, yb, h: 180 }));
  const holesBy = slices.map((s, k) => {
    const passing = defects.map((d, i) => (k < d.stopAt ? { ...holeAt(s, s.xf - d.u, yb - d.rise, 9, 16), defect: i } : null));
    return [...passing.filter(Boolean), ...EXTRA_HOLES[k]];
  });
  slices.forEach((s, k) => {
    drawShadow(p, s);
    drawSlab(p, s, holesBy[k]);
    p.caps(ROMAN[k], s.xf - 21, yb + 38, { size: 18, anchor: 'middle', ink: k === SECURITY ? 'accent' : 'ink', track: 0.04 });
  });
  defects.forEach((d, i) => drawDefect(p, slices, holesBy, d, i, from));
}
