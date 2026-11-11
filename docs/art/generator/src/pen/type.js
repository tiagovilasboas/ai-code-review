// Typesetting for a plate: outlined EB Garamond text, measuring, fitting, wrapping, paragraphs
// and faux small capitals. Mixed into Plate; `this` is the plate.

import { num } from './geometry.js';

export const type = {
  // Outlined text. font: roman | bold | italic. track in em. anchor: start | middle | end.
  text(str, x, y, o = {}) {
    const face = this.fonts[o.font ?? 'roman'];
    const size = o.size ?? 20;
    const track = o.track ?? 0;
    const s = size / face.upm;
    const { glyphs, advance } = face.layout(str, track);
    const width = advance * s;
    const dx = o.anchor === 'middle' ? -width / 2 : o.anchor === 'end' ? -width : 0;
    const uses = glyphs
      .map(({ glyph, x: gx }) => {
        const id = this.book.idFor(face, glyph);
        return id ? `<use href="#${id}"${gx ? ` x="${Math.round(gx)}"` : ''}/>` : '';
      })
      .join('');
    const place = `translate(${num(x + dx)} ${num(y)}) scale(${+s.toFixed(5)})`;
    if (o.halo) {
      const paper = this.color('paper');
      this.add(`<g fill="${paper}" stroke="${paper}" stroke-width="${Math.round(o.halo / s)}" stroke-linejoin="round" transform="${place}">${uses}</g>`);
    }
    this.add(`<g fill="${this.color(o.ink)}" transform="${place}">${uses}</g>`);
    return width;
  },

  measure(str, o = {}) {
    return this.fonts[o.font ?? 'roman'].width(str, o.size ?? 20, o.track ?? 0);
  },

  // The size at which `str` fits in `max` px, never above o.size.
  fit(str, max, o = {}) {
    const size = o.size ?? 20;
    return Math.min(size, Math.floor((size * max * 10) / this.measure(str, o)) / 10);
  },

  // Greedy word wrap into lines no wider than `max` px.
  wrap(str, max, o = {}) {
    const lines = [];
    for (const word of str.split(' ')) {
      const last = lines[lines.length - 1];
      if (last !== undefined && this.measure(`${last} ${word}`, o) <= max) lines[lines.length - 1] = `${last} ${word}`;
      else lines.push(word);
    }
    // A list separator never starts or ends a line: the break itself separates.
    return lines.map((l) => l.replace(/ ·$/, '').replace(/^· /, ''));
  },

  // Wrapped paragraph; returns the baseline of the last line.
  para(str, x, y, o = {}) {
    const lead = o.lead ?? (o.size ?? 20) * 1.3;
    const lines = this.wrap(str, o.width ?? 400, o);
    lines.forEach((line, i) => this.text(line, x, y + i * lead, o));
    return y + (lines.length - 1) * lead;
  },

  // Small capitals faked from capitals: lower-case letters drawn as capitals at 78% size.
  caps(str, x, y, o = {}) {
    const size = o.size ?? 18;
    const track = o.track ?? 0.08;
    const runs = [];
    for (const ch of str) {
      const small = ch !== ch.toUpperCase();
      const last = runs[runs.length - 1];
      if (last && last.small === small) last.text += ch.toUpperCase();
      else runs.push({ small, text: ch.toUpperCase() });
    }
    const gap = track * size;
    const widths = runs.map((r) => this.measure(r.text, { ...o, size: r.small ? size * 0.78 : size, track }));
    const total = widths.reduce((a, b) => a + b, 0) + gap * (runs.length - 1);
    let cx = o.anchor === 'middle' ? x - total / 2 : o.anchor === 'end' ? x - total : x;
    runs.forEach((r, i) => {
      this.text(r.text, cx, y, { ...o, anchor: 'start', size: r.small ? size * 0.78 : size, track });
      cx += widths[i] + gap;
    });
    return total;
  },
};
