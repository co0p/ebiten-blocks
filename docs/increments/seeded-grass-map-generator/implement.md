# Implement: Seeded Grass Map Generator

## 0. Context

- Goal: Provide a developer-facing CLI (`cmd/genesis`) that generates a grass-only tile map as a JSON matrix from a seed, tile width, and tile height, using only `tileGrass1` and `tileGrass2`, deterministically.
- Non-goals: No runtime map loading or rendering, no additional terrain types or obstacles, no ECS/scene changes.
- Design approach: A pure-Go map engine package in `pkg` handles map representation, deterministic grass-only generation, and JSON (un)marshalling; `cmd/genesis` is a thin CLI wrapper that parses arguments, invokes the engine, and writes JSON.
- Constraints: `constitution-mode: lite` — keep steps small and pragmatic; respect layering rules from `CONSTITUTION.md` (no dependencies from `pkg` to `game` or `cmd`, no Ebiten in `pkg`).

References: `increment.md`, `design.md`, `CONSTITUTION.md`.

Status: Not started  
Next step: Step 1 – Introduce map type and JSON round-trip in pkg

## 1. Workstreams

- **Workstream A – Map Engine (pkg)**  
  Implement the deterministic grass-only map generator and JSON encoding/decoding in a new `pkg` package.

- **Workstream B – Genesis CLI (cmd/genesis)**  
  Wire a CLI that parses seed and dimensions, calls the map engine, and writes JSON to disk with clear errors.

- **Workstream C – Integration & Polish**  
  End-to-end verification of CLI behavior, plus minor refinements to logs and docs.

## 2. Steps

### Step 1: Introduce map type and JSON round-trip in pkg

- **Workstream:** A – Map Engine  
- **Based on Design:** Design §Contracts and Data (Map JSON Structure, JSON Schema Fragment)  
- **Files:** `pkg/map/map.go`, `pkg/map/map_test.go`

- **TDD Cycle:**  
  - **Red – Failing test first:**  
    - In `pkg/map/map_test.go`, add tests that:
      - Construct a `Map` value with explicit `width`, `height`, `seed`, and a small `tiles` matrix.  
      - Serialize it to JSON and then deserialize it back.  
      - Assert that `width`, `height`, `seed`, and `tiles` dimensions and contents round-trip unchanged.  
      - Assert that clearly invalid JSON (e.g., mismatched `width`/`height` vs. `tiles` dimensions) results in an error from the decode path.
  - **Green – Make the test(s) pass:**  
    - Introduce a `Map` struct in `pkg/map/map.go` that matches the design’s JSON schema (`width`, `height`, `seed`, `tiles [][]string`).  
    - Implement JSON (un)marshalling and basic validation so that:  
      - Round-tripping a valid `Map` preserves all fields.  
      - Obvious structural inconsistencies (e.g., row length mismatch) cause errors during decoding or an explicit validation step.
  - **Refactor – Clean up with tests green:**  
    - Extract internal validation helpers to keep the `Map` API simple and avoid duplication.  
    - Ensure naming and field visibility reflect intended use (e.g., exported type with unexported helpers).

- **CI / Checks:**  
  - Run `go test ./pkg/...`.

---

### Step 2: Implement deterministic grass-only generator in pkg

- **Workstream:** A – Map Engine  
- **Based on Design:** Design §Proposed Solution (Map Engine), §Contracts and Data (Determinism and Seed Handling), §Testing and Safety Net  
- **Files:** `pkg/map/map.go`, `pkg/map/map_test.go`

