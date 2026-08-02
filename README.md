# glowing-agent

A deliberately unserious, local-only **agentic coder simulator**. It looks
busy, uses imaginary tools, and never reads or changes your files.

## TUI

Run the full-screen Bubble Tea workbench in an interactive terminal:

```powershell
go run .
```

Enter a task, configure the simulation, then watch the agent find the root
cause of civilisation instead of your bug. The workbench includes a multi-line
task editor, preset selector, reproducible seed, thinking depth, replay speed,
scrolling log, and a responsive session-details sidebar. On narrower terminals,
the sidebar settings fold back into the main workbench.

| Key | Action |
| --- | --- |
| `Tab` / `Shift+Tab` | Move between task and settings |
| `←` / `→` | Change preset, thinking depth, or speed |
| `Enter` | Start a simulation |
| `Shift+Enter` | Insert a new task line |
| `Space` | Pause or resume playback |
| `r` | Replay the current incident |
| `n` | Return to task editing |
| `PgUp` / `PgDn` | Scroll the event log |
| `q` / `Ctrl+C` | Quit (`q` works outside text fields) |

The TUI needs at least a 60 × 24 terminal. It uses the terminal's alternate
screen and restores the original screen when it exits.

## JSON automation

For scripts and CI, `--json` bypasses the TUI and writes only one JSON result:

```powershell
go run . --json --preset button --seed 42 --thinking-depth high
"Add AI to the dashboard" | go run . --json --seed 7
```

`--preset`, `--seed`, and `--thinking-depth` are JSON-mode options; use the
interactive settings in normal TUI mode. Run `go run . --help` for details.

Build a native binary for the current operating system with `go build .`, or
install it with `go install .`. The same source builds on Windows, macOS, and
Linux.

## Browser demo

The static browser version remains available as a separate demo and is deployed
to GitHub Pages when changes reach `main`. In the repository, enable
**Settings → Pages → Source: GitHub Actions** once; it will be available at
<https://mcas-996.github.io/glowing-agent/>.

## Design inspiration

The interface is visually inspired by Charmbracelet's
[Crush](https://github.com/charmbracelet/crush). Its terminal workbench layout
is an independent implementation: no Crush source code, logo, wordmark,
screenshots, or other brand assets are included in this project.

## Development

```powershell
go test ./...
node --test
```
