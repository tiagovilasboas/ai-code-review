// Loads EB Garamond (SIL OFL 1.1, @fontsource/eb-garamond, pinned in package.json) and turns
// text into outlined glyph paths, so the SVGs need no font (a GitHub <img> cannot load one).

import { createRequire } from 'node:module';
import opentype from 'opentype.js';

const require = createRequire(import.meta.url);

const FACES = {
  roman: '@fontsource/eb-garamond/files/eb-garamond-latin-500-normal.woff',
  bold: '@fontsource/eb-garamond/files/eb-garamond-latin-600-normal.woff',
  italic: '@fontsource/eb-garamond/files/eb-garamond-latin-400-italic.woff',
};

const ID_PREFIX = { roman: 'r', bold: 'b', italic: 'i' };

export class Face {
  constructor(key, font) {
    this.key = key;
    this.font = font;
    this.upm = font.unitsPerEm;
  }

  // Glyph positions for one line in font units, with kerning and optional tracking (em).
  layout(str, track = 0) {
    const glyphs = this.font.stringToGlyphs(str);
    const extra = track * this.upm;
    const out = [];
    let x = 0;
    glyphs.forEach((glyph, i) => {
      out.push({ glyph, x });
      x += glyph.advanceWidth;
      const next = glyphs[i + 1];
      if (next) x += this.font.getKerningValue(glyph, next) + extra;
    });
    return { glyphs: out, advance: x };
  }

  width(str, size, track = 0) {
    return (this.layout(str, track).advance * size) / this.upm;
  }
}

export function loadFonts() {
  const fonts = {};
  for (const [key, spec] of Object.entries(FACES)) {
    fonts[key] = new Face(key, opentype.loadSync(require.resolve(spec)));
  }
  return fonts;
}

// Glyph outlines are stored at half resolution (2 font units per step, scaled back on the <path>)
// as relative commands with no repeated letters: about half the bytes of opentype's absolute form.
const GRID = 2;

function joinNumbers(values) {
  return values.map((v, i) => (i > 0 && v >= 0 ? ` ${v}` : String(v))).join('');
}

function compactD(commands) {
  const q = (v) => Math.round(v / GRID);
  let [cx, cy, sx, sy] = [0, 0, 0, 0];
  let last = '';
  let out = '';
  const emit = (letter, values) => {
    out += (letter === last && letter !== 'z' ? (values[0] >= 0 ? ' ' : '') : letter) + joinNumbers(values);
    last = letter === 'm' ? 'l' : letter;
  };
  for (const c of commands) {
    if (c.type === 'Z') { emit('z', []); [cx, cy] = [sx, sy]; continue; }
    const [x, y] = [q(c.x), q(c.y)];
    if (c.type === 'M') { emit('m', [x - cx, y - cy]); [sx, sy] = [x, y]; }
    else if (c.type === 'L') emit('l', [x - cx, y - cy]);
    else if (c.type === 'Q') emit('q', [q(c.x1) - cx, q(c.y1) - cy, x - cx, y - cy]);
    else if (c.type === 'C') emit('c', [q(c.x1) - cx, q(c.y1) - cy, q(c.x2) - cx, q(c.y2) - cy, x - cx, y - cy]);
    [cx, cy] = [x, y];
  }
  return out;
}

// Each glyph is stored once per document as <path id> in <defs>; text reuses it with <use>.
export class GlyphBook {
  constructor(prefix) {
    this.prefix = prefix;
    this.paths = new Map();
  }

  idFor(face, glyph) {
    const id = `${this.prefix}${ID_PREFIX[face.key]}${glyph.index}`;
    if (!this.paths.has(id)) this.paths.set(id, compactD(glyph.getPath(0, 0, face.upm).commands));
    return this.paths.get(id) ? id : null;
  }

  defs() {
    return [...this.paths].filter(([, d]) => d).map(([id, d]) => `<path id="${id}" transform="scale(${GRID})" d="${d}"/>`).join('');
  }
}
