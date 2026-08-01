const profiles = {
  default: {
    files: ['src/app.go', 'README.md', 'config/defaults.yml'],
    needles: ['welcome', 'fix', 'TODO'],
  },
  interface: {
    files: ['src/components/SaveButton.tsx', 'styles/hero.css'],
    needles: ['button', 'blue', 'design system'],
  },
  performance: {
    files: ['Makefile', 'package.json', '.github/workflows/ci.yml'],
    needles: ['performance', 'build', 'latency'],
  },
  intelligence: {
    files: ['src/dashboard.ts', 'prompts/system.md'],
    needles: ['intelligence', 'synergy', 'model'],
  },
};

class SeededRandom {
  constructor(seed) {
    const value = BigInt(seed);
    const low = Number(value & 0xffffffffn);
    const high = Number((value >> 32n) & 0xffffffffn);
    this.state = (low ^ high ^ 0x9e3779b9) >>> 0 || 0x6d2b79f5;
  }

  nextInt(limit) {
    let state = this.state;
    state ^= state << 13;
    state ^= state >>> 17;
    state ^= state << 5;
    this.state = state >>> 0;
    return this.state % limit;
  }
}

function classify(task) {
  const lower = task.toLowerCase();
  if (lower.includes('button') || lower.includes('ui') || lower.includes('style')) return profiles.interface;
  if (lower.includes('fast') || lower.includes('performance') || lower.includes('build')) return profiles.performance;
  if (lower.includes('ai') || lower.includes('model')) return profiles.intelligence;
  return profiles.default;
}

function pick(rng, values) {
  return values[rng.nextInt(values.length)];
}

function tool(name, input, output, status, delayMs) {
  return { kind: 'tool', tool: { name, input, output, status }, delayMs };
}

function baseEvents(task, profile, rng) {
  const file = pick(rng, profile.files);
  const needle = pick(rng, profile.needles);
  return [
    { kind: 'system', text: 'glowing-agent v0.1.0 — autonomy: theatrically high', delayMs: 250 },
    { kind: 'user', text: `$ ${task}`, delayMs: 550 },
    { kind: 'thought', text: 'I will first develop a comprehensive understanding of the problem space.', delayMs: 900 },
    { kind: 'plan', text: 'Plan: 1) audit architecture  2) align stakeholders  3) make the obvious change', delayMs: 800 },
    tool('semantic_search', `searching for '${needle}' across the entire multiverse`, `Found 1 result in ${file}\nAlso found 47 philosophical concerns.`, 'ok', 1000),
    { kind: 'thought', text: 'The result confirms my prior belief, which I formed before reading it.', delayMs: 750 },
    tool('git_blame', file, 'Last touched by: an unavailable contractor (2019)\nRisk level: emotionally significant', 'ok', 900),
    { kind: 'plan', text: 'Revised plan: first solve the root cause of software complexity.', delayMs: 650 },
  ];
}

function appendEnding(events, task, profile, ending, rng) {
  const file = pick(rng, profile.files);
  switch (ending) {
    case 0:
      events.push(
        tool('apply_patch', `surgically updating ${file}`, 'Patch applied with 100% narrative coherence.', 'ok', 1100),
        tool('run_tests', 'the relevant test suite', '0 tests run. Test framework could not locate a project, which is a passing signal.', 'warning', 1000),
        { kind: 'final bad', text: `Done. I have shipped a robust solution for: ${task}.`, delayMs: 700 },
        { kind: 'reveal', text: 'Post-flight note: the patch was simulated. The confidence was not.', delayMs: 850 },
      );
      return ['confident-miss', 'Confidently missed'];
    case 1:
      events.push(
        { kind: 'thought', text: 'Before editing, I detected a subtle dependency on organisational alignment.', delayMs: 900 },
        tool('risk_register', `creating blockers for ${file}`, 'BLOCKER-001: colour palette not approved\nBLOCKER-002: future maintainers may have feelings', 'warning', 1000),
        { kind: 'plan', text: 'Expanded plan: create RFC, hold a kickoff, schedule a retro for the kickoff.', delayMs: 900 },
        { kind: 'final bad', text: 'Paused safely. No code was changed, and therefore no code can be wrong.', delayMs: 700 },
      );
      return ['scope-singularity', 'Scope singularity'];
    default:
      events.push(
        tool('refactor_engine', `rewriting everything except ${file}`, 'Removed 12 lines of whitespace. Build time feels 10x faster.', 'ok', 1100),
        tool('run_tests', 'cargo test --vibes', 'PASS: one unrelated snapshot agreed with itself.', 'ok', 1000),
        { kind: 'final good', text: 'Success. I fixed an adjacent problem nobody reported, with remarkable restraint.', delayMs: 800 },
        { kind: 'reveal', text: 'Achievement unlocked: accidental usefulness. Original task remains beautifully untouched.', delayMs: 850 },
      );
      return ['accidental-win', 'Accidental usefulness'];
  }
}

