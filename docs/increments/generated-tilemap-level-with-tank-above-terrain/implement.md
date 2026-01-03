# Implement: Generated Tilemap Level With Tank Above Terrain

## 0. Context

- Goal: Show a generated tilemap level in both a dedicated map demo and the main run scene, with the tank clearly rendered above the terrain, while preserving existing movement and firing behavior (see `increment.md`).
- Design: Scenes own a generated `pkg/map.Map`, compose it into a single tilemap image, represent that image as an ECS entity with render order/z index, and extend the render system to draw sprite-bearing entities in z order (see `design.md`).
- Constraints: Respect `constitution-mode: lite` and existing architecture (Ebiten usage confined to game layer and scene interface; `pkg` packages remain Ebiten-free, see `CONSTITUTION.md`).
- Non-goals: No terrain-dependent movement, collisions, or enemy/level progression changes in this increment.

Status: Not started  
Next step: Step 1 – Add terrain query helpers on top of `pkg/map`

## 1. Workstreams

- Workstream A – Map and terrain query abstraction (`pkg/map` + thin adapters).
- Workstream B – Rendering and z-order support (components, render system, tilemap composition).
- Workstream C – Run scene integration (main gameplay scene uses generated tilemap entity).
- Workstream D – Map demo scene (focused scene and entry that showcases map + tank above it).
- Workstream E – Tests and observability (tests across layers, minimal logging/overlays).

## 2. Steps

- [ ] Step 1: Add terrain query helpers on top of `pkg/map`
- [ ] Step 2: Introduce render order / z data in components
- [ ] Step 3: Make `RenderSystem` z-aware while preserving current behavior
- [ ] Step 4: Compose tilemap image and register it as a sprite
- [ ] Step 5: Integrate map and tilemap entity into the run scene
- [ ] Step 6: Add a dedicated map demo scene using the same pattern
- [ ] Step 7: Add logging and minor observability hooks for maps
- [ ] Step 8: Add end-to-end verification hooks for scenes with tilemaps

### Step 1: Add terrain query helpers on top of `pkg/map`

- Workstream: A
- Based on Design: §2 (Terrain query abstraction), §5 (Map and Terrain Abstraction, World-to-tile mapping)
- Files: `pkg/map/map.go`, `pkg/map/map_test.go`
- TDD Cycle:
  - Red – Failing test first:
    - Extend `pkg/map/map_test.go` to construct a small grass map (for example, 3×2) using `NewGrassMap` and assert that:
      - A helper like `TileAt(x, y)` returns the expected tile ID strings for in-bounds tile coordinates.
      - A helper like `TileAtWorld(worldX, worldY)` maps world coordinates at tile centers to the correct tile coordinates or IDs, based on a chosen tile size and origin.
      - Out-of-bounds world coordinates (negative or beyond map extents) yield a well-defined result (for example, an error or a sentinel value) as described in `design.md`.
    - Run `go test ./pkg/map -run Test` and confirm the new tests fail.
  - Green – Make the test(s) pass:
    - Implement minimal helper functions on/around `Map` in `pkg/map/map.go` that:
      - Compute tile indices from world coordinates given the fixed tile size and origin assumptions.
      - Safely index into `Tiles` and return the corresponding tile ID or a terrain classification.
      - Return a defined result for out-of-bounds queries.
    - Re-run `go test ./pkg/map` and ensure all tests pass.
  - Refactor – Clean up with tests green:
    - Refine helper names and signatures for clarity (for example, `TileAt`, `TileAtWorld`, `TerrainAtWorld`), keeping the package Ebiten-free.
    - Remove duplication and document coordinate conventions in comments.
- CI / Checks:
  - `go test ./pkg/map`
  - `go test ./...` (sanity check)

### Step 2: Introduce render order / z data in components

- Workstream: B
- Based on Design: §2 (Existing entity rendering with z ordering), §5 (Render Order / Z Index)
- Files: `game/components/components.go`, `game/components/components_test.go`
- TDD Cycle:
  - Red – Failing test first:
    - In `game/components/components_test.go`, add tests that:
      - Create an entity with both a sprite component and a new render-order/z representation (for example, a dedicated component or extended sprite struct) and verify that the ECS world can store and retrieve this information correctly.
      - Verify that default render order (when not explicitly set) matches the current behavior (for example, existing entities render as they do today).
    - Run `go test ./game/components -run Test` and observe failures.
  - Green – Make the test(s) pass:
    - Extend `game/components/components.go` to introduce a minimal render-order/z abstraction, such as:
      - A new component type that holds a small integer or categorical layer value, or
      - An additional field on the sprite-related data, consistent with `design.md`.
    - Ensure the new type is wired into the ECS component type IDs and that tests can attach/detach it from entities.
    - Re-run `go test ./game/components` and ensure all tests pass.
  - Refactor – Clean up with tests green:
    - Normalize naming (for example, `RenderLayer` or `ZOrder`) and default values (ground vs actors) to match the design.
    - Update any helper constructors or fixtures to set sensible defaults.
