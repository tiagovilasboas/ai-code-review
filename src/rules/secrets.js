'use strict';

const { labelsFor } = require('../owasp-catalog');
const { compactFindings, createFinding, isCommentLine } = require('./common');

/**
 * @typedef {import('./common').Finding} Finding
 * @typedef {import('../parse-diff').DiffFile} DiffFile
 */

/** A field or constant used as a map key (finding.Secret:); its value is a label. */
const QUALIFIED_KEY =
  /[\w.]+\.(?:password|passwd|secret|token|api[_-]?key|private[_-]?key|access[_-]?key|credential)\s*:/gi;

const SECRET_ASSIGN =
  /(?:password|passwd|secret|token|api[_-]?key|private[_-]?key|access[_-]?key|credential)\s*[:=]\s*["'][^"']+["']/i;
const KNOWN_PREFIX = /["'](?:pk_live_|pk_test_|sk_live_|sk_test_|AKIA|ghp_|github_pat_|xox[baprs]-)/;
const CORS_STAR = /\bcors\s*\(\s*\{[^}]*origin\s*:\s*["']\*["']/;

/**
 * @param {DiffFile} file
 * @returns {Finding[]}
 */
function apply(file) {
  /** @type {(Finding | null)[]} */
  const rows = [];

  for (const line of file.added) {
    if (isCommentLine(line.text)) {
      continue;
    }
    if (/process\.env/.test(line.text)) {
      continue;
    }
    if (SECRET_ASSIGN.test(line.text.replace(QUALIFIED_KEY, '')) || KNOWN_PREFIX.test(line.text)) {
      rows.push(
        createFinding({
          title: 'Hardcoded credential in source',
          findingClass: 'secret',
          path: file.path,
          line: line.line,
          cwe: 'CWE-798',
          owasp: labelsFor('secret'),
          severity: 'CRITICAL',
        }),
      );
    }
    if (CORS_STAR.test(line.text)) {
      rows.push(
        createFinding({
          title: 'Overly permissive CORS origin',
          findingClass: 'cors',
          path: file.path,
          line: line.line,
          cwe: 'CWE-942',
          owasp: labelsFor('cors'),
          severity: 'MEDIUM',
        }),
      );
    }
  }

  return compactFindings(rows);
}

module.exports = { apply };
