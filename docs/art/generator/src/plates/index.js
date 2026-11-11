// The plates, in order. siteCopy: also published under site/art/ for the project site.

import { hero } from './hero.js';
import { section } from './section.js';
import { pipeline } from './pipeline.js';
import { tiers } from './tiers.js';
import { convergence } from './convergence.js';
import { harness } from './harness.js';
import { loop } from './loop.js';
import { edge } from './edge.js';
import { og } from './og.js';

export const PLATES = [
  { name: 'hero', make: hero },
  { name: 'seven-layers', make: section, siteCopy: true },
  { name: 'review-pipeline', make: pipeline, siteCopy: true },
  { name: 'risk-tiers', make: tiers },
  { name: 'convergence', make: convergence },
  { name: 'harness-map', make: harness, siteCopy: true },
  { name: 'learning-loop', make: loop },
  { name: 'cloudflare-edge', make: edge, siteCopy: true },
  { name: 'og', make: og, themes: ['dark'] },
];
