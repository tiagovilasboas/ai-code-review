// Plate VI: any harness. An orrery: the process, defined once, at the centre; the six harness
// adapters on the ring. claude-code runs the full process (a triple line), the others in sequence (dots).

import { Plate } from '../pen/plate.js';
import { num } from '../pen/geometry.js';
import { rule } from '../pen/ornaments.js';
import { medallion, orb } from '../figures/solids.js';
import { sectionRef } from '../source.js';
import { sheet, figure, meta } from './sheet.js';

const CX = 410;
const CY = 372;
const RX = 286;
const RY = 182;
const HUB = 94;
const NODE = 24;
const KEY_X = 800;
const KEY_W = 400;
// Folder names under harness/; the first runs the full process.
const HARNESSES = ['claude-code', 'kiro', 'codex', 'cursor', 'copilot', 'gemini'];
const DOTS = ' stroke-dasharray="0.1 6" stroke-linecap="round"';

const triple = (p, [x1, y1], [x2, y2], width = 1.3) => {
  const len = Math.hypot(x2 - x1, y2 - y1);
  const [nx, ny] = [-(y2 - y1) / len, (x2 - x1) / len];
  [-5, 0, 5].forEach((o) => p.swell([[x1 + nx * o, y1 + ny * o], [x2 + nx * o, y2 + ny * o]], width, 'ink', 0.5));
};

function orbit(p) {
  for (const [grow, width] of [[0, 0.9], [6, 0.4]]) {
    p.add(`<ellipse cx="${CX}" cy="${CY}" rx="${RX + grow}" ry="${RY + grow}" fill="none" stroke="${p.color()}" stroke-width="${p.w(width)}"/>`);
  }
  HARNESSES.forEach((name, i) => {
    const a = -Math.PI / 2 + (i * Math.PI) / 3;
    const [x, y] = [CX + RX * Math.cos(a), CY + RY * Math.sin(a)];
    const len = Math.hypot(x - CX, y - CY);
    const [ux, uy] = [(x - CX) / len, (y - CY) / len];
    const from = [CX + ux * (HUB + 8), CY + uy * (HUB + 8)];
    const to = [x - ux * (NODE + 8), y - uy * (NODE + 8)];
    if (i === 0) triple(p, from, to);
    else p.stroke(`M${num(from[0])} ${num(from[1])}L${num(to[0])} ${num(to[1])}`, 1.6, 'ink', DOTS);
    orb(p, x, y, NODE);
    const c = Math.cos(a);
    if (Math.abs(c) < 0.2) p.text(name, x, y + (Math.sin(a) < 0 ? -NODE - 14 : NODE + 28), { size: 19, anchor: 'middle' });
    else p.text(name, x + Math.sign(c) * (NODE + 12), y + 6, { size: 19, anchor: c > 0 ? 'start' : 'end' });
  });
}

function hub(p, t) {
  medallion(p, CX, CY, HUB, { rim: 0.12, shadow: false });
  p.text(t.centre, CX, CY - 12, { size: p.fit(t.centre, 140, { size: 24 }), anchor: 'middle' });
  p.text(t.centreSub, CX, CY + 12, { size: p.fit(t.centreSub, 140, { size: 17, font: 'italic' }), font: 'italic', anchor: 'middle' });
  p.wrap(t.files, 130, { size: 14 }).forEach((line, i) => p.text(line, CX, CY + 38 + i * 17, { size: 14, anchor: 'middle' }));
}

// An engraved plaque for the install command, with a hatched drop shadow.
function plaque(p, str, x, y) {
  const size = p.fit(str, KEY_W - 36, { size: 16 });
  const h = 44;
  const box = (dx, dy, w, hh) => `M${x + dx} ${y + dy}h${w}v${hh}h${-w}Z`;
  p.fill(box(6, 6, KEY_W, h), p.tone(4, 45));
  p.fill(box(0, 0, KEY_W, h), p.color('paper'));
  p.stroke(box(0, 0, KEY_W, h), 1.3);
  p.stroke(box(4, 4, KEY_W - 8, h - 8), 0.5);
  p.text(str, x + KEY_W / 2, y + 28, { size, anchor: 'middle' });
}

function key(p, t) {
  let y = 196;
  triple(p, [KEY_X, y - 6], [KEY_X + 50, y - 6], 1.1);
  p.caps(t.full, KEY_X + 66, y, { size: 16 });
  y = p.para(t.keyFull, KEY_X + 66, y + 24, { size: 16, font: 'italic', width: KEY_W - 66 }) + 50;
  p.stroke(`M${KEY_X} ${y - 6}H${KEY_X + 50}`, 1.6, 'ink', DOTS);
  p.caps(t.sequential, KEY_X + 66, y, { size: 16 });
  y = p.para(t.keySeq, KEY_X + 66, y + 24, { size: 16, font: 'italic', width: KEY_W - 66 }) + 40;
  rule(p, KEY_X, KEY_X + KEY_W, y, 1.8);
  y = p.para(t.same, KEY_X, y + 38, { size: 18, width: KEY_W, lead: 25 }) + 30;
  plaque(p, t.install, KEY_X, y);
}

export function harness({ theme, fonts, s }) {
  const t = s.harness;
  const p = new Plate({ id: 'hm', width: 1280, height: 800, theme, fonts, seed: 1606, lang: s.lang, ...meta(s, t.svgTitle, t.desc) });
  sheet(p, s, { n: 6, caption: t.caption, refs: [sectionRef(2)] });
  orbit(p);
  hub(p, t);
  key(p, t);
  figure(p, s, 6, t.figure, 640, 648);
  return p;
}
