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
const thinkingPhrases = [
  'The request appears small, but its assumptions may extend beyond the visible change.',
  'Before acting, I should distinguish the stated requirement from the requirement it quietly implies.',
  'The surrounding code may be correct for reasons that are no longer documented.',
  'A quick fix is still a hypothesis until the system has had an opportunity to disagree.',
  'The most relevant constraint may be the one that has not yet been named.',
  'This behaviour deserves a second look from the perspective of the next maintainer.',
  'The evidence is encouraging, though it has not yet earned the right to be conclusive.',
  'A local change can carry a surprisingly non-local interpretation.',
  'The implementation path is clear enough to be suspicious.',
  'I should verify whether the apparent edge case is actually the common case in disguise.',
  'The task may be asking for a code change while revealing a boundary in the product model.',
  'It is worth separating what is observable from what merely feels architecturally significant.',
  'A stable solution needs to preserve the intent that led to the current behaviour.',
  'The safest next step is to make the implicit contract explicit in my reasoning.',
  'The available context supports a direction, not yet a conclusion.',
  'There may be a dependency here that only becomes visible when the simple path succeeds.',
  'The shortest implementation is not always the smallest conceptual change.',
  'I should account for the consequences that the happy path has politely omitted.',
  'This is probably straightforward, which makes it an excellent place to inspect the premise.',
  'The system is offering an answer; the remaining question is whether it is answering the right problem.',
];

function normaliseThinkingDepth(depth) {
  return thinkingDepths.includes(depth) ? depth : 'none';
}

function generateThinkingPhrases(count, rng) {
  const phrases = [];
  let previousIndex = -1;

  for (let index = 0; index < count; index += 1) {
    let phraseIndex = rng.nextInt(thinkingPhrases.length - (previousIndex === -1 ? 0 : 1));
    if (previousIndex !== -1 && phraseIndex >= previousIndex) phraseIndex += 1;
    phrases.push(thinkingPhrases[phraseIndex]);
    previousIndex = phraseIndex;
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

export function generateSimulation(task, requestedSeed = null, requestedThinkingDepth = 'none', requestedThinkingTime = Date.now()) {
  const seed = requestedSeed ?? Date.now();
  const rng = new SeededRandom(seed);
  const profile = classify(task);
  const ending = rng.nextInt(3);
  const events = baseEvents(task, profile, rng);
  const [endingID, endingName] = appendEnding(events, task, profile, ending, rng);
  const thinkingDepth = normaliseThinkingDepth(requestedThinkingDepth);
  const thinkingMultiplier = 2 ** thinkingDepths.indexOf(thinkingDepth);
  const thinkingRng = new SeededRandom(BigInt(seed) ^ BigInt(requestedThinkingTime));

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
