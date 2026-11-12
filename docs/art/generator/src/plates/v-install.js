// Vignette "Instale em qualquer harness": one seal labelled reviewers/ presses the same mark
// onto six sheets, one per harness. Borderless, for the README.

import { Plate } from '../pen/plate.js';
import { arrow } from '../pen/ornaments.js';
import { medallion } from '../figures/solids.js';
import { num, polyD } from '../pen/geometry.js';
import { PHONE_W, phoneLegend, closePhone } from './phone.js';

const W = 1280;
const H = 424;
const SX = 640;
const SY = 98;
const SR = 60;
const SHEET_W = 168;
const SHEET_Y = 250;
const SHEET_H = 112;
const COUNT = 6;

const rect = (x, y, w, h) => `M${x} ${y}h${w}v${h}h${-w}Z`;
const sheetX = (i) => 60 + i * ((W - 120 - SHEET_W) / (COUNT - 1));

// The stamp's mark: a ring around a lozenge with a tick, drawn the same at every size.
function mark(p, cx, cy, r) {
  p.add(`<circle cx="${cx}" cy="${num(cy)}" r="${r}" fill="none" stroke="${p.color()}" stroke-width="${p.w(1.3)}"/>`);
  p.add(`<circle cx="${cx}" cy="${num(cy)}" r="${num(r * 0.8)}" fill="none" stroke="${p.color()}" stroke-width="${p.w(0.6)}"/>`);
  p.fill(polyD([[cx - r * 0.5, cy], [cx, cy - r * 0.5], [cx + r * 0.5, cy], [cx, cy + r * 0.5]]), p.tone(4, 45));
  p.swell([[cx - r * 0.22, cy], [cx - r * 0.04, cy + r * 0.2], [cx + r * 0.3, cy - r * 0.2]], 1.6, 'paper', 0.4);
}

function seal(p, t) {
  p.fill(rect(SX - 14, 12, 28, 30), p.tone(3, 90));
  p.stroke(rect(SX - 14, 12, 28, 30), 1.2);
  p.fill(rect(SX - 26, 2, 52, 12), p.tone(2, 0));
  p.stroke(rect(SX - 26, 2, 52, 12), 1.2);
  medallion(p, SX, SY, SR, { rim: 0.2 });
  p.text(t.seal, SX, SY + 6, { size: p.fit(t.seal, SR * 1.3, { size: 22 }), anchor: 'middle' });
}

function sheetAt(p, i, name) {
  const x = sheetX(i);
  p.fill(rect(x + 5, SHEET_Y + 5, SHEET_W, SHEET_H), p.tone(4, 45));
  p.fill(rect(x, SHEET_Y, SHEET_W, SHEET_H), p.color('paper'));
  p.stroke(rect(x, SHEET_Y, SHEET_W, SHEET_H), 1.3);
  p.stroke(rect(x + 4, SHEET_Y + 4, SHEET_W - 8, SHEET_H - 8), 0.5);
  const size = p.fit(name, SHEET_W - 24, { size: 21 });
  p.text(name, x + SHEET_W / 2, SHEET_Y + 32, { size, anchor: 'middle' });
  p.stroke(`M${x + 20} ${SHEET_Y + 44}h${SHEET_W - 40}`, 0.6);
  mark(p, x + SHEET_W / 2, SHEET_Y + 78, 22);
}

// A stem from the seal and a bus rule, with one arrow dropping onto each sheet.
function impressions(p) {
  const bus = 206;
  const first = sheetX(0) + SHEET_W / 2;
  const last = sheetX(COUNT - 1) + SHEET_W / 2;
  p.swell([[SX, SY + SR + 2], [SX, bus]], 2, 'ink', 0.2);
  p.swell([[first, bus], [last, bus]], 2, 'ink', 0.1);
  for (let i = 0; i < COUNT; i++) {
    const tx = sheetX(i) + SHEET_W / 2;
    arrow(p, [tx, bus], [tx, SHEET_Y - 6], { width: 1.3 });
  }
}

export function vInstall({ theme, fonts, s }) {
  const t = s.vInstall;
  const p = new Plate({ id: 'vi', width: W, height: H, theme, fonts, seed: 3113, lang: s.lang, title: t.svgTitle, desc: t.desc });
  t.harnesses.forEach((name, i) => sheetAt(p, i, name));
  impressions(p);
  seal(p, t);
  p.text(t.legend, W / 2, H - 20, { size: 18, font: 'italic', anchor: 'middle' });
  return p;
}

// Phone cut: the seal on top, drawn larger, and the six harnesses as cards in two columns, each
// stamped with the same mark. A stem runs down between the columns with an arrow to every card.
const ZOOM = 1.85;
const CARD = { w: 290, h: 96, cols: [24, 406], row0: 384, pitch: 120 };

function card(p, x, cy, name) {
  const { w, h } = CARD;
  const y = cy - h / 2;
  p.fill(rect(x + 6, y + 6, w, h), p.tone(4, 45));
  p.fill(rect(x, y, w, h), p.color('paper'));
  p.stroke(rect(x, y, w, h), 1.6);
  p.stroke(rect(x + 5, y + 5, w - 10, h - 10), 0.6);
  p.text(name, x + 24, cy + 11, { size: p.fit(name, w - 110, { size: 34 }) });
  mark(p, x + w - 48, cy, 26);
}

export function vInstallPhone({ theme, fonts, s }) {
  const t = s.vInstall;
  const p = new Plate({ id: 'vim', width: PHONE_W, height: 0, theme, fonts, seed: 3113, lang: s.lang, title: t.svgTitle, desc: t.desc });
  const mid = PHONE_W / 2;
  const rows = Math.ceil(t.harnesses.length / 2);
  const rowY = (r) => CARD.row0 + r * CARD.pitch;
  p.swell([[mid, (SY + SR) * ZOOM], [mid, rowY(rows - 1)]], 2.6, 'ink', 0.2);
  t.harnesses.forEach((name, i) => {
    const [col, cy] = [i % 2, rowY(Math.floor(i / 2))];
    card(p, CARD.cols[col], cy, name);
    arrow(p, [mid, cy], [col ? CARD.cols[1] - 6 : CARD.cols[0] + CARD.w + 12, cy], { width: 1.8 });
  });
  p.add(`<g transform="translate(${num(mid - SX * ZOOM)} 0) scale(${ZOOM})">`);
  seal(p, t);
  p.add('</g>');
  return closePhone(p, phoneLegend(p, t.legend, rowY(rows - 1) + CARD.h / 2 + 66));
}