- **TDD Cycle:**  
  - **Red – Failing test first:**  
    - Add tests around a constructor-like function (e.g., `NewGrassMap(seed, width, height)`) that assert:  
      - The returned `Map` has `width` and `height` exactly as requested.  
      - `tiles` is non-nil with `height` rows and `width` columns per row.  
      - Every `tiles[row][col]` is either `"tileGrass1"` or `"tileGrass2"`.  
      - Two calls with the same seed and dimensions yield identical `tiles` matrices.  
      - Calls with different seeds (same dimensions) typically yield different `tiles` matrices (no strict uniqueness requirement, just a basic difference check).
  - **Green – Make the test(s) pass:**  
    - Implement `NewGrassMap` in `pkg/map/map.go` to:  
      - Accept a seed (use an agreed type per design) and dimensions.  
      - Initialize a local pseudo-random number generator from the seed.  
      - Fill `tiles` in a stable order (e.g., row-major) by selecting between `"tileGrass1"` and `"tileGrass2"` based on PRNG output.  
      - Populate `width`, `height`, `seed`, and `tiles` fields accordingly.  
    - Ensure that no global RNG is used; all randomness should be local to this function.
  - **Refactor – Clean up with tests green:**  
    - Extract small helpers for PRNG setup and tile choice if they improve readability.  
    - Keep the public surface focused (e.g., one main constructor for this increment).

- **CI / Checks:**  
  - Run `go test ./pkg/...`.

---

### Step 3: Validate structural invariants and error paths in pkg

- **Workstream:** A – Map Engine  
- **Based on Design:** Design §Contracts and Data (Constraints, Schema Enforcement), §Testing and Safety Net  
- **Files:** `pkg/map/map.go`, `pkg/map/map_test.go`

- **TDD Cycle:**  
  - **Red – Failing test first:**  
    - Extend `pkg/map/map_test.go` to cover invalid cases:  
      - Attempts to create or decode a `Map` with `width <= 0` or `height <= 0`.  
      - `tiles` slice length not equal to `height`.  
      - Any row in `tiles` whose length does not equal `width`.  
      - Any tile value not equal to `"tileGrass1"` or `"tileGrass2"`.  
    - Assert that such cases return errors and do not yield a valid `Map` instance.
  - **Green – Make the test(s) pass:**  
    - Implement or strengthen validation logic (e.g., a `Validate()` method or constructor checks) such that all the above invalid cases are rejected with clear errors.  
    - Integrate validation into JSON decoding and any constructors to ensure invariants are always enforced.
  - **Refactor – Clean up with tests green:**  
    - Consolidate validation into a single internal function or method to avoid code duplication.  
    - Ensure error messages are consistent and informative for callers.

- **CI / Checks:**  
  - Run `go test ./pkg/...`.

---

### Step 4: Implement core CLI function for genesis (testable API)

- **Workstream:** B – Genesis CLI  
- **Based on Design:** Design §Proposed Solution (CLI responsibilities), §Architecture and Boundaries  
- **Files:** `cmd/genesis/main.go`, `cmd/genesis/main_test.go` (or a dedicated test file alongside)

- **TDD Cycle:**  
  - **Red – Failing test first:**  
    - Define a testable function (e.g., `Run(seedStr, widthStr, heightStr, outputPath string) error`) in `cmd/genesis` and write tests that:  
      - With valid seed and positive integer dimensions, write a JSON file to a temporary path, then read it back and confirm:  
        - The file exists and parses.  
        - Dimensions and tile values satisfy the invariants already tested in `pkg/map`.  
      - With invalid dimensions (zero, negative, or non-integer input), return an error and do not create the output file.  
      - With an unusable output path (e.g., directory that does not exist or is unwritable), return an error and leave no partial JSON file.
  - **Green – Make the test(s) pass:**  
    - Implement `Run` to:  
      - Parse `seedStr`, `widthStr`, and `heightStr` into appropriate types, enforcing positivity on width/height.  
      - Call `pkg/map`’s grass-map constructor with the parsed seed and dimensions.  
      - Marshal the resulting map to JSON.  
      - Write the JSON to `outputPath` atomically (e.g., via temp file + rename or by ensuring error handling prevents partial files).  
      - Return descriptive errors for invalid arguments and I/O failures.
  - **Refactor – Clean up with tests green:**  
    - Factor out small helpers for argument parsing or file writing if that keeps `Run` focused.  
    - Ensure no CLI-specific concerns (like process exit codes) are embedded in `Run`; those belong in `main()`.

- **CI / Checks:**  
  - Run `go test ./cmd/... ./pkg/...`.

---

### Step 5: Wire CLI arguments in main() and basic logging

