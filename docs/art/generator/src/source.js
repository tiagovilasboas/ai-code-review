// The source every plate credits: the Claude Code team's post at Anthropic. Its section headings
// are quoted verbatim, in English, in every locale. This repository is an independent
// implementation of that process, not affiliated with or endorsed by Anthropic.

export const POST_URL = 'https://claude.dev/blog/solving-code-review-at-anthropic/';

// The post's seven layer headings, in its order.
export const POST_HEADINGS = [
  'The work is verified before anyone else looks',
  'AI review on every pull request',
  'Security review that can block',
  'Automated checks that block',
  'Human approval, scaled by risk',
  'Merge queue, flags and a gradual release',
  'Fast detection, fast rollback',
];

// Second reference on the Cloudflare plate (product documentation topics, kept in English).
export const CLOUDFLARE_REF = 'Ref. developers.cloudflare.com: Workers secrets, Access JWT validation, WAF';

// "Ref. § <heading>" for the post section of layer n (1-based).
export const sectionRef = (n) => `Ref. § ${POST_HEADINGS[n - 1]}`;
