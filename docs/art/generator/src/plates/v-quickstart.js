// Vignette: an engraved tablet carrying one command and the lines it prints, ending in a verdict.
// Borderless (BARE themes): no frame, no caption, no figure number.

import { Plate } from '../pen/plate.js';
import { arrow } from '../pen/ornaments.js';
import { block } from '../figures/solids.js';

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
