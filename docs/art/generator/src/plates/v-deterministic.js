// Vignette: a stamping press. The same diff, fed in twice, comes out with the same verdict seal.
// Borderless (BARE themes): no frame, no caption, no figure number.

import { Plate } from '../pen/plate.js';
import { polyD } from '../pen/geometry.js';
import { arrow, burst } from '../pen/ornaments.js';
import { block } from '../figures/solids.js';
import { PHONE_W, PHONE_TEXT, phoneLegend, closePhone } from './phone.js';

const ROWS = [126, 262];
const SHEET = { x: 150, w: 104, h: 96 };
const PRESS = { x: 470, w: 130 };
const SEAL = { x: 900, r: 44 };

// A diff sheet: a ruled page with removed (-) and added (+) lines, and a hatched shadow behind it.
function diffSheet(p, cy, sheet = SHEET) {
  const { x, w, h } = sheet;
  const y = cy - h / 2;
  p.fill(polyD([[x + 7, y + 7], [x + w + 7, y + 7], [x + w + 7, y + h + 7], [x + 7, y + h + 7]]), p.tone(4, 45));
  p.fill(polyD([[x, y], [x + w, y], [x + w, y + h], [x, y + h]]), p.color('paper'));
  p.stroke(`M${x} ${y}H${x + w}V${y + h}H${x}Z`, 1.3);
  [['−', 62], ['+', 78], ['+', 50], [' ', 70], ['−', 40]].forEach(([mark, len], i) => {
    const ly = y + 16 + i * 16;
    if (mark.trim()) p.text(mark, x + 11, ly + 4, { size: 13 });
    p.swell([[x + 24, ly], [x + 24 + len, ly]], 1.3, 'ink', 0.2);
  });
}

// The press: a hatched block with a plunger above it and a bearing line below.
function press(p, cy, at = PRESS) {
  const { x, w } = at;
  block(p, { x, y: cy - 24, w, h: 48, d: 16, front: [2, 0] });
  block(p, { x: x + w / 2 - 22, y: cy - 62, w: 44, h: 26, d: 10, front: [1, 90], side: [3, 90] });
  p.swell([[x + w / 2, cy - 78], [x + w / 2, cy - 62]], 3, 'ink', 0.5);
  p.swell([[x - 12, cy + 34], [x + w + 28, cy + 34]], 1.2, 'ink', 0.05);
}

// The verdict seal: a double ring in the accent ink around the word BLOCK, with an impact burst.
function seal(p, cy, word, at = SEAL) {
  const { x, r, size = 19, rise = 6, burst: [bx, by] = [r + 20, 8 - r] } = at;
  const circle = (rr, w) => p.add(`<circle cx="${x}" cy="${cy}" r="${rr}" fill="${p.color('paper')}" stroke="${p.color('accent')}" stroke-width="${p.w(w)}"/>`);
  circle(r, 2);
  p.add(`<circle cx="${x}" cy="${cy}" r="${r - 6}" fill="none" stroke="${p.color('accent')}" stroke-width="${p.w(0.8)}"/>`);
  p.caps(word, x, cy + rise, { size, anchor: 'middle', ink: 'accent', track: 0.12 });
  burst(p, x + bx, cy + by, { inner: 4, outer: 9, rays: 8 });
}

function row(p, cy, label, t) {
  diffSheet(p, cy);
  p.text(label, SHEET.x + SHEET.w / 2, cy + SHEET.h / 2 + 24, { size: 15, font: 'italic', anchor: 'middle' });
  arrow(p, [SHEET.x + SHEET.w + 22, cy], [PRESS.x - 18, cy]);
  press(p, cy);
  arrow(p, [PRESS.x + PRESS.w + 36, cy], [SEAL.x - SEAL.r - 20, cy], { ink: 'accent' });
  seal(p, cy, t.verdict);
}

// A bracket joining the two seals, with the words that say what it proves.
function equality(p, t) {
  const x = SEAL.x + SEAL.r + 52;
  const [a, b] = ROWS;
  p.stroke(`M${x - 10} ${a}H${x}V${b}H${x - 10}`, 1.4);
  p.text('=', x + 20, (a + b) / 2 + 9, { size: 30, anchor: 'middle' });
  p.para(t.same, x + 40, (a + b) / 2 - 8, { size: 18, width: 150, lead: 22, font: 'italic' });
}

export function vDeterministic({ theme, fonts, s }) {
  const t = s.vDeterministic;
  const p = new Plate({ id: 'vd', width: 1280, height: 400, theme, fonts, seed: 1701, lang: s.lang, title: t.svgTitle, desc: t.desc });
  p.caps(t.colInput, SHEET.x + SHEET.w / 2, 24, { size: 15, anchor: 'middle', track: 0.14 });
  p.caps(t.colReviewer, PRESS.x + PRESS.w / 2, 24, { size: 15, anchor: 'middle', track: 0.14 });
  p.caps(t.colVerdict, SEAL.x, 24, { size: 15, anchor: 'middle', track: 0.14 });
  ROWS.forEach((cy, i) => row(p, cy, t.runs[i], t));
  equality(p, t);
  p.text(t.legend, 640, 380, { size: 19, font: 'italic', anchor: 'middle' });
  return p;
}

// Phone cut: the two runs stacked and drawn larger, an equals sign between the two seals, and the
// column headings left out.
const PHONE = { sheet: { x: 6, w: 104, h: 96 }, press: { x: 158, w: 104 }, seal: { x: 374, r: 49, size: 20, rise: 6.5, burst: [46, -48] } };
const ZOOM = 1.6;
const [LEFT, TOP] = [14, 24];
const RUNS = [84, 274];

export function vDeterministicPhone({ theme, fonts, s }) {
  const t = s.vDeterministic;
  const p = new Plate({ id: 'vdm', width: PHONE_W, height: 0, theme, fonts, seed: 1701, lang: s.lang, title: t.svgTitle, desc: t.desc });
  const { sheet, press: pr, seal: sl } = PHONE;
  const px = (v) => LEFT + v * ZOOM;
  const py = (v) => TOP + v * ZOOM;
  p.add(`<g transform="translate(${LEFT} ${TOP}) scale(${ZOOM})">`);
  RUNS.forEach((cy) => {
    diffSheet(p, cy, sheet);
    arrow(p, [sheet.x + sheet.w + 14, cy], [pr.x - 8, cy]);
    press(p, cy, pr);
    arrow(p, [pr.x + pr.w + 26, cy], [sl.x - sl.r - 8, cy], { ink: 'accent' });
    seal(p, cy, t.verdict, sl);
  });
  p.add('</g>');
  RUNS.forEach((cy, i) => p.text(t.runs[i], px(sheet.x + sheet.w / 2), py(cy + sheet.h / 2) + 48, { size: PHONE_TEXT, font: 'italic', anchor: 'middle' }));
  p.text('=', px(sl.x), py((RUNS[0] + RUNS[1]) / 2) + 16, { size: 52, anchor: 'middle' });
  return closePhone(p, phoneLegend(p, t.legend, py(RUNS[1] + sheet.h / 2) + 48 + 76));
}
