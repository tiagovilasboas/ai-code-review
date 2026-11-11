// Engraved solids for the diagrams: a block in oblique projection and a medallion.
// Light comes from the upper left: top faces are lightest, right sides darkest.

import { num, polyD } from '../pen/geometry.js';

// Box with its front face at (x, y, w, h) and depth d receding up and to the right.
// front: [tone, angle] or null for a paper face that carries text.
export function block(p, { x, y, w, h, d = 16, front = null, top = [1, 30], side = [3, 90], ink = 'ink' }) {
  const dx = d;
  const dy = -d * 0.6;
  const F = [[x, y], [x + w, y], [x + w, y + h], [x, y + h]];
  const T = [[x, y], [x + dx, y + dy], [x + w + dx, y + dy], [x + w, y]];
  const S = [[x + w, y], [x + w + dx, y + dy], [x + w + dx, y + h + dy], [x + w, y + h]];
  p.fill(polyD([F[0], T[1], T[2], S[2], F[2], F[3]]), p.color('paper'));
  if (front) p.fill(polyD(F), p.tone(front[0], front[1], ink));
  p.fill(polyD(T), p.tone(top[0], top[1], ink));
  p.fill(polyD(S), p.tone(side[0], side[1], ink));
  p.swell([F[0], F[1]], 1, ink);
  p.swell([F[3], F[2]], 1.8, ink);
  p.swell([F[0], F[3]], 1.2, ink);
  p.swell([F[1], F[2]], 1.4, ink);
  p.swell([T[1], T[2]], 0.9, ink);
  p.swell([F[0], T[1]], 0.9, ink);
  p.swell([T[2], S[2]], 1.2, ink);
  p.swell([F[1], T[2]], 0.9, ink);
  p.swell([F[2], S[2]], 1.2, ink);
  return { x, y, w, h, d };
}

// A coin-like medallion: a hatched rim, a paper field for a numeral or word, and a cast shadow.
export function medallion(p, cx, cy, r, { ink = 'ink', rim = 0.22, shadow = true } = {}) {
  const c = (rr, attrs, dx = 0, dy = 0) => `<circle cx="${num(cx + dx)}" cy="${num(cy + dy)}" r="${num(rr)}" ${attrs}/>`;
  if (shadow) p.add(c(r, `fill="${p.tone(4, 45)}"`, r * 0.16, r * 0.2));
  p.add(c(r, `fill="${p.color('paper')}"`));
  const inner = r * (1 - rim);
  p.add(`<path d="M${num(cx - r)} ${num(cy)}a${num(r)} ${num(r)} 0 1 0 ${num(2 * r)} 0a${num(r)} ${num(r)} 0 1 0 ${num(-2 * r)} 0Z`
    + `M${num(cx - inner)} ${num(cy)}a${num(inner)} ${num(inner)} 0 1 0 ${num(2 * inner)} 0a${num(inner)} ${num(inner)} 0 1 0 ${num(-2 * inner)} 0Z" fill-rule="evenodd" fill="${p.tone(3, 45, ink)}"/>`);
  p.add(c(r, `fill="none" stroke="${p.color(ink)}" stroke-width="${p.w(1.6)}"`));
  p.add(c(inner, `fill="none" stroke="${p.color(ink)}" stroke-width="${p.w(0.8)}"`));
}

// A shaded sphere, lit from the upper left: a light tone overall and a dark crescent away from the light.
export function orb(p, cx, cy, r, { ink = 'ink' } = {}) {
  const disc = (x, y, rr) => `M${num(x - rr)} ${num(y)}a${num(rr)} ${num(rr)} 0 1 0 ${num(2 * rr)} 0a${num(rr)} ${num(rr)} 0 1 0 ${num(-2 * rr)} 0Z`;
  p.fill(disc(cx, cy, r), p.color('paper'));
  p.fill(disc(cx, cy, r), p.tone(1, 30, ink));
  const clip = p.clip(`<circle cx="${num(cx)}" cy="${num(cy)}" r="${num(r)}"/>`);
  p.fill(disc(cx, cy, r) + disc(cx - r * 0.32, cy - r * 0.32, r * 0.96), p.tone(4, 30, ink), ` fill-rule="evenodd" clip-path="${clip}"`);
  p.add(`<circle cx="${num(cx)}" cy="${num(cy)}" r="${num(r)}" fill="none" stroke="${p.color(ink)}" stroke-width="${p.w(1.3)}"/>`);
}
