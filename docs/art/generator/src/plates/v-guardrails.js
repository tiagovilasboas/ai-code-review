// Vignette for "Guardrails": three engraved columns carry one lintel, one guardrail each.
// Borderless: no sheet, no frame, no caption; labels come from strings.vGuardrails.

import { Plate } from '../pen/plate.js';
import { polyD } from '../pen/geometry.js';
import { block } from '../figures/solids.js';
import { PHONE_W, PHONE_TEXT, closePhone } from './phone.js';

const CX = [270, 640, 1010];
const TOP = 104;
const [Y1, Y2] = [126, 300];
const GROUND = 322;

function capital(p, cx) {
  p.fill(polyD([[cx - 44, 114], [cx + 44, 114], [cx + 56, TOP + 10], [cx - 56, TOP + 10]]), p.color('paper'));
  p.fill(polyD([[cx - 44, 114], [cx + 44, 114], [cx + 40, 126], [cx - 40, 126]]), p.tone(2, 0));
  p.swell([[cx - 62, TOP], [cx + 62, TOP]], 1.6, 'ink', 0.1);
  p.swell([[cx - 62, TOP + 10], [cx + 62, TOP + 10]], 1.4, 'ink', 0.1);
  p.swell([[cx - 62, TOP], [cx - 62, TOP + 10]], 1, 'ink', 0.3);
  p.swell([[cx + 62, TOP], [cx + 62, TOP + 10]], 1, 'ink', 0.3);
  p.stroke(`M${cx - 44} 114Q${cx - 46} 122 ${cx - 40} 126M${cx + 44} 114Q${cx + 46} 122 ${cx + 40} 126`, 1.1);
}

// Shaft with slight swelling (entasis), flutes, and a dark side away from the light.
function shaft(p, cx) {
  const left = (y) => cx - 40 - 1.6 * Math.sin(((y - Y1) / (Y2 - Y1)) * Math.PI) - ((y - Y1) / (Y2 - Y1)) * 2;
  const rows = Array.from({ length: 9 }, (_, i) => Y1 + ((Y2 - Y1) * i) / 8);
  const L = rows.map((y) => [left(y), y]);
  const R = rows.map((y) => [2 * cx - left(y), y]);
  const outline = polyD([...L, ...[...R].reverse()]);
  p.fill(outline, p.color('paper'));
  p.fill(outline, p.tone(1, 90));
  const clip = p.clip(`<path d="${outline}"/>`);
  p.fill(polyD([[cx + 10, Y1], [cx + 60, Y1], [cx + 60, Y2], [cx + 10, Y2]]), p.tone(3, 90), ` clip-path="${clip}"`);
  for (let k = -4; k <= 4; k++) p.swell(rows.map((y) => [cx + (k / 4.5) * (cx - left(y)), y]), 0.8, 'ink', 0.1, 3);
  p.swell(L, 1.3, 'ink', 0.2, 3);
  p.swell(R, 1.6, 'ink', 0.2, 3);
}

function base(p, cx) {
  p.fill(polyD([[cx - 52, Y2], [cx + 52, Y2], [cx + 56, 310], [cx - 56, 310]]), p.tone(2, 0));
  p.swell([[cx - 56, 310], [cx + 56, 310]], 1.3, 'ink', 0.1);
  block(p, { x: cx - 66, y: 310, w: 132, h: 12, d: 12, front: [3, 0] });
}

function pediment(p, x0, x1) {
  const mid = (x0 + x1) / 2;
  const tri = [[x0, 58], [x1, 58], [mid, 12]];
  p.fill(polyD(tri), p.color('paper'));
  p.fill(polyD(tri), p.tone(1, 25));
  p.swell([tri[0], tri[2], tri[1]], 1.6, 'ink', 0.1);
  p.swell([[x0 + 65, 52], [mid, 22], [x1 - 65, 52]], 0.9, 'ink', 0.1);
}

function ground(p, x0, x1) {
  const k = (x1 - x0) / 1140;
  p.swell([[x0, GROUND], [x1, GROUND]], 1.8, 'ink', 0.05);
  [[6, 0.9, 40], [11, 0.7, 120], [16, 0.5, 220]].forEach(([dy, w, inset]) => p.swell([[x0 + inset * k, GROUND + dy], [x1 - inset * k, GROUND + dy]], w, 'ink', 0.05));
}

// The whole temple: columns at `cxs` under a lintel that starts at x and is w wide.
function temple(p, cxs, x, w) {
  ground(p, x - 70, x + w + 70);
  cxs.forEach((cx) => { shaft(p, cx); capital(p, cx); base(p, cx); });
  block(p, { x, y: 70, w, h: 34, d: 20, front: [1, 0] });
  pediment(p, x + 10, x + w + 20);
}

function labels(p, t) {
  t.columns.forEach(([name, rest], i) => {
    p.fill(polyD([[CX[i] - 5, 363], [CX[i], 358], [CX[i] + 5, 363], [CX[i], 368]]), p.color());
    p.caps(name, CX[i], 394, { size: 20, anchor: 'middle', track: 0.1 });
    p.para(rest, CX[i], 420, { size: 17, width: 330, lead: 21, anchor: 'middle', font: 'italic' });
  });
}

export function vGuardrails({ theme, fonts, s }) {
  const t = s.vGuardrails;
  const p = new Plate({ id: 'vg', width: 1280, height: 446, theme, fonts, seed: 2711, lang: s.lang, title: t.svgTitle, desc: t.desc });
  temple(p, CX, 140, 1000);
  labels(p, t);
  return p;
}

// Phone cut: the same temple with the columns closer, numbered, and the three guardrails listed
// under it in the same order.
const PHONE_CX = [130, 360, 590];
const NUMERALS = ['I', 'II', 'III'];

export function vGuardrailsPhone({ theme, fonts, s }) {
  const t = s.vGuardrails;
  const p = new Plate({ id: 'vgm', width: PHONE_W, height: 0, theme, fonts, seed: 2711, lang: s.lang, title: t.svgTitle, desc: t.desc });
  p.add('<g transform="translate(0 8)">');
  temple(p, PHONE_CX, 76, 548);
  p.add('</g>');
  PHONE_CX.forEach((cx, i) => p.caps(NUMERALS[i], cx, GROUND + 62, { size: 34, anchor: 'middle', track: 0.06 }));
  let y = GROUND + 62;
  t.columns.forEach(([name, rest], i) => {
    y += 84;
    p.caps(`${NUMERALS[i]} · ${name}`, PHONE_W / 2, y, { size: 41, anchor: 'middle', track: 0.08 });
    y += 46;
    p.text(rest, PHONE_W / 2, y, { size: PHONE_TEXT, font: 'italic', anchor: 'middle' });
  });
  return closePhone(p, y);
}
