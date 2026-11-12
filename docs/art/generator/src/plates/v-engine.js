// Vignette "Como funciona": a diff sheet feeds three meshing gears (parse, regras, relatorio)
// and a report sheet comes out, the first (blocking) row in the accent ink. Borderless, for the README.

import { Plate } from '../pen/plate.js';
import { arrow } from '../pen/ornaments.js';
import { polyD } from '../pen/geometry.js';
import { PHONE_W, PHONE_TEXT, phoneLegend, closePhone } from './phone.js';

const W = 1280;
const H = 396;
const GY = 170;
const GX = [418, 560, 702];
const TEETH = 16;
const RO = 78;
const RI = 65;

const rect = (x, y, w, h) => `M${x} ${y}h${w}v${h}h${-w}Z`;

// A paper sheet with a hatched drop shadow and an inner rule.
function sheet(p, x, y, w, h) {
  p.fill(rect(x + 6, y + 6, w, h), p.tone(4, 45));
  p.fill(rect(x, y, w, h), p.color('paper'));
  p.stroke(rect(x, y, w, h), 1.4);
  p.stroke(rect(x + 4, y + 4, w - 8, h - 8), 0.5);
}

function gearD(cx, cy, phase) {
  const step = (Math.PI * 2) / TEETH;
  const at = (a, r) => [cx + r * Math.cos(a), cy + r * Math.sin(a)];
  const pts = [];
  for (let k = 0; k < TEETH; k++) {
    const a = phase + k * step;
    pts.push(at(a - step * 0.42, RI), at(a - step * 0.2, RO), at(a + step * 0.2, RO), at(a + step * 0.42, RI));
  }
  return polyD(pts);
}

function gear(p, i) {
  const cx = GX[i];
  const d = gearD(cx, GY, i === 1 ? Math.PI / TEETH : 0);
  p.fill(d, p.color('paper'));
  p.fill(d, p.tone(i === 1 ? 3 : 2, i === 1 ? 135 : 45));
  p.stroke(d, 1.5, 'ink', ' stroke-linejoin="round"');
  const ring = (r, w, paint = 'none') => p.add(`<circle cx="${cx}" cy="${GY}" r="${r}" fill="${paint}" stroke="${p.color()}" stroke-width="${p.w(w)}"/>`);
  ring(50, 0.9, p.color('paper'));
  ring(45, 0.5);
  ring(15, 1.2, p.color('paper'));
  for (let k = 0; k < 4; k++) {
    const a = (k * Math.PI) / 2 + Math.PI / 4 + i;
    p.swell([[cx + 15 * Math.cos(a), GY + 15 * Math.sin(a)], [cx + 45 * Math.cos(a), GY + 45 * Math.sin(a)]], 2.2, 'ink', 0.3);
  }
}

function diffSheet(p, t) {
  const x = 50;
  const y = 50;
  sheet(p, x, y, 220, 240);
  p.text(t.hunk, x + 16, y + 28, { size: 15, font: 'italic' });
  const lens = [150, 110, 168, 96, 130, 150, 84];
  lens.forEach((len, k) => {
    const ly = y + 58 + k * 25;
    const sign = k === 2 || k === 3 ? '-' : k === 4 || k === 5 ? '+' : '';
    p.text(sign, x + 16, ly + 4, { size: 15 });
    p.stroke(`M${x + 34} ${ly}h${len}`, sign ? 1.6 : 0.9);
  });
  p.caps(t.diff, x + 110, y + 280, { size: 16, anchor: 'middle' });
}

function reportSheet(p, t) {
  const x = 915;
  const y = 50;
  const w = 320;
  sheet(p, x, y, w, 240);
  p.text(t.reportTitle, x + 20, y + 30, { size: 17, font: 'italic' });
  p.stroke(`M${x + 20} ${y + 42}H${x + w - 20}`, 0.8);
  t.rows.forEach(([sev, where, why], k) => {
    const ry = y + 80 + k * 42;
    const ink = k === 0 ? 'accent' : 'ink';
    p.text(sev, x + 20, ry, { size: 13, font: 'bold', ink, track: 0.04 });
    p.text(where, x + 100, ry, { size: 15, ink });
    p.text(why, x + 100, ry + 19, { size: 14, font: 'italic', ink });
    if (k === 0) {
      p.stroke(`M${x + 14} ${ry - 22}h${w - 28}v${46}h${-(w - 28)}Z`, 1.3, 'accent');
    }
  });
  p.caps(t.report, x + w / 2, y + 280, { size: 16, anchor: 'middle' });
}

