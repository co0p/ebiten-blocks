# Design: Generated Tilemap Level With Tank Above Terrain

## 1. Context and Problem

This increment aims to make a generated tilemap level and clear tank-over-terrain layering visible and playable in both a dedicated map demo and the main game experience, while preserving existing movement and firing behavior.

Today, the main gameplay scene uses a flat background color; the player tank and other entities are rendered directly onto the screen without any visible level structure. A seeded, Ebiten-free map module already exists and can generate a grass-only tilemap, but it is not yet integrated into scenes or rendering. Players therefore do not see terrain and cannot tell that a map is being generated.

In addition, the map data structure needs to support querying the terrain at a given position so that future increments can adapt behavior (movement, collisions, AI) based on the terrain under the tank or other entities.

This design describes how to:

- Integrate generated tilemaps into the run scene and a map demo scene.
- Render the tilemap as a ground layer beneath the existing entity rendering.
- Provide a terrain query abstraction on top of the map data,

while respecting:

- The project constitution (lite mode, small safe steps, clear boundaries between game logic, ECS systems, and infrastructure).
- The existing architecture (three layers: binaries, game packages, reusable engine-style packages, with Ebiten usage constrained to specific layers).

References:

- Project constitution: `CONSTITUTION.md`.
- Architecture overview and diagrams: `ARCHITECTURE.md`.
- Increment definition: `increment.md` in this folder.
- Seeded grass map generator design: grass map design document under `docs/increments`.

## 2. Proposed Solution (Technical Overview)

At a high level, the solution introduces generated tilemaps as scene-owned data, adds a tilemap rendering path, and exposes terrain queries without changing existing ECS behavior.

Key ideas:

- **Map ownership per scene**: Both the map demo scene and the main run scene own an instance of the existing map data structure, constructed at scene initialization using a seed and chosen dimensions that fit the logical game area.
- **Terrain query abstraction**: The map module (or a thin adapter around it) exposes functions to:
  - Retrieve the tile identifier at given tile coordinates.
  - Map world coordinates (the space in which entities move) to tile coordinates and thus to terrain information.
- **Tilemap rendering as an entity**: A tilemap rendering helper in the game layer composites all tiles into a single tilemap image (using tile sprites from the assets registry), registers that image like other sprites, and represents the tilemap as an ECS entity with a transform, sprite reference, and a render order/z index indicating it belongs to the ground layer.
- **Existing entity rendering with z ordering**: The current entity rendering system is extended to respect a render order/z index when drawing sprite-bearing entities. The tilemap entity uses the lowest z so it is drawn first, and the tank and other dynamic entities use higher z values so they are drawn above the tilemap image, preserving the visual layering without special-casing the tilemap.
- **Map demo scene**: A dedicated scene focused on showcasing map generation uses the same pattern (scene-owned map, tilemap image composition, tilemap entity, z-aware entity rendering) so that players can see a generated level plus a tank moving above it in isolation.

The result is a clear, testable layering: map generation and terrain queries in the map layer; rendering of tiles as ground; ECS-based entity rendering on top.

## 3. Scope and Non-Scope (Technical)

**In scope**

- Using the existing map data structure to generate a tilemap in the map demo scene and the main run scene.
- Adding a terrain query abstraction on top of the map data, including:
  - Tile lookup by tile coordinates.
  - Terrain lookup by world coordinates.
- Rendering the tilemap as a ground layer beneath existing entity/tank rendering.
- Ensuring that the presence of the tilemap does not break existing movement and firing behavior.

**Out of scope**

- Changing movement, physics, or AI behavior based on terrain (these will consume the terrain query abstraction in later increments).
- Implementing collisions or destructible obstacles based on the tilemap.
- Major refactors of ECS, scene management, or overall rendering architecture.
- Introducing new external dependencies, APIs, or persistence mechanisms.

## 4. Architecture and Boundaries

This design stays within the existing layered architecture:

- **Binaries**: Continue to own user entry points (for example, main game and scenes demo). They construct games/scenes but do not know about map internals.
- **Game layer**:
  - Scenes manage game state over time (including ECS worlds and now map instances).
  - Rendering helpers and systems use Ebiten to draw both the tilemap and entities.
- **Reusable engine-style packages**:
  - The map module provides tilemap data and terrain queries without depending on Ebiten or game-specific concerns.
  - ECS and input modules remain unchanged.

Ebiten usage remains localized to the game layer and scene interface, in line with the architecture document. The map module stays Ebiten-free.

A typical frame in the run scene after this change:

1. The scene receives an update and, if needed, interacts with the map for terrain queries (e.g., in future increments).
2. For drawing, the scene:
  - Ensures the tilemap entity (backed by the composed tilemap image) exists in the ECS world with an appropriate ground-layer z index.
  - Invokes the z-aware entity rendering system, which draws all sprite-bearing entities in render order so that the tilemap entity is drawn first and the tank and other entities are drawn above it.

