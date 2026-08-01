# glowing-agent

A deliberately unserious, local-only **agentic coder simulator**. It looks busy,
uses imaginary tools, and never reads or changes your files.

## Run

```powershell
go run .
```

Open [http://localhost:8080](http://localhost:8080). Choose a preset or type a
task, then replay the same incident with the displayed seed whenever you need
to reproduce the confidence.

## GitHub Pages

The static version runs entirely in the browser and is deployed automatically
when changes reach `main`. In the repository, enable **Settings → Pages →
Source: GitHub Actions** once; the site will then be available at
<https://mcas-996.github.io/glowing-agent/>.

## Development

```powershell
go test ./...
```

The project deliberately has no third-party dependencies, external model calls,
or shell/file-system tool access.
