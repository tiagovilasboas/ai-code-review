// Plate III: one run of the pr-review-process skill as eight engraved stations. The security engine
// runs first and any BLOCK stops the run; three hunters fan out in fresh context; verifiers vote.

import { Plate } from '../pen/plate.js';
import { arrow, arrowAlong, burst } from '../pen/ornaments.js';
import { medallion } from '../figures/solids.js';
import { sectionRef } from '../source.js';
import { sheet, figure, meta } from './sheet.js';

const Y = 250;
const R = 30;
const XS = [110, 250, 390, 545, 700, 850, 1010, 1165];
const SECURITY = 2;
const HUNTERS = 3;
const FILTER = 5;
const HUNTER_Y = [Y - 52, Y, Y + 52];

function stationNumber(p, n, x, y, ink) {
  p.text(String(n), x, y + 8, { size: 24, anchor: 'middle', ink });
}

function flow(p) {
  for (let i = 0; i < XS.length - 1; i++) {
    if (i === HUNTERS - 1 || i === HUNTERS) continue;
    arrow(p, [XS[i] + R + 6, Y], [XS[i + 1] - R - 6, Y], { width: 1.3 });
  }
  for (const hy of HUNTER_Y) {
    arrowAlong(p, [[XS[2] + R + 4, Y], [XS[3] - 30, hy]], { width: 1.1 });
    arrowAlong(p, [[XS[3] + 18, hy], [XS[4] - R - 6, Y]], { width: 1.1 });
  }
}

function stations(p, s, t) {
  t.stations.forEach(([name, detail], i) => {
    const x = XS[i];
    const ink = i === SECURITY ? 'accent' : 'ink';
    if (i === HUNTERS) {
      HUNTER_Y.forEach((hy) => medallion(p, x, hy, 14, { rim: 0.3 }));
      p.text(String(i + 1), x, Y + 5, { size: 14, anchor: 'middle' });
    } else {
      medallion(p, x, Y, R, { ink });
      stationNumber(p, i + 1, x, Y, ink);
    }
    const top = i === HUNTERS ? HUNTER_Y[2] + 46 : Y + R + 34;
    const width = i === HUNTERS ? 190 : 140;
    const nameEnd = p.para(name, x, top, { size: 19, width, lead: 22, anchor: 'middle', ink });
    const end = p.para(detail, x, nameEnd + 22, { size: 15, width, lead: 19, anchor: 'middle', font: 'italic', ink });
    if (i === FILTER) p.para(t.labels, x, end + 22, { size: 13.5, width: 140, lead: 17, anchor: 'middle' });
  });
  p.para(t.hunters, XS[HUNTERS], HUNTER_Y[0] - 30, { size: 14, width: 150, lead: 17, anchor: 'middle', font: 'italic' });
}

function block(p, t) {
  const x = XS[SECURITY];
  arrowAlong(p, [[x, Y - R - 6], [x, Y - 88]], { ink: 'accent', width: 1.5 });
  burst(p, x, Y - 96, { ink: 'accent', inner: 6, outer: 13 });
  const [word, ...rest] = t.block.split(': ');
  p.caps(word, x, Y - 150, { size: 17, anchor: 'middle', ink: 'accent', track: 0.12 });
  p.text(rest.join(': '), x, Y - 126, { size: 15, anchor: 'middle', font: 'italic', ink: 'accent' });
}

export function pipeline({ theme, fonts, s }) {
  const t = s.pipeline;
  const p = new Plate({ id: 'pl', width: 1280, height: 600, theme, fonts, seed: 1303, lang: s.lang, ...meta(s, t.svgTitle, t.desc) });
  const { bottom } = sheet(p, s, { n: 3, caption: t.caption, refs: [sectionRef(2)] });
  flow(p);
  stations(p, s, t);
  block(p, t);
  figure(p, s, 3, t.figure, 640, bottom - 4, { max: 1100, size: 17 });
  return p;
}
