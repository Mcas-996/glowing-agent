import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const html = await readFile(new URL('../static/index.html', import.meta.url), 'utf8');
const css = await readFile(new URL('../static/style.css', import.meta.url), 'utf8');

test('browser demo uses the terminal workbench structure', () => {
  const workspace = html.indexOf('class="workspace"');
  const terminal = html.indexOf('id="terminal"');
  const composer = html.indexOf('class="composer"');
  const sidebar = html.indexOf('class="sidebar"');

  assert.ok(workspace >= 0);
  assert.ok(terminal > workspace);
  assert.ok(composer > terminal);
  assert.ok(sidebar > composer);
  assert.match(html, /class="section-title"/);
  assert.match(html, /id="result" class="result hidden"/);
});

test('browser workbench folds its sidebar on narrow screens', () => {
  assert.match(css, /grid-template-columns:\s*minmax\(0, 1fr\) 320px/);
  assert.match(css, /@media \(max-width: 900px\)/);
  assert.match(css, /\.sidebar\s*\{[\s\S]*border-left:/);
});
