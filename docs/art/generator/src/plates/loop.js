// Plate VII: the learning loop. A wheel of five stations turns around the bench: an incident, the
// question of which layer should have caught it, the rule it becomes, a bench case, and the score.

import { Plate } from '../pen/plate.js';
import { arrowAlong } from '../pen/ornaments.js';
import { medallion } from '../figures/solids.js';
import { sectionRef } from '../source.js';
import { sheet, figure, meta } from './sheet.js';

const CX = 640;
const CY = 400;
const R = 196;
const HUB = 74;
const NODE = 26;
const TEXT_W = 260;
const angleOf = (k) => -Math.PI / 2 + (k * 2 * Math.PI) / 5;
const at = (a, r = R) => [CX + r * Math.cos(a), CY + r * Math.sin(a)];

function wheel(p) {
  const ring = (r) => `M${CX - r} ${CY}a${r} ${r} 0 1 0 ${2 * r} 0a${r} ${r} 0 1 0 ${-2 * r} 0Z`;
  p.fill(ring(R + 11) + ring(R - 11), p.tone(1, 0), ' fill-rule="evenodd"');
  for (const r of [R - 11, R + 11]) p.add(`<circle cx="${CX}" cy="${CY}" r="${r}" fill="none" stroke="${p.color()}" stroke-width="${p.w(0.5)}"/>`);
  // Spokes between the stations, from the hub to the inner rim.
  for (let k = 0; k < 5; k++) {
    const a = angleOf(k + 0.5);
    p.swell([at(a, HUB + 4), at(a, R - 11)], 3, 'ink', 0.35);
  }
  // The rim turns clockwise: an arc arrow from each station to the next.
  const gap = (NODE + 10) / R;
  for (let k = 0; k < 5; k++) {
    const a1 = angleOf(k) + gap;
    const a2 = angleOf(k + 1) - gap;
    const points = Array.from({ length: 25 }, (_, i) => at(a1 + ((a2 - a1) * i) / 24));
    arrowAlong(p, points, { width: 1.5 });
  }
}

function stationText(p, [name, detail], k, [x, y]) {
  const a = angleOf(k);
  const c = Math.cos(a);
  const nameO = { size: 19 };
  const detailO = { size: 15, font: 'italic' };
  if (Math.abs(c) < 0.2) {
    const lines = p.wrap(detail, 520, detailO);
    const base = y - NODE - 18 - (lines.length - 1) * 19;
    p.text(name, x, base - 24, { ...nameO, anchor: 'middle' });
    lines.forEach((line, i) => p.text(line, x, base + i * 19, { ...detailO, anchor: 'middle' }));
    return;
  }
  const anchor = c > 0 ? 'start' : 'end';
  const tx = x + Math.sign(c) * (NODE + 16);
  const nameEnd = p.para(name, tx, y - 2, { ...nameO, anchor, width: TEXT_W, lead: 23 });
  p.para(detail, tx, nameEnd + 22, { ...detailO, anchor, width: TEXT_W, lead: 19 });
}

export function loop({ theme, fonts, s }) {
  const t = s.loop;
  const p = new Plate({ id: 'll', width: 1280, height: 820, theme, fonts, seed: 1707, lang: s.lang, ...meta(s, t.svgTitle, t.desc) });
  sheet(p, s, { n: 7, caption: t.caption, refs: [sectionRef(7)] });
  wheel(p);
  medallion(p, CX, CY, HUB, { rim: 0.16, shadow: false });
  p.text(t.centre, CX, CY + 2, { size: 22, anchor: 'middle' });
  p.text(t.centreSub, CX, CY + 24, { size: 15, font: 'italic', anchor: 'middle' });
  t.stations.forEach((station, k) => {
    const [x, y] = at(angleOf(k));
    medallion(p, x, y, NODE, { rim: 0.24, shadow: false });
    p.text(String(k + 1), x, y + 6, { size: 18, anchor: 'middle' });
    stationText(p, station, k, [x, y]);
  });
  figure(p, s, 7, t.figure, 640, 668);
  return p;
}