- CI / Checks:
  - `go test ./game/components`
  - `go test ./...`

### Step 3: Make `RenderSystem` z-aware while preserving current behavior

- Workstream: B
- Based on Design: §2 (Existing entity rendering with z ordering), §4 (Architecture and Boundaries), §5 (Render Order / Z Index)
- Files: `game/systems/render.go`, `game/systems/render_test.go`
- TDD Cycle:
  - Red – Failing test first:
    - In `game/systems/render_test.go`, add tests that:
      - Given a world with two sprite-bearing entities, one representing the tilemap (low z) and one representing the tank (higher z), verify that the render system orders them such that the tilemap draw call logically precedes the tank draw call. This can be asserted via a test double or inspection of ordering logic rather than pixel output.
      - Confirm that when no explicit z is set, the behavior matches the current rendering order (to avoid regressions).
    - Run `go test ./game/systems -run TestRender` and observe the new tests fail.
  - Green – Make the test(s) pass:
    - Modify `RenderSystem` in `game/systems/render.go` to:
      - Collect sprite-bearing entities along with their render order/z value.
      - Sort or otherwise iterate over entities in increasing z before calling `screen.DrawImage`.
      - Default render order to the current behavior when the z component is absent.
    - Ensure tank and other entities still draw correctly while allowing the tilemap entity to be drawn first when present.
    - Re-run `go test ./game/systems` and confirm all tests pass.
  - Refactor – Clean up with tests green:
    - Extract any small helpers for building the drawable list or sorting by z to keep `RenderSystem` readable.
    - Avoid leaking render-order details into unrelated systems.
- CI / Checks:
  - `go test ./game/systems`
  - `go test ./...`

### Step 4: Compose tilemap image and register it as a sprite

- Workstream: B
- Based on Design: §2 (Tilemap rendering as an entity), §4 (Architecture and Boundaries)
- Files: new helper in `game` or `game/assets` (for example, `game/assets/tilemap.go`), `game/assets/assets.go`, tests in `game/assets/assets_test.go` or a new test file for tilemap composition
- TDD Cycle:
  - Red – Failing test first:
    - Add tests that:
      - Given a small `pkg/map.Map` with known tile IDs, the composition helper:
        - Looks up the corresponding tile sprites via the assets registry.
        - Produces an Ebiten image of expected dimensions (width = map.Width * tileWidth, height = map.Height * tileHeight).
        - Registers the composed image under a known sprite ID (for example, `"tilemap_ground"`).
    - Use simple fake or stub sprites to avoid depending on real assets at test time.
    - Run `go test ./game/assets -run Test` and observe failures.
  - Green – Make the test(s) pass:
    - Implement the tilemap composition helper to:
      - Iterate over map tiles, fetch each tile image via `assets.GetSprite`, and draw into a new Ebiten image at the correct position.
      - Register the resulting image in the assets registry under a deterministic ID.
    - Ensure composition is invoked in a way that can be reused by both run and demo scenes.
    - Re-run `go test ./game/assets` and confirm all tests pass.
  - Refactor – Clean up with tests green:
    - Factor out reusable pieces (for example, tile size constants, mapping from tile ID to sprite ID) to keep code simple and maintainable.
- CI / Checks:
  - `go test ./game/assets`
  - `go test ./...`

### Step 5: Integrate map and tilemap entity into the run scene

- Workstream: C
- Based on Design: §2 (Map ownership per scene, tilemap entity), §4 (Architecture and Boundaries)
- Files: `game/scenes/run/run.go`, `game/scenes/run/run_test.go`
- TDD Cycle:
  - Red – Failing test first:
    - Extend `game/scenes/run/run_test.go` to assert that, after constructing a new run scene:
      - A `pkg/map.Map` instance exists with non-zero width, height, and non-empty tiles.
      - An ECS entity exists representing the tilemap, with:
        - A sprite component referring to the composed tilemap sprite ID.
        - A render-order/z value indicating it is in the ground layer.
      - The existing player entity remains present and properly configured (Transform, Velocity, ControlIntent, MovementParams, Sprite).
    - Run `go test ./game/scenes -run TestRunScene` and observe failures.
  - Green – Make the test(s) pass:
    - Update `Scene` in `run.go` to:
      - Create and store a `pkg/map.Map` using the terrain query helpers and chosen dimensions/seed.
      - Call the tilemap composition helper to create/register the combined tilemap image.
      - Create the tilemap ECS entity with Transform (positioning the map in world coordinates), Sprite (composed tilemap image ID), and ground-layer render order/z component.
      - Leave the existing player entity creation and systems wiring intact.
    - Re-run `go test ./game/scenes` and ensure tests pass.
  - Refactor – Clean up with tests green:
    - Extract any map-creation or tilemap-entity wiring code into small helpers inside the run scene for readability and reuse.
- CI / Checks:
  - `go test ./game/scenes`
  - `go test ./...`

