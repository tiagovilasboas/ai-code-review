// One slice of the swiss-cheese model, engraved: an upright slab in cabinet projection with
// its depth receding up and to the left, so the broad face looks left, toward what comes in.
// Holes are tunnels through the slab: a dark cross-hatched wall, paper where you see through.

import { num, polyD } from '../pen/geometry.js';

export function slab({ xf, yb, h, depth = 50, rise = 35, t = 8 }) {
  const k = rise / depth;
  const F0 = [xf, yb - h];
  const F1 = [xf - depth, yb - h - rise];
  const F2 = [xf - depth, yb - rise];
  const F3 = [xf, yb];
  const off = ([x, y], dx = t) => [x + dx, y];
  return {
    xf, yb, h, depth, rise, t, k,
    face: [F0, F1, F2, F3],
    top: [F0, F1, off(F1), off(F0)],
    front: [F0, off(F0), off(F3), F3],
    silhouette: [F1, off(F1), off(F0), off(F3), F3, F2],
    edges: { F0, F1, F2, F3, T0: off(F0), T1: off(F1), T3: off(F3) },
    matrix: `matrix(-1 ${num(-k)} 0 1 ${num(xf)} ${num(yb - h)})`,
    // Face coordinates (u back from the front edge, v down from the top) for a screen point.
    local: (x, y) => [xf - x, y - (yb - h) + k * (xf - x)],
  };
}

// A hole centred on screen point (x, y), sized in face units.
export function holeAt(s, x, y, rx, ry) {
  const [u, v] = s.local(x, y);
  return { u, v, rx, ry };
}

export function drawShadow(plate, s, length = 34) {
  const { F2, F3 } = s.edges;
  const back = [F2[0] + s.t, F2[1]];
  const front = [F3[0] + s.t, F3[1]];
  plate.fill(polyD([front, back, [back[0] + length, back[1]], [front[0] + length, front[1]]]), plate.tone(4, 45));
}

function holeMarkup(plate, s, { u, v, rx, ry }) {
  const e = (cx, cy, attrs) => `<ellipse cx="${num(cx)}" cy="${num(cy)}" rx="${rx}" ry="${ry}" ${attrs}/>`;
  const clip = plate.clip(e(u, v, ''));
  const paper = plate.color('paper');
  return e(u, v, `fill="${paper}"`)
    + `<g clip-path="${clip}">${e(u, v, `fill="${plate.tone(5, 90)}"`)}${e(u - s.t, v - s.t * s.k, `fill="${paper}"`)}</g>`
    + e(u, v, `fill="none" stroke="${plate.color()}" stroke-width="${plate.w(1.1)}"`);
}

export function drawSlab(plate, s, holes = []) {
  plate.fill(polyD(s.silhouette), plate.color('paper'));
  // Base tone follows the perspective; a second set of lines darkens the far half.
  const rects = `<rect width="${s.depth}" height="${s.h}" fill="${plate.tone(2, 0)}"/>`
    + `<rect x="${s.depth * 0.55}" width="${s.depth * 0.45}" height="${s.h}" fill="${plate.tone(1, 90)}"/>`;
  plate.add(`<g transform="${s.matrix}">${rects}${holes.map((hole) => holeMarkup(plate, s, hole)).join('')}</g>`);
  plate.fill(polyD(s.top), plate.tone(1, 35));
  plate.fill(polyD(s.front), plate.tone(3, 90));
  const { F0, F1, F2, F3, T0, T1, T3 } = s.edges;
  plate.swell([F1, F0], 1.1);
  plate.swell([F1, F2], 1);
  plate.swell([F2, F3], 1.5);
  plate.swell([F0, F3], 1.1);
  plate.swell([T1, T0], 0.9);
  plate.swell([F0, T0], 1);
  plate.swell([T0, T3], 2);
  plate.swell([F3, T3], 1.6);
}

// Mask for something travelling along +x at depth `u` (face units back from the front edge):
// only the part of the slab in front of it hides it, and the hole lets it show through.
export function maskMarkup(s, u, hole) {
  const { xf, yb, h, t, k } = s;
  const front = [[xf - u, yb - h - k * u], [xf - u + t, yb - h - k * u], s.edges.T0, s.edges.T3, s.edges.F3, [xf - u, yb - k * u]];
  const opening = hole
    ? `<g transform="${s.matrix}"><ellipse cx="${num(hole.u)}" cy="${num(hole.v)}" rx="${hole.rx}" ry="${hole.ry}" fill="#fff"/></g>`
    : '';
  return `<path d="${polyD(front)}" fill="#000"/>${opening}`;
}
