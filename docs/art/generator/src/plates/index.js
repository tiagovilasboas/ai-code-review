// The plates, in order. siteCopy: also published under site/art/ for the project site.
// bare: a borderless README vignette, drawn with the BARE themes (no frame, no background).
// A `-phone` plate is the same vignette cut for a phone: 720 wide, no text under 32 px.

import { hero } from './hero.js';
import { section } from './section.js';
import { pipeline } from './pipeline.js';
import { tiers } from './tiers.js';
import { convergence } from './convergence.js';
import { harness } from './harness.js';
import { loop } from './loop.js';
import { edge } from './edge.js';
import { og } from './og.js';
import { vDeterministic, vDeterministicPhone } from './v-deterministic.js';
import { vQuickstart, vQuickstartPhone } from './v-quickstart.js';
import { vCatches, vCatchesPhone } from './v-catches.js';
import { vEngine, vEnginePhone } from './v-engine.js';
import { vInstall, vInstallPhone } from './v-install.js';
import { vGuardrails, vGuardrailsPhone } from './v-guardrails.js';

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
  { name: 'v-deterministic', make: vDeterministic, bare: true },
  { name: 'v-quickstart', make: vQuickstart, bare: true },
  { name: 'v-catches', make: vCatches, bare: true },
  { name: 'v-engine', make: vEngine, bare: true },
  { name: 'v-install', make: vInstall, bare: true },
  { name: 'v-guardrails', make: vGuardrails, bare: true },
  { name: 'v-deterministic-phone', make: vDeterministicPhone, bare: true },
  { name: 'v-quickstart-phone', make: vQuickstartPhone, bare: true },
  { name: 'v-catches-phone', make: vCatchesPhone, bare: true },
  { name: 'v-engine-phone', make: vEnginePhone, bare: true },
  { name: 'v-install-phone', make: vInstallPhone, bare: true },
  { name: 'v-guardrails-phone', make: vGuardrailsPhone, bare: true },
];
