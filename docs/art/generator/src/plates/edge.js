// Plate VIII: the Cloudflare edge. A gate with its portcullis down stops the diff; on the left, the
// ledger of what the CLI catches, with severity and action; on the right, what stays with the reviewer.

import { Plate } from '../pen/plate.js';
import { num, polyD } from '../pen/geometry.js';
import { arrow, burst } from '../pen/ornaments.js';
import { block } from '../figures/solids.js';
import { sectionRef, CLOUDFLARE_REF } from '../source.js';
import { sheet, figure, meta } from './sheet.js';

const TOP = 128;
const LEFT = 64;
const LEFT_W = 474;
const RIGHT = 776;
const RIGHT_W = 440;
const GX = 640;
const LINTEL_Y = 214;
const GROUND = 470;
const DOTTED = ' stroke-dasharray="0.1 4" stroke-linecap="round"';

function heading(p, head, sub, x, width) {
  p.caps(head, x, TOP, { size: 17, track: 0.1 });
  p.text(sub, x, TOP + 24, { size: p.fit(sub, width, { size: 15, font: 'italic' }), font: 'italic' });
}

function ledger(p, t) {
  heading(p, t.cliHead, t.cliSub, LEFT, LEFT_W);
  let y = TOP + 68;
  for (const [desc, severity, action] of t.rows) {
    const ink = action === 'BLOCK' ? 'accent' : 'ink';
    const last = p.para(desc, LEFT, y, { size: 17, width: LEFT_W - 140, lead: 21 });
    p.caps(`${severity} · ${action}`, LEFT + LEFT_W, y, { size: 13, anchor: 'end', ink, track: 0.06 });
    p.stroke(`M${LEFT} ${last + 12}H${LEFT + LEFT_W}`, 0.9, 'ink', DOTTED);
    y = last + 38;
  }
}

function reviewerList(p, t) {
  heading(p, t.revHead, t.revSub, RIGHT, RIGHT_W);
  let y = TOP + 68;
  for (const item of t.items) {
    p.fill(polyD([[RIGHT, y - 6], [RIGHT + 5, y - 11], [RIGHT + 10, y - 6], [RIGHT + 5, y - 1]]), p.color());
    y = p.para(item, RIGHT + 22, y, { size: 17, width: RIGHT_W - 22, lead: 21 }) + 34;
  }
  y += 6;
  p.stroke(`M${RIGHT} ${y}H${RIGHT + RIGHT_W}`, 0.9, 'ink', DOTTED);
  y = p.para(t.note, RIGHT, y + 38, { size: 17, font: 'italic', width: RIGHT_W, lead: 22 });
  p.para(t.ids, RIGHT, y + 34, { size: 15, width: RIGHT_W, lead: 21, track: 0.02 });
}

// Front view: two piers, a lintel, the portcullis lowered to the ground, and the diff stopped before it.
function gate(p, t) {
  const left = GX - 74;
  const right = GX + 48;
  p.fill(`M${left + 28} ${LINTEL_Y + 30}H${right}V${GROUND}H${left + 28}Z`, p.tone(2, 0));
  for (let x = left + 41; x <= right - 8; x += 19) p.swell([[x, LINTEL_Y + 30], [x, GROUND - 14]], 3.2, 'ink', 0.7);
  for (let y = LINTEL_Y + 70; y < GROUND - 20; y += 50) p.swell([[left + 28, y], [right, y]], 2.4, 'ink', 0.7);
  for (let x = left + 41; x <= right - 8; x += 19) p.fill(polyD([[x - 3, GROUND - 14], [x + 3, GROUND - 14], [x, GROUND - 4]]), p.color());
  block(p, { x: left, y: LINTEL_Y + 30, w: 26, h: GROUND - LINTEL_Y - 30, d: 12, front: [2, 90] });
  block(p, { x: right, y: LINTEL_Y + 30, w: 26, h: GROUND - LINTEL_Y - 30, d: 12, front: [2, 90] });
  block(p, { x: left - 12, y: LINTEL_Y, w: right - left + 50, h: 30, d: 12 });
  p.caps('BLOCK', GX, LINTEL_Y + 22, { size: 17, anchor: 'middle', ink: 'accent', track: 0.2 });
  p.swell([[GX - 104, GROUND], [GX + 116, GROUND]], 2, 'ink', 0.1);
  p.add(`<ellipse cx="${GX + 6}" cy="${GROUND + 9}" rx="100" ry="6" fill="${p.tone(3, 0)}"/>`);
  // The diff: a sheet with a turned corner, sent up into the gate and stopped at the portcullis.
  const [sx, sy, sw, sh] = [GX - 24, GROUND + 52, 48, 60];
  p.fill(polyD([[sx, sy], [sx + sw - 12, sy], [sx + sw, sy + 12], [sx + sw, sy + sh], [sx, sy + sh]]), p.color('paper'));
  p.fill(`M${sx + 7} ${sy + 18}h${sw - 14}v${sh - 26}h${-(sw - 14)}Z`, p.tone(2, 0));
  p.stroke(polyD([[sx, sy], [sx + sw - 12, sy], [sx + sw, sy + 12], [sx + sw, sy + sh], [sx, sy + sh]]), 1.2);
  p.stroke(`M${sx + sw - 12} ${sy}V${sy + 12}H${sx + sw}`, 0.8);
  p.text(t.gate, sx + sw + 14, sy + sh / 2 + 6, { size: 17, font: 'italic' });
  arrow(p, [GX, sy - 6], [GX, GROUND + 8], { ink: 'accent', width: 1.6 });
  burst(p, GX, GROUND + 4, { ink: 'accent', inner: 8, outer: 15 });
}

export function edge({ theme, fonts, s }) {
  const t = s.edge;
  const p = new Plate({ id: 'ce', width: 1280, height: 800, theme, fonts, seed: 1808, lang: s.lang, ...meta(s, t.svgTitle, t.desc) });
  sheet(p, s, { n: 8, caption: t.caption, refs: [sectionRef(3), CLOUDFLARE_REF] });
  ledger(p, t);
  reviewerList(p, t);
  gate(p, t);
  figure(p, s, 8, t.figure, GX, 624, { size: 17 });
  return p;
}
