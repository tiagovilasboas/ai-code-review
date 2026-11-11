// Builds every plate in English and Brazilian Portuguese, light and dark, into docs/art/ (pt-BR
// under docs/art/pt-BR/), copies the plates the site shows into site/art/, and writes the social
// card to site/og.svg and site/og.png. Output is deterministic: same sources, same bytes.
//
//   npm --prefix docs/art/generator run build
//   npm --prefix docs/art/generator run build -- --preview <dir>   (PNG of every SVG, for review)
//
// PNGs need rsvg-convert on PATH.

import { copyFileSync, mkdirSync, readFileSync, statSync, writeFileSync } from 'node:fs';
import { execFileSync } from 'node:child_process';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { THEMES } from './pen/theme.js';
import { loadFonts } from './pen/fonts.js';
import { PLATES } from './plates/index.js';

const here = dirname(fileURLToPath(import.meta.url));
const art = join(here, '..', '..');
const site = join(art, '..', '..', 'site');
const LOCALES = { en: '', 'pt-BR': 'pt-BR' };
const BUDGET = 200 * 1024;

const strings = (locale) => JSON.parse(readFileSync(join(here, 'strings', `${locale}.json`), 'utf8'));
const png = (svg, out, width) => execFileSync('rsvg-convert', ['-w', String(width), '-o', out, svg]);

function writePlate(dir, file, plate) {
  mkdirSync(dir, { recursive: true });
  const path = join(dir, file);
  writeFileSync(path, plate.toSVG());
  const size = statSync(path).size;
  if (size > BUDGET) throw new Error(`${path} is ${size} bytes, over the ${BUDGET} byte budget`);
  return path;
}

function build(preview) {
  const fonts = loadFonts();
  const written = [];
  for (const [locale, sub] of Object.entries(LOCALES)) {
    const s = strings(locale);
    for (const { name, make, themes = ['light', 'dark'], siteCopy } of PLATES) {
      for (const themeName of themes) {
        const file = `${name}-${themeName}.svg`;
        const path = writePlate(join(art, sub), file, make({ theme: THEMES[themeName], fonts, s }));
        written.push(path);
        if (locale === 'en' && siteCopy) {
          mkdirSync(join(site, 'art'), { recursive: true });
          copyFileSync(path, join(site, 'art', file));
        }
        if (locale === 'en' && name === 'og') {
          copyFileSync(path, join(site, 'og.svg'));
          png(path, join(site, 'og.png'), 1280);
        }
        if (preview) {
          mkdirSync(join(preview, locale), { recursive: true });
          png(path, join(preview, locale, file.replace(/\.svg$/, '.png')), 1280);
        }
      }
    }
  }
  for (const path of written) console.log(`${(statSync(path).size / 1024).toFixed(1).padStart(6)} KB  ${path.slice(art.length + 1)}`);
}

const at = process.argv.indexOf('--preview');
build(at > 0 ? process.argv[at + 1] : null);
