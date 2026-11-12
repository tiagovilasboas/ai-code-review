// Vignette: an engraved tablet carrying one command and the lines it prints, ending in a verdict.
// Borderless (BARE themes): no frame, no caption, no figure number.

import { Plate } from '../pen/plate.js';
import { arrow } from '../pen/ornaments.js';
import { block } from '../figures/solids.js';
import { PHONE_W, PHONE_TEXT, phoneLegend, closePhone } from './phone.js';

const TABLET = { x: 90, y: 64, w: 800, h: 250, d: 26 };
const PAD = 40;
const CMD_Y = TABLET.y + 58;
const LINE_Y = (i) => TABLET.y + 118 + i * 30;
const VERDICT_Y = TABLET.y + 226;

// The tablet: a paper face with a hatched top and edge, and a cast shadow on the ground.
function tablet(p) {
  const { x, y, w, h, d } = TABLET;
  p.swell([[x + 30, y + h + 16], [x + w + d - 10, y + h + 16]], 1.6, 'ink', 0.05);
  p.swell([[x + 70, y + h + 22], [x + w - 40, y + h + 22]], 0.8, 'ink', 0.05);
  block(p, { x, y, w, h, d, top: [2, 30], side: [4, 90] });
}

// The command line: a prompt, then the command in one run so it reads as copy-and-paste.
function command(p, t) {
  const x = TABLET.x + PAD;
  p.text('$', x, CMD_Y, { size: 24 });
  const size = p.fit(t.command, TABLET.w - 2 * PAD - 30, { size: 24, font: 'bold' });
  p.text(t.command, x + 30, CMD_Y, { size, font: 'bold' });
  p.swell([[x, CMD_Y + 24], [TABLET.x + TABLET.w - PAD, CMD_Y + 24]], 1.2, 'ink', 0.05);
}

// What the command prints: plain lines, then the verdict line in the accent ink.
function output(p, t) {
  const x = TABLET.x + PAD;
  t.output.forEach((line, i) => p.text(line, x, LINE_Y(i), { size: 19, font: 'italic' }));
  p.swell([[x, VERDICT_Y - 26], [TABLET.x + TABLET.w - PAD, VERDICT_Y - 26]], 0.9, 'ink', 0.05);
  p.caps(t.verdictWord, x, VERDICT_Y, { size: 22, ink: 'accent', track: 0.12 });
  p.text(t.verdict, x + p.measure(t.verdictWord.toUpperCase(), { size: 22, track: 0.12 }) + 28, VERDICT_Y, { size: 19, font: 'italic', ink: 'accent' });
}

// Two margin notes, each with an arrow back to the line it explains.
function notes(p, t) {
  const x = TABLET.x + TABLET.w + TABLET.d + 70;
  [[t.noteCommand, CMD_Y - 6, 'ink'], [t.noteVerdict, VERDICT_Y - 6, 'accent']].forEach(([note, y, ink]) => {
    arrow(p, [x - 10, y], [TABLET.x + TABLET.w + TABLET.d + 12, y], { ink });
    p.para(note, x + 6, y - 2, { size: 18, width: 250, lead: 22, font: 'italic', ink });
  });
}

export function vQuickstart({ theme, fonts, s }) {
  const t = s.vQuickstart;
  const p = new Plate({ id: 'vq', width: 1280, height: 400, theme, fonts, seed: 1702, lang: s.lang, title: t.svgTitle, desc: t.desc });
  tablet(p);
  command(p, t);
  output(p, t);
  notes(p, t);
  p.text(t.legend, 640, 376, { size: 19, font: 'italic', anchor: 'middle' });
  return p;
}

// Phone cut: the tablet fills the width, the command breaks before the repository name, and the
// margin notes are left out.
const PHONE = { x: 24, y: 56, w: 644, h: 500, d: 26 };

export function vQuickstartPhone({ theme, fonts, s }) {
  const t = s.vQuickstart;
  const p = new Plate({ id: 'vqm', width: PHONE_W, height: 0, theme, fonts, seed: 1702, lang: s.lang, title: t.svgTitle, desc: t.desc });
  const { y, w, h, d } = PHONE;
  const x = PHONE.x + 30;
  const right = PHONE.x + w - 30;
  p.swell([[PHONE.x + 24, y + h + 16], [PHONE.x + w + d - 8, y + h + 16]], 2.2, 'ink', 0.05);
  block(p, { ...PHONE, top: [2, 30], side: [4, 90] });
  const cut = t.command.indexOf('ai-code-review');
  const lines = cut > 0 ? [t.command.slice(0, cut), t.command.slice(cut)] : [t.command];
  const size = Math.min(...lines.map((l) => p.fit(l, right - x - 34, { size: PHONE_TEXT, font: 'bold' })));
  p.text('$', x, y + 66, { size: PHONE_TEXT });
  lines.forEach((l, i) => p.text(l, x + 34, y + 66 + i * 42, { size, font: 'bold' }));
  p.swell([[x, y + 136], [right, y + 136]], 2, 'ink', 0.05);
  t.output.forEach((l, i) => p.text(l, x, y + 196 + i * 50, { size: p.fit(l, right - x, { size: PHONE_TEXT, font: 'italic' }), font: 'italic' }));
  p.swell([[x, y + 330], [right, y + 330]], 1.6, 'ink', 0.05);
  p.caps(t.verdictWord, x, y + 398, { size: 48, ink: 'accent', track: 0.12 });
  p.text(t.verdict, x, y + 456, { size: PHONE_TEXT, font: 'italic', ink: 'accent' });
  return closePhone(p, phoneLegend(p, t.legend, y + h + 84));
}
