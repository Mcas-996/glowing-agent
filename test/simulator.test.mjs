import assert from 'node:assert/strict';
import test from 'node:test';
import { generateSimulation } from '../static/simulator.mjs';

test('a task and seed always reproduce the same simulation', () => {
  assert.deepEqual(generateSimulation('Fix the typo', 42), generateSimulation('Fix the typo', 42));
});

test('the simulator can produce every ending', () => {
  const endings = new Set();
  for (let seed = 0; seed < 1000 && endings.size < 3; seed += 1) {
    endings.add(generateSimulation('Add a button', seed).ending);
  }
  assert.deepEqual(endings, new Set(['confident-miss', 'scope-singularity', 'accidental-win']));
});