## 5. Contracts and Data

### Map and Terrain Abstraction

The solution revolves around extending the existing map data usage with a clear contract for terrain queries. Conceptually, the map module or an adapter provides:

- **Tile lookup**
  - A function that takes integer tile coordinates (tile column and row) and returns a tile identifier string or a terrain kind.
  - Out-of-bounds behavior is well-defined (for example, returning a sentinel value or treating out-of-bounds as non-walkable), to be clearly documented.

- **World-to-tile mapping**
  - A function that takes world coordinates (as used in transform components) and returns tile coordinates and/or terrain kind.
  - This function assumes a fixed tile size and a coordinate origin; the design standardizes these assumptions so that all consumers interpret the map consistently.

### Tile Identifiers and Terrain Kinds

- **Tile identifiers**: Existing strings such as `"tileGrass1"` and `"tileGrass2"` remain the visual identifiers used by the map and assets module.
- **Terrain kinds**: For future behavior, it is useful to introduce logical terrain categories (for example, grass, road, sand, water). This design acknowledges that need and expects the map module or a simple mapping layer to associate tile identifiers with terrain kinds, even if only one or two kinds (e.g., grass variants) are present initially.

### Render Order / Z Index

To support drawing the combined tilemap image as an entity while keeping the tank and other entities visually above it, the rendering contract is extended with a render order or z index concept:

- Each sprite-bearing entity can carry a small integer or categorical render order value indicating its layer (for example, ground, actors, effects, UI).
- The tilemap entity uses the ground layer (lowest z), ensuring it is drawn first.
- The player tank and other gameplay entities use higher z values so they are drawn above the tilemap image.
- The rendering system is responsible for ordering draw calls based on this render order before issuing them to Ebiten, without exposing z details to higher-level game logic.

### Coordinate System Assumptions

To ensure consistent behavior across scenes and systems, the following assumptions are documented:

- **Tile size**: Each tile covers a fixed region in world space (for example, a square of a chosen width and height). World-to-tile mapping uses this fixed size.
- **Origin**: The world origin for the map is defined (for example, the top-left corner of the map at world coordinates (0,0)).
- **Extent**: The map’s width and height in tiles determine the playable area; behavior outside this area is defined by the out-of-bounds policy.

No new external APIs or persistent storage schemas are introduced; all contracts are internal to the game and engine-style packages.

## 6. Testing and Safety Net

Given the lite constitution mode, the safety net focuses on essential correctness of terrain queries and visible behavior of the scenes.

**Terrain query behavior**

- Unit-level tests around the map and terrain query abstraction should verify:
  - Correct tile lookup for valid tile coordinates.
  - Correct mapping from world coordinates to tile coordinates for a set of representative positions (including edges and centers of tiles).
  - Defined behavior for out-of-bounds world positions (for example, when the tank moves near or beyond the map edges).

**Scene behavior with tilemaps**

- Tests that exercise the run scene and map demo scene through their public interfaces should verify:
  - A map instance is generated at scene initialization with the expected dimensions and non-empty tile data.
  - The presence of a tilemap does not break the creation of the player entity or other entities.
  - Existing movement and firing logic continue to function as before when the tilemap is present.

**Rendering behavior (indirectly tested)**

- Direct pixel-perfect rendering tests are not required. Instead, tests can:
  - Assert that the rendering pipeline does not panic or fail when a tilemap is present.
  - Optionally, use simple properties (such as number of draw calls or non-zero tiles) in a controlled environment where unit tests can observe them without depending on specific graphics output.

Tests should run via the standard `go test ./...` command and remain fast enough for normal development and CI usage.

## 7. CI/CD and Rollout

The design introduces no new external dependencies or CI steps.

- **CI**
  - Existing workflows that build the game and run tests continue to be sufficient.
  - New tests around terrain queries and scenes integrate into the existing test suites.

- **Rollout**
  - Once merged, the new behavior (generated tilemap and layering) is always on for the map demo and main game paths; no feature flags are required.
  - If issues arise (for example, confusing visuals or regressions), the change can be rolled back by reverting the tilemap integration.

- **Compatibility**
  - No external clients or APIs are affected. Changes are purely internal to game visuals and scene behavior.

## 8. Observability and Operations

In lite mode, observability focuses on visible behavior and simple logs.

- **Logging**
  - At scene initialization, log key information such as:
    - Map seed used.
    - Map dimensions (width and height in tiles).
    - Successful map generation events and any errors encountered.
  - For debugging, optionally log terrain queries at selected points (for example, terrain under the player) when a debug mode is enabled.

- **On-screen debug aids (optional)**
  - When a debug overlay is active, display concise information such as:
    - Current tile coordinates under the tank.
    - Terrain kind under the tank.
    - Seed of the current map.

