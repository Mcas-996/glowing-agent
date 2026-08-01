const $ = (selector) => document.querySelector(selector);
const terminal = $('#terminal');
const run = $('#run');
const replay = $('#replay');
const task = $('#task');
const seed = $('#seed');
const speed = $('#speed');
let simulation = null;
let activePreset = null;

const presets = [
  ['typo', 'One tiny typo', 'Fix the typo in the welcome message'],
  ['button', 'Ship a button', 'Add a blue Save button to the settings page'],
  ['speed', 'Make it faster', 'Make the build ten times faster'],
  ['ai', 'Add AI', 'Add AI to the dashboard'],
];

for (const [id, label, value] of presets) {
  const button = document.createElement('button');
  button.className = 'preset'; button.type = 'button'; button.textContent = label;
  button.addEventListener('click', () => { task.value = value; activePreset = id; markPreset(); });
  button.dataset.id = id; $('#presets').append(button);
}
task.addEventListener('input', () => { activePreset = null; markPreset(); });
function markPreset() { document.querySelectorAll('.preset').forEach(b => b.classList.toggle('selected', b.dataset.id === activePreset)); }

run.addEventListener('click', async () => {
  const text = task.value.trim();
  if (!text) { task.focus(); return; }
  const request = { task: text };
  if (activePreset) request.presetId = activePreset;
  if (seed.value.trim()) request.seed = Number(seed.value.trim());
  if (!Number.isSafeInteger(request.seed) && seed.value.trim()) { seed.setCustomValidity('Use a whole-number seed.'); seed.reportValidity(); return; }
  seed.setCustomValidity(''); run.disabled = true; run.firstChild.textContent = 'Consulting the void ';
  try {
    const response = await fetch('/api/simulations', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(request) });
    const data = await response.json();
    if (!response.ok) throw new Error(data.error || 'The simulation unionised.');
    simulation = data; seed.value = data.seed; await play(data);
  } catch (error) { terminal.innerHTML = `<div class="event reveal">Simulation failed: ${escapeHTML(error.message)}</div>`; }
  finally { run.disabled = false; run.firstChild.textContent = 'Run simulation '; replay.disabled = !simulation; }
});

replay.addEventListener('click', () => simulation && play(simulation));

async function play(data) {
  replay.disabled = true; terminal.innerHTML = ''; $('#result').classList.add('hidden');
  for (const event of data.events) {
    appendEvent(event);
    await wait(Math.max(55, event.delayMs * Number(speed.value)));
  }
  $('#ending').textContent = data.endingName;
  $('#confidence').textContent = `${data.metrics.confidence}%`;
  $('#tokens').textContent = data.metrics.tokensBurned.toLocaleString();
  $('#files').textContent = data.metrics.filesActuallySet;
  $('#result').classList.remove('hidden'); replay.disabled = false;
}

function appendEvent(event) {
  const line = document.createElement('div');
  line.className = `event ${event.kind}`;
  if (event.tool) {
    const t = event.tool;
    line.classList.add(t.status);
    line.innerHTML = `<div><span class="tool-name">▣ ${escapeHTML(t.name)}</span> <span class="tool-input">${escapeHTML(t.input)}</span></div><div class="tool-output">${escapeHTML(t.output)}</div>`;
  } else { line.textContent = event.text; }
  terminal.append(line); terminal.scrollTop = terminal.scrollHeight;
}
function wait(ms) { return new Promise(resolve => setTimeout(resolve, ms)); }
function escapeHTML(value) { const node = document.createElement('div'); node.textContent = value; return node.innerHTML; }