function stationNames(p, t) {
  t.stations.forEach((name, i) => p.text(name, GX[i], GY + 118, { size: 22, anchor: 'middle' }));
}

export function vEngine({ theme, fonts, s }) {
  const t = s.vEngine;
  const p = new Plate({ id: 've', width: W, height: H, theme, fonts, seed: 2701, lang: s.lang, title: t.svgTitle, desc: t.desc });
  diffSheet(p, t);
  reportSheet(p, t);
  [0, 1, 2].forEach((i) => gear(p, i));
  stationNames(p, t);
  arrow(p, [296, GY], [GX[0] - RO - 14, GY], { width: 1.4 });
  arrow(p, [GX[2] + RO + 14, GY], [892, GY], { width: 1.4 });
  p.text(t.legend, W / 2, H - 18, { size: 18, font: 'italic', anchor: 'middle' });
  return p;
}

// Phone cut: the same machine read top to bottom. A short diff sheet, the three gears drawn larger,
// and a report with one finding of each severity.
const ZOOM = 1.5;
const SHOWN = [0, 1, 3];
const DIFF = { x: 200, y: 20, w: 320, h: 166 };
const GEARS_TOP = 262;
const REPORT = { x: 24, w: 666, row: 100 };

function phoneDiff(p, t) {
  const { x, y, w, h } = DIFF;
  sheet(p, x, y, w, h);
  p.text(t.hunk, x + 22, y + 48, { size: p.fit(t.hunk, w - 44, { size: PHONE_TEXT, font: 'italic' }), font: 'italic' });
  [['-', 220], ['+', 170], ['+', 250]].forEach(([sign, len], k) => {
    const ly = y + 84 + k * 30;
    p.text(sign, x + 22, ly + 10, { size: PHONE_TEXT });
    p.stroke(`M${x + 50} ${ly}h${len}`, 2.4);
  });
}

function phoneReport(p, t, y) {
  const { x, w, row } = REPORT;
  const h = 84 + SHOWN.length * row + 14;
  sheet(p, x, y, w, h);
  p.text(t.reportTitle, x + 26, y + 50, { size: PHONE_TEXT, font: 'italic' });
  p.stroke(`M${x + 26} ${y + 66}H${x + w - 26}`, 1.2);
  SHOWN.forEach((index, k) => {
    const [sev, where, why] = t.rows[index];
    const ry = y + 84 + k * row;
    const ink = k === 0 ? 'accent' : 'ink';
    p.text(sev, x + 26, ry + 40, { size: PHONE_TEXT, font: 'bold', ink, track: 0.04 });
    p.text(where, x + 222, ry + 40, { size: p.fit(where, w - 248, { size: PHONE_TEXT }), ink });
    p.text(why, x + 222, ry + 80, { size: p.fit(why, w - 248, { size: PHONE_TEXT, font: 'italic' }), font: 'italic', ink });
    if (k === 0) p.stroke(`M${x + 12} ${ry + 4}h${w - 24}v${row - 8}h${-(w - 24)}Z`, 2, 'accent');
  });
  return y + h;
}

export function vEnginePhone({ theme, fonts, s }) {
  const t = s.vEngine;
  const p = new Plate({ id: 'vem', width: PHONE_W, height: 0, theme, fonts, seed: 2701, lang: s.lang, title: t.svgTitle, desc: t.desc });
  const mid = PHONE_W / 2;
  phoneDiff(p, t);
  arrow(p, [mid, DIFF.y + DIFF.h + 16], [mid, GEARS_TOP - 14], { width: 2 });
  p.add(`<g transform="translate(${mid - GX[1] * ZOOM} ${GEARS_TOP - (GY - RO) * ZOOM}) scale(${ZOOM})">`);
  [0, 1, 2].forEach((i) => gear(p, i));
  stationNames(p, t);
  p.add('</g>');
  const names = GEARS_TOP + (118 + RO) * ZOOM;
  arrow(p, [mid, names + 24], [mid, names + 70], { width: 2 });
  const bottom = phoneReport(p, t, names + 86);
  return closePhone(p, phoneLegend(p, t.legend, bottom + 66));
}
