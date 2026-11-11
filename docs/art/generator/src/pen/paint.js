// Paint for a plate: line tones and stipple as <pattern>s, clip paths, fills, strokes and
// swelling lines. Mixed into Plate; `this` is the plate.

import { hatchPattern, stipplePattern } from './hatching.js';
import { swellD } from './geometry.js';

export const paint = {
  // Paint server for a line tone (1 light .. 5 dark) running along `angle` degrees.
  tone(level, angle = 0, ink = 'ink') {
    const id = `${this.id}t${level}a${String(angle).replace('-', 'm')}${ink === 'ink' ? '' : 'x'}`;
    this.def(id, hatchPattern(id, level, angle, this.color(ink), this.theme.weight));
    return `url(#${id})`;
  },

  stipple(density, ink = 'ink') {
    const id = `${this.id}s${density}${ink === 'ink' ? '' : 'x'}`;
    this.def(id, stipplePattern(id, density, this.color(ink), this.theme.weight, this.seed + density));
    return `url(#${id})`;
  },

  // clipPath from shape markup, in the user space of whatever references it.
  clip(markup) {
    const id = this.nextId('c');
    this.def(id, `<clipPath id="${id}">${markup}</clipPath>`);
    return `url(#${id})`;
  },

  fill(d, paint, extra = '') {
    return this.add(`<path d="${d}" fill="${paint}"${extra}/>`);
  },

  stroke(d, width, ink = 'ink', extra = '') {
    return this.add(`<path d="${d}" fill="none" stroke="${this.color(ink)}" stroke-width="${this.w(width)}"${extra}/>`);
  },

  swell(points, width, ink = 'ink', taper = 0.25, steps = 10) {
    return this.fill(swellD(points, this.w(width) + 0.2, taper, steps), this.color(ink));
  },
};
