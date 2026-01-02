# Project Constitution

constitution-mode: lite

## 1. Purpose and Scope

This repository contains tankismus, a Go game built with Ebiten and an ECS-style architecture. It is a sandbox for experimenting with gameplay, systems, and architecture while keeping the code readable and educational.

This constitution describes how we want to build and evolve the project:

- How we structure increments, designs, and implementation plans.
- How we document architectural decisions when they matter.
- How we approach testing, CI, and observability in a lightweight but deliberate way.


## 2. Implementation & Doc Layout

### Increment artifacts

- **Location**: `docs/increments/<slug>/`
- **Files**:
  - `increment.md` — product-level WHAT for a specific increment.
  - `design.md` — technical design (HOW) for that increment.
  - `implement.md` — implementation plan for that increment.

Each increment lives in its own folder and should be small and demonstrable.

### ADR artifacts

- **Location**: `docs/adr/`
- **Filename pattern**: `ADR-YYYY-MM-DD-<slug>.md`
- **Usage**:
  - Capture significant, long-lived architectural decisions (for example, major changes to ECS layout, rendering model, or persistence strategy).
  - Keep them short and focused on context, decision, and consequences.

In lite mode, ADRs are used sparingly, only when a decision will affect the project for a long time or is likely to be revisited.

### Improve artifacts

- **Location**: `docs/improve/`
- **Filename pattern**: `YYYY-MM-DD-improve-<slug>.md`
- **Usage**:
  - Reflect on system health, performance, or developer experience.
  - Outline non-trivial refactors or cleanups that cut across features.

Improve docs are optional and should be created only when they help coordinate or remember larger improvement efforts.

### Other docs

- High-level design and architecture notes live in `docs/` (for example, `docs/ecs.md` and any future `docs/ARCHITECTURE.md`).
- The root `README.md` remains the main entry point for understanding the game, its goals, and how to run it.

## 3. Design & Delivery Principles

### Small, safe steps

- Prefer many small, reversible changes over large, risky ones.
- Increments should be narrow in scope (for example, a single abstraction like a scene manager or a focused gameplay mechanic).
- Aim for changes that can be implemented, tested, and demonstrated quickly.

Examples:

- Introduce a new scene manager behind a simple interface before wiring complex gameplay into it.
- Add a new enemy behavior in a separate increment rather than bundling it with map changes and UI updates.

### Refactoring as everyday work

- Treat refactoring as part of normal development, not a separate phase.
- It is acceptable and expected to reshape code when adding features to keep it clear and maintainable.
- Avoid letting experimental code harden into long-term debt when small cleanups would keep things simple.

Examples:

- Extract a reusable movement or collision helper when you notice similar logic in multiple systems.
- Simplify a package layout when a feature grows beyond an initial quick spike.

### Pragmatic simplicity and clear boundaries

- Use package and type names that reflect the game domain (for example, scenes, systems, entities, lighting) rather than generic utility names.
- Keep abstractions as simple as possible for current needs; avoid over-generalizing or introducing frameworks prematurely.
- Favor clear boundaries between core game logic, ECS systems, and infrastructure (input, rendering, assets) where it helps understanding.

Examples:

- Keep scene management and timing separate from individual scene logic, so scenes can focus on behavior.
- Keep ECS systems focused on a single responsibility (for example, movement or rendering) instead of mixing unrelated concerns.

### No backward compatibility or hidden fallbacks

- Do not maintain long-lived runtime fallbacks or legacy code paths once an increment is merged; prefer deleting old behavior outright.
- When assumptions are violated (for example, required components are missing), fail fast via tests or explicit panics instead of silently degrading behavior.
- Short-lived experiment branches are allowed, but the main branch should not carry configuration switches that toggle between old and new models.

## 4. Testing, CI/CD, and Observability

### Testing (blackbox-oriented)

- Default to **blackbox testing**: test behavior through public APIs and observable outcomes, not internal implementation details.
- Write tests that exercise components the way game code or tools would use them (for example, driving a system through its exported functions or a game/scene interface).
- Prefer tests that describe expected behavior over tests that mirror internal fields or private helpers.

Expectations:

- Non-trivial logic (for example, timing behavior, scene selection, movement rules, collision handling) should be covered by automated tests.
- Use Go's standard testing tools (`go test ./...`) as the baseline.
- When useful, a small number of higher-level tests that run core flows (for example, stepping a game/scene through several updates) are encouraged, as long as they remain fast and reliable.

### CI/CD

- Aim for a simple, fast CI setup that at minimum:
  - Builds the project.
  - Runs `go test ./...`.
  - Optionally runs linters (`go vet`, static analysis) when they provide value.
- Keep the pipeline straightforward; avoid heavy release processes.
- Local development should remain easy: `go run` for demos or main binaries should work without extra tooling.

### Observability

- Rely primarily on clear, structured logs and visible in-game behavior for observability.
- Make it easy to understand core runtime behavior such as:
  - Which scene is active.
  - Basic timing information (for example, elapsed time, update rates) when debugging.
  - Key game state transitions (for example, start/end of a run).
- For debugging, simple on-screen debug overlays or log statements are acceptable; full observability stacks are not expected in lite mode.

## 5. ADR and Improve Usage

### ADRs

- Use ADRs in `docs/adr/` when making decisions that:
  - Change or introduce major architectural patterns (for example, a new networking model or a major ECS restructuring).
  - Are expected to be long-lived and important to remember.
- Keep ADRs concise: context, decision, alternatives considered, and consequences.

### Improve docs

- Use Improve docs in `docs/improve/` when:
  - You want to record observations about performance, complexity, or developer experience.
  - You are planning a significant refactor or series of refactors that span multiple increments.
- Focus Improve docs on actionable ideas and desired direction, not on exhaustive audits.

In lite mode, both ADRs and Improve docs are tools to use when they clearly help; day-to-day work should remain lightweight and focused on building and refining the game through small, testable increments.