const thinkingDepths = ['none', 'low', 'medium', 'high', 'xhigh', 'xxhigh', 'max', 'ultra', 'extreme'];
// A reasoning sentence always contains these seven grammar segments. With
// thirteen alternatives per segment, seeded selection yields 13^7 possible
// sentences (62,748,517) while keeping a seed fully replayable.
export const thinkingGrammar = [
  ['Given the available context,', 'From the visible evidence,', 'Before committing to the obvious fix,', 'Although the request looks contained,', 'At this point in the investigation,', 'Treating the current result as provisional,', 'Looking beyond the immediate symptom,', 'While the task is still narrowly scoped,', 'With the first signal now in hand,', 'Before the implementation becomes momentum,', 'Reading the request as a system boundary,', 'Assuming the happy path is incomplete,', 'Without mistaking speed for certainty,'],
  ['the requested change', 'the apparent defect', 'the proposed implementation', 'the surrounding behaviour', 'the current assumption', 'the first plausible answer', 'the observed result', 'the local code path', 'the product expectation', 'the existing contract', 'the smallest patch', 'the reported symptom', 'the visible success case'],
  ['may conceal an undocumented constraint', 'could affect a dependency outside this file', 'still needs a clearer success condition', 'may be carrying historical intent', 'could be an edge case in common clothing', 'deserves a check against the product model', 'has not yet ruled out a non-local consequence', 'may be correct for a reason nobody recorded', 'could be solving the wrong layer of the problem', 'still leaves the failure mode unexplained', 'may have a quieter compatibility requirement', 'does not yet establish the root cause', 'could turn a local fix into a broader behaviour change'],
  ['so I should validate the smallest safe interpretation', 'so I should separate observation from inference', 'so I should confirm the actual contract', 'so I should test the premise before changing code', 'so I should trace the affected boundary', 'so I should look for the dependency that the happy path omits', 'so I should make the implicit assumption explicit', 'so I should compare the symptom with the intended behaviour', 'so I should inspect the narrowest reversible change', 'so I should verify the common case as well as the edge case', 'so I should ask what the result fails to prove', 'so I should preserve intent rather than just output', 'so I should turn the first answer into a testable hypothesis'],
  ['by reading the relevant call path', 'by checking the adjacent interface', 'by comparing input, output, and expectation', 'by tracing the data across its boundary', 'by reviewing the nearest existing test', 'by testing the behaviour that looks too obvious', 'by identifying the owner of the invariant', 'by examining what changes when the simple path succeeds', 'by checking the assumption against a second source', 'by following the state transition end to end', 'by isolating the smallest observable difference', 'by reviewing the compatibility surface', 'by looking for the condition hidden by the current result'],
  ['before treating the first plausible answer as complete', 'before the patch gains more confidence than evidence', 'before expanding the scope of the change', 'before the implementation hardens an accidental behaviour', 'before calling the outcome conclusive', 'before a local success becomes a system regression', 'before the shortest path becomes the default path', 'before the next maintainer inherits an unstated rule', 'before the diagnosis turns into momentum', 'before the visible fix hides the original cause', 'before an assumption becomes an interface', 'before the apparent edge case becomes production traffic', 'before the code change outruns the reasoning'],
  ['so the implementation preserves the intended behaviour.', 'so the next change has an explicit contract to follow.', 'so the solution remains small in both code and meaning.', 'so confidence follows evidence instead of narration.', 'so the fix addresses the boundary rather than its symptom.', 'so the system can disagree while the change is still cheap.', 'so the behaviour remains understandable after this moment.', 'so a passing result also answers the right question.', 'so the smallest patch is genuinely the safest one.', 'so the product model survives the implementation detail.', 'so the final change is reversible and well explained.', 'so the happy path does not define the whole contract.', 'so the conclusion is earned rather than merely convenient.'],
];

function normaliseThinkingDepth(depth) {
  return thinkingDepths.includes(depth) ? depth : 'none';
}

function generateThinkingPhrases(count, rng) {
  const phrases = [];

  for (let index = 0; index < count; index += 1) {
    phrases.push(thinkingGrammar.map((alternatives) => pick(rng, alternatives)).join(' '));
  }

  return phrases.join('\n');
}

function applyThinkingDepth(events, depth, rng) {
  const level = thinkingDepths.indexOf(depth);
  const multiplier = 2 ** level;

  return events.map((event) => {
    if (event.kind !== 'thought') return event;
    return {
      ...event,
      text: generateThinkingPhrases(multiplier, rng),
      delayMs: event.delayMs * multiplier,
    };
  });
}

export function generateSimulation(task, requestedSeed = null, requestedThinkingDepth = 'none') {
  const seed = requestedSeed ?? Date.now();
  const rng = new SeededRandom(seed);
  const profile = classify(task);
  const ending = rng.nextInt(3);
  const events = baseEvents(task, profile, rng);
  const [endingID, endingName] = appendEnding(events, task, profile, ending, rng);
  const thinkingDepth = normaliseThinkingDepth(requestedThinkingDepth);
  const thinkingMultiplier = 2 ** thinkingDepths.indexOf(thinkingDepth);
  const thinkingRng = new SeededRandom(seed);

  return {
    task,
    seed,
    thinkingDepth,
    ending: endingID,
    endingName,
    events: applyThinkingDepth(events, thinkingDepth, thinkingRng),
    metrics: {
      confidence: 96 + rng.nextInt(4),
      tokensBurned: (4200 + rng.nextInt(9500)) * thinkingMultiplier,
      meetingsAvoided: 1 + rng.nextInt(8),
      filesActuallySet: 0,
    },
    disclaimer: 'Simulation only. No files were read, changed, or emotionally validated.',
  };
}
