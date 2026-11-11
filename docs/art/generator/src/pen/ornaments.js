// Plate furniture: double-rule border, guilloche corner rosettes, swelled rules,
// engraved arrows and impact marks.

import { num, pts, polyD } from './geometry.js';

// Interlaced guilloche bands around a centre, as one path (drawn once, reused with <use>).
function rosetteD(r) {
  const bands = [
    { base: 0.78, amp: 0.2, lobes: 10, copies: 4 },
    { base: 0.4, amp: 0.15, lobes: 7, copies: 3 },
  ];
  const parts = [];
  for (const { base, amp, lobes, copies } of bands) {
    for (let j = 0; j < copies; j++) {
      const phase = (j * Math.PI * 2) / copies;
      const ring = [];
      const steps = lobes * 12;
      for (let k = 0; k < steps; k++) {
        const t = (k / steps) * Math.PI * 2;
        const rr = r * (base + amp * Math.sin(lobes * t + phase));
        ring.push([rr * Math.cos(t), rr * Math.sin(t)]);
      }
      parts.push(`M${pts(ring)}Z`);
    }
  }
  return parts.join('');
}

export function rosette(plate, cx, cy, r = 18) {
  const id = `${plate.id}ro`;
  plate.def(id, `<path id="${id}" d="${rosetteD(r)}" fill="none" stroke="${plate.color()}" stroke-width="${plate.w(0.4)}"/>`);
  plate.add(`<use href="#${id}" x="${num(cx)}" y="${num(cy)}"/>`);
  plate.add(`<circle cx="${num(cx)}" cy="${num(cy)}" r="${num(r * 0.1)}" fill="${plate.color()}"/>`);
}

// Thin double rule with a rosette in each inner corner.
export function frame(plate, { outer = 12, gap = 6, corner = 22 } = {}) {
  const { width: W, height: H } = plate;
  const rect = (i) => `M${i} ${i}H${W - i}V${H - i}H${i}Z`;
  plate.stroke(rect(outer), 2);
  plate.stroke(rect(outer + gap), 0.7);
  const c = outer + gap + corner;
  for (const [x, y] of [[c, c], [W - c, c], [c, H - c], [W - c, H - c]]) rosette(plate, x, y);
}

// A swelled rule with a small lozenge at its centre.
export function rule(plate, x1, x2, y, width = 2.4) {
  const mid = (x1 + x2) / 2;
  plate.swell([[x1, y], [mid - 12, y]], width, 'ink', 0.05);
  plate.swell([[mid + 12, y], [x2, y]], width, 'ink', 0.05);
  plate.fill(polyD([[mid - 7, y], [mid, y - 4], [mid + 7, y], [mid, y + 4]]), plate.color());
}

// Straight engraved arrow: fine shaft, barbed head.
export function arrow(plate, [x1, y1], [x2, y2], { ink = 'ink', width = 1.4, extra = '' } = {}) {
  const a = Math.atan2(y2 - y1, x2 - x1);
  const at = (d, side) => [x2 - d * Math.cos(a) - side * Math.sin(a), y2 - d * Math.sin(a) + side * Math.cos(a)];
  const shaftEnd = at(9, 0);
  const shaft = `M${num(x1)} ${num(y1)}L${num(shaftEnd[0])} ${num(shaftEnd[1])}`;
  plate.add(`<path d="${shaft}" fill="none" stroke="${plate.color('paper')}" stroke-width="5"${extra}/>`);
  plate.stroke(shaft, width, ink, extra);
  plate.fill(polyD([[x2, y2], at(15, 5.5), at(10, 0), at(15, -5.5)]), plate.color(ink), extra);
}

// Arrow along a polyline (for curves, pass enough points); the head follows the last segment.
export function arrowAlong(plate, points, { ink = 'ink', width = 1.3, extra = '' } = {}) {
  const [x1, y1] = points[points.length - 2];
  const [x2, y2] = points[points.length - 1];
  const a = Math.atan2(y2 - y1, x2 - x1);
  const at = (d, side) => [x2 - d * Math.cos(a) - side * Math.sin(a), y2 - d * Math.sin(a) + side * Math.cos(a)];
  const shaft = `M${pts([...points.slice(0, -1), at(9, 0)])}`;
  plate.stroke(shaft, width, ink, ` stroke-linecap="round" stroke-linejoin="round"${extra}`);
  plate.fill(polyD([[x2, y2], at(14, 5), at(9.5, 0), at(14, -5)]), plate.color(ink));
}

// Radiating strokes marking the point where something was stopped.
export function burst(plate, cx, cy, { ink = 'accent', inner = 7, outer = 15, rays = 10 } = {}) {
  for (let k = 0; k < rays; k++) {
    const t = (k / rays) * Math.PI * 2 + 0.3;
    const r2 = outer + (k % 2 ? 5 : 0);
    plate.swell([[cx + inner * Math.cos(t), cy + inner * Math.sin(t)], [cx + r2 * Math.cos(t), cy + r2 * Math.sin(t)]], 2, ink, 0.1);
  }
}
