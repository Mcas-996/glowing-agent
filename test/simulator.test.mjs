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

test('higher thinking depths grow thought content and time exponentially', () => {
  const task = 'Fix the typo';
  const none = generateSimulation(task, 42, 'none');
  const low = generateSimulation(task, 42, 'low');
  const high = generateSimulation(task, 42, 'high');
  const thoughtAt = (simulation, index) => simulation.events.filter((event) => event.kind === 'thought')[index];

  assert.equal(thoughtAt(low, 0).delayMs, thoughtAt(none, 0).delayMs * 2);
  assert.equal(thoughtAt(high, 0).delayMs, thoughtAt(none, 0).delayMs * 8);
  assert.equal(thoughtAt(high, 0).text.split('\n').length, 8);
  assert.equal(high.metrics.tokensBurned, none.metrics.tokensBurned * 8);
});
