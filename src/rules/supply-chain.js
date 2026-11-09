'use strict';

const { labelsFor } = require('../owasp-catalog');
const { compactFindings, createFinding, isCommentLine } = require('./common');

/**
 * @typedef {import('./common').Finding} Finding
 * @typedef {import('../parse-diff').DiffFile} DiffFile
 */

const FIRST_PARTY_ACTIONS = /^(?:actions|github)\//;
const PINNED_SHA = /@[0-9a-f]{40}$/i;
const EXPECTED_NPM_HOST = /registry\.npmjs\.org/;

/**
 * @param {string} filePath
 * @returns {string}
 */
function basename(filePath) {
  const parts = filePath.split('/');
  return parts[parts.length - 1] || filePath;
}

/**
 * @param {DiffFile} file
 * @returns {Finding[]}
 */
function apply(file) {
  /** @type {(Finding | null)[]} */
  const rows = [];
  const name = basename(file.path);

  if (name === 'package.json') {
    for (const line of file.added) {
      if (isCommentLine(line.text)) {
        continue;
      }
      if (!/"(?:postinstall|preinstall|install)"\s*:/.test(line.text)) {
        continue;
      }
      rows.push(
        createFinding({
          title: 'Install-time script',
          findingClass: 'install-script',
          path: file.path,
          line: line.line,
          cwe: 'CWE-829',
          owasp: labelsFor('install-script'),
          severity: 'HIGH',
        }),
      );
    }
  }

  if (name === '.npmrc' || name.endsWith('.npmrc')) {
    for (const line of file.added) {
      if (isCommentLine(line.text)) {
        continue;
      }
      const match = line.text.match(/^(?:registry|@[\w-]+:registry)\s*=\s*(\S+)/);
      if (!match) {
        continue;
      }
      if (EXPECTED_NPM_HOST.test(match[1])) {
        continue;
      }
      rows.push(
        createFinding({
          title: 'Unexpected npm registry host',
          findingClass: 'registry-host',
          path: file.path,
          line: line.line,
          cwe: 'CWE-829',
          owasp: labelsFor('registry-host'),
          severity: 'HIGH',
        }),
      );
    }
  }

  if (file.path.includes('.github/workflows/') && /\.ya?ml$/.test(file.path)) {
    for (const line of file.added) {
      if (isCommentLine(line.text)) {
        continue;
      }
      const uses = line.text.match(/\buses:\s*(\S+)/);
      if (uses) {
        const action = uses[1];
        if (action.startsWith('./') || FIRST_PARTY_ACTIONS.test(action) || PINNED_SHA.test(action)) {
          continue;
        }
        rows.push(
          createFinding({
            title: 'Unpinned third-party GitHub Action',
            findingClass: 'unpinned-action',
            path: file.path,
            line: line.line,
            cwe: 'CWE-829',
            owasp: labelsFor('unpinned-action'),
            severity: 'HIGH',
          }),
        );
      }
      if (/\brun:\s*npm\s+install\b/.test(line.text) && !/--frozen-lockfile/.test(line.text)) {
        rows.push(
          createFinding({
            title: 'CI install can rewrite the lockfile',
            findingClass: 'lockfile-rewrite',
            path: file.path,
            line: line.line,
            cwe: 'CWE-1104',
            owasp: labelsFor('lockfile-rewrite'),
            severity: 'MEDIUM',
          }),
        );
      }
    }
  }

  if (/(?:package-lock\.json|yarn\.lock|pnpm-lock\.yaml)$/.test(file.path)) {
    for (const line of file.added) {
      const resolved = line.text.match(/"resolved"\s*:\s*"([^"]+)"/);
      if (resolved && !EXPECTED_NPM_HOST.test(resolved[1])) {
        rows.push(
          createFinding({
            title: 'Lockfile resolved to unexpected registry',
            findingClass: 'registry-host',
            path: file.path,
            line: line.line,
            cwe: 'CWE-829',
            owasp: labelsFor('registry-host'),
            severity: 'HIGH',
          }),
        );
      }
    }
  }

  return compactFindings(rows);
}

module.exports = { apply };
