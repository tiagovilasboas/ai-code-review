// Tone made only of lines: hatching, cross-hatching and stipple, as small SVG <pattern> tiles.
// Patterns use the user space of the element that paints with them, so hatching drawn inside a
// sheared group follows the perspective of that surface.

import { num } from './geometry.js';
import { seeded } from './random.js';

// Tone ramp, light to dark. gap = line spacing (px), w = line weight (px), cross = second direction.
export const TONES = {
  1: { gap: 6, w: 0.55 },
  2: { gap: 4.6, w: 0.7 },
  3: { gap: 3.7, w: 0.85 },
  4: { gap: 4.2, w: 0.75, cross: true },
  5: { gap: 3.2, w: 0.95, cross: true },
};

// Lines run along `angle` (degrees, 0 = horizontal, y down).
export function hatchPattern(id, tone, angle, color, weight) {
  const { gap, w, cross } = TONES[tone];
  const sw = num(w * weight * 100) / 100;
  const h = `<path d="M0 ${num(gap / 2)}H${num(gap)}" stroke="${color}" stroke-width="${sw}"/>`;
  const v = cross ? `<path d="M${num(gap / 2)} 0V${num(gap)}" stroke="${color}" stroke-width="${sw}"/>` : '';
  return `<pattern id="${id}" patternUnits="userSpaceOnUse" width="${num(gap)}" height="${num(gap)}" patternTransform="rotate(${angle})">${h}${v}</pattern>`;
}

// Stipple: seeded dots in a tile; zero-length round-capped segments keep it to one path.
export function stipplePattern(id, density, color, weight, seed) {
  const size = 48;
  const rand = seeded(seed);
  const dots = [];
  for (let i = 0; i < density; i++) dots.push(`M${num(rand() * size)} ${num(rand() * size)}h.01`);
  const sw = num(1.1 * weight * 100) / 100;
  return `<pattern id="${id}" patternUnits="userSpaceOnUse" width="${size}" height="${size}"><path d="${dots.join('')}" stroke="${color}" stroke-width="${sw}" stroke-linecap="round"/></pattern>`;
}
