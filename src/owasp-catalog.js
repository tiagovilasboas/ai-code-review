'use strict';

/**
 * Official OWASP identifiers only. Do not add an ID that is not in a cited source.
 *
 * Sources (retrieved for this kit; do not invent catalog numbers):
 * - OWASP Top 10:2021 — https://owasp.org/Top10/
 * - OWASP API Security Top 10:2023 — https://owasp.org/API-Security/editions/2023/en/
 * - OWASP ASVS 5.0.0 requirement tables — https://github.com/OWASP/ASVS/tree/v5.0.0/5.0
 * - OWASP Top 10 for Agentic Applications 2026 (ASI01:2026–ASI10:2026)
 *   https://genai.owasp.org/resource/owasp-top-10-for-agentic-applications-for-2026/
 * - OWASP Agentic Skills Top 10 (AST01–AST10)
 *   https://owasp.org/www-project-agentic-skills-top-10/
 */

const OFFICIAL_ID_RE =
  /^(?:A0[1-9]:2021|A10:2021|API[1-9]:2023|ASVS-5\.0-\d+\.\d+\.\d+|ASI[0-9]{2}:2026|AST[0-9]{2})$/;

/**
 * CLI labels for a finding class. Web / API / ASVS only — those IDs describe
 * the sink in the diff. Agentic IDs (ASI / AST) are reviewer posture and live
 * on the skill docs, not on application finding lines.
 *
 * @type {Record<string, string[]>}
 */
const CLI_LABELS = {
  authz: ['A01:2021', 'API1:2023', 'ASVS-5.0-8.2.2'],
  secret: ['A02:2021', 'ASVS-5.0-13.3.1'],
  cors: ['A05:2021', 'ASVS-5.0-3.4.2'],
  xss: ['A03:2021', 'ASVS-5.0-1.3.1'],
  ssrf: ['A10:2021', 'ASVS-5.0-1.3.6'],
  'install-script': ['A08:2021', 'ASVS-5.0-15.2.4'],
  'registry-host': ['A08:2021', 'ASVS-5.0-15.2.4'],
  'unpinned-action': ['A08:2021', 'ASVS-5.0-15.2.4'],
  'lockfile-rewrite': ['A06:2021', 'ASVS-5.0-15.1.2'],
  'untrusted-diff': ['ASI01:2026', 'AST05'],
  // Go CLI only (the JS engine has no injection rules). ASVS 5.0.0 1.2.4 / 1.2.5 / 5.3.2.
  // Path traversal is A01:2021: Top 10:2021 lists CWE-22 under Broken Access Control.
  'sql-injection': ['A03:2021', 'ASVS-5.0-1.2.4'],
  'command-injection': ['A03:2021', 'ASVS-5.0-1.2.5'],
  'path-traversal': ['A01:2021', 'ASVS-5.0-5.3.2'],
};

/**
 * Per-skill map for docs and agents. `app` is what the CLI may print.
 * `posture` is how the reviewing agent must behave (not a finding label).
 *
 * @type {Record<string, { app: string[], posture: string[] }>}
 */
const SKILL_MAP = {
  'authz-idor': {
    app: CLI_LABELS.authz,
    posture: ['ASI09:2026', 'AST05'],
  },
  'secrets-config': {
    app: [...CLI_LABELS.secret, ...CLI_LABELS.cors],
    posture: ['ASI09:2026', 'AST03'],
  },
  'xss-html': {
    app: CLI_LABELS.xss,
    posture: ['ASI09:2026', 'AST05'],
  },
  'ssrf-egress': {
    app: CLI_LABELS.ssrf,
    posture: ['ASI02:2026', 'AST03'],
  },
  'supply-chain': {
    app: ['A06:2021', 'A08:2021', 'ASVS-5.0-15.1.2', 'ASVS-5.0-15.2.4'],
    posture: ['ASI04:2026', 'AST02'],
  },
  injection: {
    app: [
      ...CLI_LABELS['sql-injection'],
      ...CLI_LABELS['command-injection'].slice(1),
      ...CLI_LABELS['path-traversal'],
    ],
    posture: ['ASI02:2026', 'AST03'],
  },
  'untrusted-diff': {
    app: CLI_LABELS['untrusted-diff'],
    posture: ['ASI01:2026', 'AST05'],
  },
};

/**
 * @param {string} id
 * @returns {boolean}
 */
function isOfficialId(id) {
  return OFFICIAL_ID_RE.test(id);
}

/**
 * @param {string} key
 * @returns {string[]}
 */
function labelsFor(key) {
  const ids = CLI_LABELS[key];
  if (!ids || ids.length === 0) {
    return [];
  }
  if (!ids.every(isOfficialId)) {
    return [];
  }
  return ids.slice();
}

/**
 * Fail-closed: reject any list that is empty or contains a non-official id.
 *
 * @param {string[]} ids
 * @returns {string[] | null}
 */
function assertOfficialIds(ids) {
  if (!Array.isArray(ids) || ids.length === 0) {
    return null;
  }
  if (!ids.every((id) => typeof id === 'string' && isOfficialId(id))) {
    return null;
  }
  return ids.slice();
}

module.exports = {
  OFFICIAL_ID_RE,
  CLI_LABELS,
  SKILL_MAP,
  isOfficialId,
  labelsFor,
  assertOfficialIds,
};
