// Number formatting and the engraver's swelling line.

export const num = (n) => {
  const r = Math.round(n * 10) / 10;
  return Object.is(r, -0) ? '0' : String(r);
};

export const pts = (points) => points.map(([x, y]) => `${num(x)} ${num(y)}`).join(' ');

export const polyD = (points) => `M${pts(points)}Z`;

// A line that swells in the middle and thins at both ends, returned as a filled outline.
// `taper` keeps a fraction of the width at the ends (0 = needle points).
export function swellD(points, width, taper = 0.25, steps = 10) {
  const seg = [];
  for (let i = 0; i < points.length - 1; i++) {
    const [x1, y1] = points[i];
    const [x2, y2] = points[i + 1];
    for (let k = 0; k < steps; k++) {
      const s = k / steps;
      seg.push([x1 + (x2 - x1) * s, y1 + (y2 - y1) * s]);
    }
  }
  seg.push(points[points.length - 1]);
  const n = seg.length - 1;
  const left = [];
  const right = [];
  seg.forEach(([x, y], i) => {
    const [ax, ay] = seg[Math.max(0, i - 1)];
    const [bx, by] = seg[Math.min(n, i + 1)];
    const len = Math.hypot(bx - ax, by - ay) || 1;
    const nx = -(by - ay) / len;
    const ny = (bx - ax) / len;
    const half = (width / 2) * (taper + (1 - taper) * Math.sin((Math.PI * i) / n));
    left.push([x + nx * half, y + ny * half]);
    right.push([x - nx * half, y - ny * half]);
  });
  return polyD([...left, ...right.reverse()]);
}
