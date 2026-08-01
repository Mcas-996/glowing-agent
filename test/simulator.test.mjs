import assert from 'node:assert/strict';
import test from 'node:test';
import { generateSimulation, thinkingGrammar } from '../static/simulator.mjs';

test('a task and seed always reproduce the same simulation', () => {
  assert.deepEqual(generateSimulation('Fix the typo', 42, 'none', 1000), generateSimulation('Fix the typo', 42, 'none', 1000));
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
  const none = generateSimulation(task, 42, 'none', 1000);
  const low = generateSimulation(task, 42, 'low', 1000);
  const high = generateSimulation(task, 42, 'high', 1000);
  const thoughtAt = (simulation, index) => simulation.events.filter((event) => event.kind === 'thought')[index];

  assert.equal(thoughtAt(low, 0).delayMs, thoughtAt(none, 0).delayMs * 2);
  assert.equal(thoughtAt(high, 0).delayMs, thoughtAt(none, 0).delayMs * 8);
  assert.equal(thoughtAt(high, 0).text.split('\n').length, 8);
  assert.equal(high.metrics.tokensBurned, none.metrics.tokensBurned * 8);
});

test('thinking grammar has seven fixed segments with thirteen seeded choices each', () => {
  assert.equal(thinkingGrammar.length, 7);
  assert.ok(thinkingGrammar.every((alternatives) => alternatives.length === 13));
});

test('thinking phrases are reproducible from the simulation seed', () => {
  const first = generateSimulation('Fix the typo', 42, 'high');
  const second = generateSimulation('Fix the typo', 42, 'high');
  const thoughts = first.events.filter((event) => event.kind === 'thought');

  assert.deepEqual(first.events, second.events);
  assert.equal(first.ending, second.ending);
  assert.ok(thoughts.every((event) => !event.text.includes('[reflection')));
  assert.ok(thoughts.every((event) => !/\b\d+\/\d+\b/.test(event.text)));
});
