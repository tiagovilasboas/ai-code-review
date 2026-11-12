// Vignette for "Guardrails": three engraved columns carry one lintel, one guardrail each.
// Borderless: no sheet, no frame, no caption; labels come from strings.vGuardrails.

import { Plate } from '../pen/plate.js';
import { polyD } from '../pen/geometry.js';
import { block } from '../figures/solids.js';

const CX = [270, 640, 1010];
const TOP = 104;
const [Y1, Y2] = [126, 300];
const GROUND = 322;

function capital(p, cx) {
  p.fill(polyD([[cx - 44, 114], [cx + 44, 114], [cx + 56, TOP + 10], [cx - 56, TOP + 10]]), p.color('paper'));
  p.fill(polyD([[cx - 44, 114], [cx + 44, 114], [cx + 40, 126], [cx - 40, 126]]), p.tone(2, 0));
  p.swell([[cx - 62, TOP], [cx + 62, TOP]], 1.6, 'ink', 0.1);
  p.swell([[cx - 62, TOP + 10], [cx + 62, TOP + 10]], 1.4, 'ink', 0.1);
  p.swell([[cx - 62, TOP], [cx - 62, TOP + 10]], 1, 'ink', 0.3);
  p.swell([[cx + 62, TOP], [cx + 62, TOP + 10]], 1, 'ink', 0.3);
  p.stroke(`M${cx - 44} 114Q${cx - 46} 122 ${cx - 40} 126M${cx + 44} 114Q${cx + 46} 122 ${cx + 40} 126`, 1.1);
}

// Shaft with slight swelling (entasis), flutes, and a dark side away from the light.
function shaft(p, cx) {
  const left = (y) => cx - 40 - 1.6 * Math.sin(((y - Y1) / (Y2 - Y1)) * Math.PI) - ((y - Y1) / (Y2 - Y1)) * 2;
  const rows = Array.from({ length: 9 }, (_, i) => Y1 + ((Y2 - Y1) * i) / 8);
  const L = rows.map((y) => [left(y), y]);
  const R = rows.map((y) => [2 * cx - left(y), y]);
  const outline = polyD([...L, ...[...R].reverse()]);
  p.fill(outline, p.color('paper'));
  p.fill(outline, p.tone(1, 90));
  const clip = p.clip(`<path d="${outline}"/>`);
  p.fill(polyD([[cx + 10, Y1], [cx + 60, Y1], [cx + 60, Y2], [cx + 10, Y2]]), p.tone(3, 90), ` clip-path="${clip}"`);
  for (let k = -4; k <= 4; k++) p.swell(rows.map((y) => [cx + (k / 4.5) * (cx - left(y)), y]), 0.8, 'ink', 0.1, 3);
  p.swell(L, 1.3, 'ink', 0.2, 3);
  p.swell(R, 1.6, 'ink', 0.2, 3);
}

function base(p, cx) {
  p.fill(polyD([[cx - 52, Y2], [cx + 52, Y2], [cx + 56, 310], [cx - 56, 310]]), p.tone(2, 0));
  p.swell([[cx - 56, 310], [cx + 56, 310]], 1.3, 'ink', 0.1);
  block(p, { x: cx - 66, y: 310, w: 132, h: 12, d: 12, front: [3, 0] });
}

function pediment(p) {
  const tri = [[150, 58], [1160, 58], [655, 12]];
  p.fill(polyD(tri), p.color('paper'));
  p.fill(polyD(tri), p.tone(1, 25));
  p.swell([tri[0], tri[2], tri[1]], 1.6, 'ink', 0.1);
  p.swell([[215, 52], [655, 22], [1095, 52]], 0.9, 'ink', 0.1);
}

function ground(p) {
  p.swell([[70, GROUND], [1210, GROUND]], 1.8, 'ink', 0.05);
  [[6, 0.9, 40], [11, 0.7, 120], [16, 0.5, 220]].forEach(([dy, w, inset]) => p.swell([[70 + inset, GROUND + dy], [1210 - inset, GROUND + dy]], w, 'ink', 0.05));
}

function labels(p, t) {
  t.columns.forEach(([name, rest], i) => {
    p.fill(polyD([[CX[i] - 5, 363], [CX[i], 358], [CX[i] + 5, 363], [CX[i], 368]]), p.color());
    p.caps(name, CX[i], 394, { size: 20, anchor: 'middle', track: 0.1 });
    p.para(rest, CX[i], 420, { size: 17, width: 330, lead: 21, anchor: 'middle', font: 'italic' });
  });
}

export function vGuardrails({ theme, fonts, s }) {
  const t = s.vGuardrails;
  const p = new Plate({ id: 'vg', width: 1280, height: 446, theme, fonts, seed: 2711, lang: s.lang, title: t.svgTitle, desc: t.desc });
  ground(p);
  CX.forEach((cx) => { shaft(p, cx); capital(p, cx); base(p, cx); });
  block(p, { x: 140, y: 70, w: 1000, h: 34, d: 20, front: [1, 0] });
  pediment(p);
  labels(p, t);
  return p;
}