### Step 6: Add a dedicated map demo scene using the same pattern

- Workstream: D
- Based on Design: §2 (Map demo scene), §4 (Architecture and Boundaries)
- Files: new scene file under `game/scenes` (for example, `game/scenes/mapdemo/mapdemo.go` and `mapdemo_test.go`), `cmd/scenes-demo/main.go`
- TDD Cycle:
  - Red – Failing test first:
    - Create `mapdemo_test.go` to assert that the demo scene:
      - On construction, creates a `pkg/map.Map` and corresponding tilemap entity with ground-layer z.
      - Creates a player tank entity similar to the run scene (Transform, Velocity, ControlIntent, MovementParams, Sprite).
      - On repeated constructions (or restarts), produces generated maps (for example, differing seeds or tile patterns) rather than a fixed background, in line with acceptance criteria.
    - Run `go test ./game/scenes -run TestMapDemo` and observe failures.
  - Green – Make the test(s) pass:
    - Implement the map demo scene to:
      - Reuse the same map creation, tilemap composition, and tilemap entity wiring as the run scene.
      - Wire input and ECS systems so the player can move the tank over the tilemap.
      - Ensure the scene can be started via `cmd/scenes-demo/main.go` or equivalent entry point.
    - Re-run `go test ./game/scenes` and confirm all tests pass.
  - Refactor – Clean up with tests green:
    - Extract shared helpers (e.g., for map creation and tilemap entity setup) used by both run and demo scenes to avoid duplication.
    - Keep scene-specific behavior (e.g., transitions, HUD) separate.
- CI / Checks:
  - `go test ./game/scenes`
  - `go test ./cmd/scenes-demo`
  - `go test ./...`

### Step 7: Add logging and minor observability hooks for maps

- Workstream: E
- Based on Design: §8 (Observability and Operations)
- Files: `game/scenes/run/run.go`, `game/scenes/mapdemo/mapdemo.go`, optional debug overlay helpers
- TDD Cycle:
  - Red – Failing test first:
    - Where practical, add small tests or checks that:
      - Format log messages correctly when a map is initialized (seed, width, height).
      - Optionally, when a debug mode flag is enabled, expose data such as tile coordinates or terrain kind under the player.
    - Run relevant tests and confirm failures for new expectations.
  - Green – Make the test(s) pass:
    - Add log statements in scene initialization paths that record map seed and dimensions.
    - Optionally, add a simple debug overlay mechanism that can display map or terrain info when enabled, keeping it off by default.
    - Re-run tests and ensure they pass.
  - Refactor – Clean up with tests green:
    - Ensure logs are concise and do not flood output.
    - Keep any debug overlays lightweight and easy to disable.
- CI / Checks:
  - `go test ./game/scenes`
  - `go test ./...`

### Step 8: Add end-to-end verification hooks for scenes with tilemaps

- Workstream: E
- Based on Design: §6 (Testing and Safety Net), increment Success Criteria
- Files: existing higher-level tests (if any) or new tests that exercise game/scene flows; documentation of manual checks
- TDD Cycle:
  - Red – Failing test first:
    - Add or extend a higher-level test that:
      - Constructs the run scene (and demo scene) via their public constructors and steps through a minimal update/draw cycle, asserting that:
        - No panics occur.
        - The ECS world contains both a tilemap entity with ground-layer z and a player entity with expected components.
        - Map tiles are non-empty.
    - Run `go test ./...` and confirm failures in the new tests.
  - Green – Make the test(s) pass:
    - Adjust scene initialization or wiring only as needed to satisfy these tests, without changing the design scope.
    - Optionally, update documentation or a simple manual checklist to include visual verification steps (map visible, tank above map, maps differ across runs).
    - Re-run `go test ./...` and ensure everything passes.
  - Refactor – Clean up with tests green:
    - Simplify test setup helpers and ensure the tests run quickly enough for normal use.
- CI / Checks:
  - `go test ./...`

## 3. Rollout & Validation Notes

- Suggested PR grouping:
  - PR 1: Steps 1–3 (terrain query helpers, render-order/z component, z-aware render system + tests).
  - PR 2: Steps 4–5 (tilemap composition helper and run scene integration + tests).
  - PR 3: Steps 6–8 (map demo scene, logging/observability, end-to-end verification hooks + tests).

- Suggested validation checkpoints:
  - After PR 1: All unit tests pass; rendering still works as before with default z, and simple inspection shows no regressions.
  - After PR 2: Running the main game shows a visible generated tilemap under the tank; basic movement and firing still work.
  - After PR 3: Running the map demo shows generated tilemaps and tank above them; restarting demonstrates generation; logs show seed and dimensions as expected.

- Manual validation hints:
  - Start the main game and confirm:
    - A tiled level is visible as ground.
    - The tank is clearly above the tiles during movement and firing.
    - Restarting produces different map patterns (or at least clearly generated layouts).
  - Start the map demo and confirm similar behavior in a focused environment.
