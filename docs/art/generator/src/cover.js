// Themes the README cover engraving for GitHub's light and dark pages, at desktop and phone size.
//
//   npm --prefix docs/art/generator run cover
//
// docs/art/cover/source.webp is the master: black ink and one clay accent on white paper (how it
// was made is in docs/art/README.md). Each pixel is split into an ink amount and an accent amount,
// and both are repainted on the page colour, so the two themes are the same drawing and the
// background matches the page. Needs rsvg-convert and cwebp on PATH.

import { copyFileSync, mkdtempSync, rmSync, statSync, writeFileSync } from 'node:fs';
import { execFileSync } from 'node:child_process';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { BARE } from './pen/theme.js';

const cover = join(dirname(fileURLToPath(import.meta.url)), '..', '..', 'cover');
const [W, H] = [2752, 1536];
// The accent as the master has it (measured from the file) and as the cover shows it.
const PAINTED = [204, 115, 80].map((v) => v / 255);
const CLAY = '#d97757';
const CUTS = { '': 1800, '-phone': 900 };
const QUALITY = 82;

const rgb = (hex) => [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16) / 255);
const row = (...v) => v.map((n) => n.toFixed(4)).join(' ');

// A pixel is white paper minus k of ink minus t of accent: R = 1 - k - (1 - aR)t, B = 1 - k - (1 - aB)t.
// Solving for k and t is linear, so one matrix puts k in the red channel and t in the green one.
// FLOOR drops the master's paper noise, so blank paper comes out as flat page colour, and the
// small gain on t brings the solid accent to full strength.
const FLOOR = 0.02;
function split() {
  const span = PAINTED[0] - PAINTED[2];
  const lift = (1 - PAINTED[0]) / span;
  const gain = 1.08 / span;
  const keep = 1 - FLOOR;
  return `<feColorMatrix type="matrix" values="${row(-(1 + lift) / keep, 0, lift / keep, 0, 1)} ${row(gain, 0, -gain, 0, -FLOOR)} 0 0 0 0 0 0 0 0 1 0"/>`;
}

// Paint k of ink and t of accent on the paper.
function paint({ paper, ink }) {
  const [bg, fg, acc] = [rgb(paper), rgb(ink), rgb(CLAY)];
  const rows = [0, 1, 2].map((c) => row(fg[c] - bg[c], acc[c] - bg[c], 0, 0, bg[c]));
  return `<feColorMatrix type="matrix" values="${rows.join(' ')} 0 0 0 1 0"/>`;
}

function themed(theme) {
  return [
    `<svg xmlns="http://www.w3.org/2000/svg" width="${W}" height="${H}">`,
    `<filter id="f" x="0" y="0" width="100%" height="100%" color-interpolation-filters="sRGB">${split()}${paint(theme)}</filter>`,
    `<rect width="${W}" height="${H}" fill="${theme.paper}"/>`,
    `<image width="${W}" height="${H}" href="source.webp" filter="url(#f)"/>`,
    '</svg>',
  ].join('');
}

function build() {
  const work = mkdtempSync(join(tmpdir(), 'cover-'));
  try {
    copyFileSync(join(cover, 'source.webp'), join(work, 'source.webp'));
    for (const name of ['light', 'dark']) {
      const svg = join(work, `${name}.svg`);
      const png = join(work, `${name}.png`);
      writeFileSync(svg, themed(BARE[name]));
      execFileSync('rsvg-convert', ['-o', png, svg]);
      for (const [suffix, width] of Object.entries(CUTS)) {
        const out = join(cover, `cover${suffix}-${name}.webp`);
        execFileSync('cwebp', ['-quiet', '-q', String(QUALITY), '-m', '6', '-sharp_yuv', '-resize', String(width), '0', png, '-o', out]);
        console.log(`${(statSync(out).size / 1024).toFixed(1).padStart(6)} KB  cover/cover${suffix}-${name}.webp`);
      }
    }
  } finally {
    rmSync(work, { recursive: true, force: true });
  }
}

build();
