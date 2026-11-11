// The source cartouche at the foot of every plate: a framed credit in the engraver's
// "after …" manner. Lines: optional references, the credit, then the address of the post.

import { num, polyD } from './geometry.js';

const SIZE = 18;
const LEAD = 24;

// A credit line is roman, with the quoted title in italic.
function segments(str) {
  return str.split(/(“[^”]*”)/).filter(Boolean).map((text) => ({ text, font: text.startsWith('“') ? 'italic' : 'roman' }));
}

function lineWidth(p, str) {
  return segments(str).reduce((sum, g) => sum + p.measure(g.text, { font: g.font, size: SIZE }), 0);
}

function drawLine(p, str, cx, y) {
  let x = cx - lineWidth(p, str) / 2;
  for (const g of segments(str)) x += p.text(g.text, x, y, { font: g.font, size: SIZE });
}

// Rectangle whose four corners are cut by concave quarter circles.
function frameD(x, y, w, h, r) {
  const a = `A${r} ${r} 0 0 0`;
  return `M${num(x + r)} ${num(y)}H${num(x + w - r)}${a} ${num(x + w)} ${num(y + r)}V${num(y + h - r)}`
    + `${a} ${num(x + w - r)} ${num(y + h)}H${num(x + r)}${a} ${num(x)} ${num(y + h - r)}V${num(y + r)}${a} ${num(x + r)} ${num(y)}Z`;
}

function flourish(p, x1, x2, y) {
  p.swell([[x1, y], [x2, y]], 1.6, 'ink', 0.1);
  const tip = x1 < x2 ? x1 : x2;
  p.fill(polyD([[tip - 6, y], [tip, y - 3.5], [tip + 6, y], [tip, y + 3.5]]), p.color());
}

// Draws the cartouche centred on the plate with its bottom edge at `bottom`; returns its top edge.
export function cartouche(p, lines, bottom) {
  const width = Math.max(...lines.map((l) => lineWidth(p, l))) + 72;
  const height = LEAD * lines.length + 22;
  const x = (p.width - width) / 2;
  const y = bottom - height;
  p.fill(frameD(x, y, width, height, 10), p.color('paper'));
  p.stroke(frameD(x, y, width, height, 10), 1.4);
  p.stroke(frameD(x + 4, y + 4, width - 8, height - 8, 7), 0.6);
  flourish(p, x - 52, x - 10, y + height / 2);
  flourish(p, x + width + 10, x + width + 52, y + height / 2);
  lines.forEach((l, i) => drawLine(p, l, p.width / 2, y + 30 + i * LEAD));
  return y;
}

// The credit splits at its last " · ": the citation on one line, the address on the next.
export function creditLines(after) {
  const cut = after.lastIndexOf(' · ');
  return cut < 0 ? [after] : [after.slice(0, cut), after.slice(cut + 3)];
}