- **Workstream:** B – Genesis CLI  
- **Based on Design:** Design §CI/CD and Rollout, §Observability and Operations  
- **Files:** `cmd/genesis/main.go`

- **TDD Cycle:**  
  - **Red – Failing test first:**  
    - In `cmd/genesis/main_test.go`, add tests (where practical) that call a wrapper around `main` logic (e.g., a function taking an argument slice) to assert:  
      - CLI arguments are correctly forwarded to `Run`.  
      - Successful invocations lead to a zero exit-like result (e.g., no error).  
      - Failed invocations (invalid args or I/O) map to non-zero behavior (e.g., error returned by the wrapper).  
    - At minimum, ensure that logging or printed messages include seed, dimensions, and output path.
  - **Green – Make the test(s) pass:**  
    - Implement `main()` to:  
      - Parse `os.Args` into seed, width, height, and output path (with a basic `--help` usage if desired).  
      - Log/print the seed, width, height, and output path at start.  
      - Delegate to `Run` and convert its error into an appropriate exit code.  
      - Print a brief success message on completion.
  - **Refactor – Clean up with tests green:**  
    - Keep `main()` minimal by delegating as much logic as possible to testable helpers.  
    - Ensure logs/messages are concise and consistent.

- **CI / Checks:**  
  - Run `go test ./cmd/... ./pkg/...`.  
  - Manual smoke check: `go run ./cmd/genesis --help` and a sample invocation with a small map.

---

### Step 6: End-to-end CLI integration test and docs touch-up

- **Workstream:** C – Integration & Polish  
- **Based on Design:** Design §Testing and Safety Net, §Observability and Operations, §Follow-up Work (foundation for later loading)  
- **Files:** `cmd/genesis/main_test.go`, optionally `docs/increments/seeded-grass-map-generator/increment.md`

- **TDD Cycle:**  
  - **Red – Failing test first:**  
    - Add an integration-style test that uses the same core `Run` function (or a small harness built on it) to:  
      - Generate a map with a fixed seed and dimensions into a temporary file.  
      - Parse the resulting JSON back using `pkg/map`’s decode logic.  
      - Assert that dimensions and tile values match expectations (same invariants as unit tests) and that re-running with the same seed/dimensions produces identical JSON.
  - **Green – Make the test(s) pass:**  
    - Address any mismatches between CLI behavior and the map engine contract (e.g., missing `seed` field or incorrect field names).  
    - Adjust error handling or parameter parsing if the integration test exposes edge cases.  
  - **Refactor – Clean up with tests green:**  
    - Optionally add a short note or example command to `increment.md` describing how to run the generator in practice, staying within the existing increment scope (no game integration).  
    - Ensure test names and structure clearly communicate the end-to-end contract.

- **CI / Checks:**  
  - Final run of `go test ./...`.  
  - Optional manual verification: generate a few maps with different seeds and visually inspect JSON differences.

## 3. Rollout & Validation Notes

- **Suggested PR Grouping:**  
  - **PR 1:** Steps 1–3 (introduce `pkg/map` with JSON round-trip, deterministic grass-only generation, and validation).  
  - **PR 2:** Steps 4–5 (implement `cmd/genesis` CLI core function and `main()` wiring with basic logging).  
  - **PR 3:** Step 6 (end-to-end integration test and any doc touch-ups).

- **Validation Checkpoints:**  
  - **After Step 3:**  
    - `go test ./pkg/...` passes.  
    - Unit tests confirm deterministic, valid grass-only map generation and strict structural invariants.  
  - **After Step 5:**  
    - `go test ./cmd/... ./pkg/...` passes.  
    - Manual CLI invocations generate JSON files that parse and exhibit expected dimensions and tile values.  
  - **After Step 6:**  
    - `go test ./...` passes.  
    - Integration tests confirm CLI and map engine alignment; repeated runs with same seed/dimensions yield identical outputs.

- **Rollback Considerations:**  
  - Since this increment is additive and does not modify existing game binaries, rollback is limited to reverting the new `pkg/map` package and `cmd/genesis` wiring if needed.  
  - No data migrations or persistent schema changes are introduced by this plan.