- **Operational considerations**
  - Map generation and rendering should be efficient enough not to cause noticeable slowdowns at intended resolutions and map sizes.
  - If performance issues emerge, they can be investigated via profiling and addressed in follow-up increments (for example, optimizing tile rendering or using caching strategies).

No new external metrics, dashboards, or alerts are required for this increment.

## 9. Risks, Trade-offs, and Alternatives

**Risks**

- **Visual clarity**: A naive grass-only map might be visually noisy or hard to parse, especially if tiles lack clear structure. This risk is acceptable for a first iteration but may require follow-up refinements.
- **Performance**: Compositing the tilemap image from individual tiles has an upfront cost; if recomputed frequently or for very large maps, this could impact load times or responsiveness. The design assumes moderate map sizes and infrequent recomposition.
- **Boundary semantics**: If out-of-bounds behavior for terrain queries is not clearly defined and tested, future behavior (for example, movement restrictions) could be inconsistent.

**Trade-offs**

- **Scene-owned map vs. ECS entities for tiles**: Keeping the logical map as scene-owned data and representing only the combined tilemap image as a single ECS entity is simpler and more in line with lite-mode goals than representing each tile as an ECS entity. This trades away per-tile ECS uniformity for reduced complexity and better performance while still allowing the rendering system to treat the tilemap like any other entity.
- **Limited observability**: Relying primarily on visual inspection and logs, rather than more advanced metrics or tracing, keeps the implementation simple but means subtle issues may require manual investigation.

**Alternatives considered**

- **Modeling tiles as ECS entities**: Represent each tile as an entity with components for terrain and rendering. Rejected for this increment due to added complexity, potential performance issues, and lack of immediate benefit for a small map-focused change.
- **Drawing a static background image**: Use a single pre-rendered map asset instead of generating and drawing tiles at runtime. Rejected because the increment explicitly calls for a visibly generated map and later terrain-aware behavior.

## 10. Follow-up Work

This design intentionally leaves room for future increments to build on the terrain query capability and visible tilemaps.

Potential follow-ups:

- **Terrain-dependent movement and turning**
  - Use the terrain query abstraction to adjust movement speeds and turning rates based on terrain kind (for example, faster on roads, slower on sand, blocked on water).

- **Collisions and obstacles from the map**
  - Extend the map and terrain model to include impassable or destructible tiles, and integrate this information into collision detection and response.

- **Richer map generation**
  - Enhance map generation to create structured features such as roads, choke points, and water bodies, ensuring that tile identifiers, terrain kinds, and visual assets remain in sync.

- **Spawning and AI behavior**
  - Use terrain and map structure to inform enemy spawn locations (for example, outside the player’s view on valid terrain) and AI behavior (for example, preferring roads or avoiding water).

These follow-ups should be captured as separate increments with their own designs and implementation plans.

## 11. References

- Project constitution: `CONSTITUTION.md`.
- Architecture overview: `ARCHITECTURE.md`.
- Increment definition for this change: `increment.md` in this folder.
- Seeded grass map generator design and related documentation under `docs/increments`.

## 12. Machine-Readable Artifacts (Diagrams)

### C4 Component-Level Diagram for Map-Enabled Run Scene

The following Mermaid diagram shows a component-level view of how the run scene, map module, tile renderer, ECS world, and rendering system interact after this design is implemented. It is intended as a component-level (C4 Level 3) view that can be incorporated into the project’s architecture documentation.

```mermaid
graph TD
    subgraph GameLayer[Game Layer]
        RunScene[Run Scene]
        MapDemoScene[Map Demo Scene]
        TileRenderer[Tile Renderer]
        RenderSystem[Entity Render System]
        Assets[Assets Registry]
    end

    subgraph EnginePackages[Engine-Style Packages]
        ECSWorld[ECS World]
        MapModule[Map Module
(data + terrain queries)]
        InputModule[Input Module]
    end

    subgraph External[External]
        Ebiten[Ebiten]
    end

    %% Scene ownership and flow
    RunScene --> MapModule
    RunScene --> ECSWorld
    RunScene --> TileRenderer
    RunScene --> RenderSystem
    MapDemoScene --> MapModule
    MapDemoScene --> ECSWorld
    MapDemoScene --> TileRenderer
    MapDemoScene --> RenderSystem

    %% Rendering dependencies
    TileRenderer --> Assets
    RenderSystem --> Assets
    TileRenderer --> Ebiten
    RenderSystem --> Ebiten

    %% Engine-style boundaries
    ECSWorld -. used by .- RenderSystem
    ECSWorld -. used by .- RunScene
    ECSWorld -. used by .- MapDemoScene
    InputModule -. used by .- RunScene
    InputModule -. used by .- MapDemoScene

    %% Ebiten encapsulation
    Ebiten -. drives loop .- RunScene
    Ebiten -. drives loop .- MapDemoScene
```
