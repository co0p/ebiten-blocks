# Increment: Generated Tilemap Level With Tank Above Terrain

## User Story

As a player trying out the game, I want a playable scene that uses a generated tilemap for the level and clearly renders my tank above that map, so that the world feels like real terrain and I can see my tank moving over it rather than on a blank background.

## Acceptance Criteria

1. When starting the dedicated map demo experience, a level composed of tiles is visibly generated at runtime rather than being hardcoded or empty.
2. In that map demo, the player’s tank is visible and clearly rendered above the tilemap (it never appears hidden behind ground tiles, and the layering feels natural).
3. From the normal main tankismus experience, the player also plays on a generated tilemap level rather than a blank or placeholder background.
4. In the main experience, the player’s tank is again clearly visible above the tilemap layer throughout normal play.
5. Existing core behaviors (basic movement and shooting as they work today) still function as before in the main experience; introducing the generated map and layering does not remove or break them.
6. There is a simple way to verify visually (for example, via distinctive tiles or patterns) that the map is indeed generated each time, not just a static background.

## Use Case

**Actors**
- Player starting or switching into a play session.
- Game application, including a dedicated map demo experience and the main tankismus experience.

**Preconditions**
- The game can be launched on a desktop environment as today.
- The player can access both the map demo and the main tankismus experience via available options.
- Basic tank controls (movement and firing) are already available in the main experience.

**Main Flow – Map Demo**
1. The player starts the map demo experience from the available options.
2. On startup, the game generates a tile-based map for the demo scene rather than relying on a blank or static background.
3. The generated map is rendered as the visible ground layer of the demo scene.
4. The player’s tank appears positioned within this scene and is rendered clearly above the tilemap layer.
5. As the player moves the tank around, the tank remains visually on top of the map and never appears hidden behind ground tiles.
6. If the player restarts the demo or starts a new run, a generated map is again created and rendered, making it apparent that the level comes from generation rather than a fixed image.

**Main Flow – Main Tankismus Experience**
1. The player starts the standard/main game experience.
2. Before the run begins (or as part of scene setup), the game generates a tile-based map for the level.
3. The generated map is rendered as the ground layer in the main gameplay scene.
4. The player’s tank is visible in the scene and rendered clearly above the tilemap.
5. During normal play (moving, turning, firing), the tank remains clearly visible above the ground layer, and the map provides a clear sense of level structure.
6. Existing core behaviors (movement and firing) continue to work as they do today, now occurring over the visible generated map.

**Alternate / Exception Flows**
- A1: Map generation produces an unexpected or visually confusing layout.
  - The scene still loads and the tank is clearly visible above the tiles, but the level layout may need tuning in future increments.
- A2: The player prefers to play only the main experience and never opens the map demo.
  - The player still benefits from this increment because the main experience uses a generated tilemap and clear tank-over-terrain layering.
- A3: Map generation fails or is unavailable (for example, in a debug build).
  - The game may fall back to a simple, clearly visible default layout, but should not regress to a blank screen where it is unclear where the tank is relative to the world.

## Context

The project aims to deliver a top-down tank game where terrain and obstacles matter. There is design intent and prior work around map generation, but from a player’s perspective the world can still feel like a largely empty or abstract field. Players currently do not consistently see a generated tilemap forming the level, nor is it always clear that the tank is moving over a layered terrain.

This increment focuses on making a generated tilemap and clear tank-over-terrain layering visible and playable in both a focused map demo and the main game experience. It supports experimentation with level generation and visual layering without yet committing to terrain-specific movement, collisions, or advanced level progression.

## Goal

**Outcome**
- Players can see and play on a generated tilemap level both in a dedicated demo and in the main tankismus experience, with the tank clearly rendered above the terrain at all times.

**Scope**
- Introduce or adapt a map demo experience that generates and renders a tile-based level with the tank clearly layered above it.
- Ensure the main tankismus experience also uses a generated tilemap as the ground layer, maintaining clear tank-over-terrain layering.
- Preserve existing basic movement and firing behavior in the main experience.

**Non-Goals**
- Do not implement terrain-dependent movement speeds or physics in this increment.
- Do not fully define or implement collisions or destructible obstacles based on the tilemap.
- Do not commit to a final architectural representation of tilemaps; internal representation choices are left for design and implementation phases.
- Do not redesign level progression, enemy spawning, or scoring.

**Why This Is a Good Increment**
- It is a small, coherent step that makes terrain generation and layering visible to players without re-architecting the game.
- It is testable via visual behavior in both the demo and main experience and via simple checks that the map is generated each time.
- It is releasable on its own: players simply see richer levels without losing existing core behavior.

## Tasks

- Task: Players can launch a dedicated map demo experience that clearly shows a generated tilemap level with the tank visibly moving above it.
  - User/Stakeholder Impact: Curious or experimenting players can see the map generation in isolation, making the terrain system more tangible.
  - Acceptance Clues: Starting the demo reliably shows a tile-based level and tank; restarts regenerate the map; the tank is visually above the tiles at all times.

- Task: The main tankismus experience uses a generated tilemap as the ground layer instead of a blank or indistinct background.
  - User/Stakeholder Impact: All players of the main experience benefit from a more grounded sense of world structure, even if terrain behaviors are still basic.
  - Acceptance Clues: Starting a normal game presents a visible tiled level; restarting produces another generated layout rather than a fixed background.

- Task: In both demo and main experiences, the tank’s visual layering over the tilemap remains clear and consistent during normal movement and firing.
  - User/Stakeholder Impact: Players can always see where their tank is relative to the world, improving readability and immersion.
  - Acceptance Clues: During play, there are no situations where the tank appears hidden behind ground tiles or confused with the background.

- Task: Existing basic movement and firing remain intact in the main experience after the map and layering changes.
  - User/Stakeholder Impact: Returning players do not experience regressions; they gain a richer visual context without losing familiar controls.
  - Acceptance Clues: Movement and firing behave as before while now occurring over a visible tilemap; no obvious regressions are observed in quick sanity checks.

- Task: There is an easy way for someone reviewing the game to verify that maps are generated rather than static.
  - User/Stakeholder Impact: Developers, testers, and interested players can confirm that the level generation is active and not just a single hardcoded layout.
  - Acceptance Clues: Simple actions like restarting the demo or main experience produce visibly different tilemap layouts or recognizable generation patterns.

## Risks and Assumptions

- Risks
  - Generated maps may initially be visually noisy or unclear, making it harder for players to parse the environment; this may require follow-up tuning.
  - Introducing tilemap rendering and layering into the main experience could accidentally obscure the tank or other elements if not carefully tuned.
  - If the demo and main experiences diverge too much visually, players might be confused about which behavior is canonical.

- Assumptions
  - Current performance is sufficient to render a tilemap plus the tank without noticeable slowdowns on typical target machines.
  - Players will accept a first iteration where terrain is primarily visual and does not yet affect movement speeds or collisions.
  - Existing input and control schemes are adequate for navigating the new tile-based levels.

## Success Criteria and Observability

- After release, launching the map demo clearly shows a generated tiled level with the tank rendered above it.
- After release, starting the main game presents a generated tiled level instead of a blank or indistinct background, with the tank clearly layered on top.
- Quick manual checks confirm that movement and firing still behave as before, now over the visible tilemap.
- Simple repeated runs (for example, starting and restarting the demo or main experience) show that maps differ across runs or follow a recognizable generation pattern.
- Any existing lightweight checks (such as smoke tests or manual checklists) can be extended to include verification that a tilemap appears and that the tank is visible above it.

## Process Notes

- This increment should be implemented through small, safe changes that preserve existing behavior while adding the generated tilemap and layering.
- It should flow through the normal build, test, and release process used by the project, including automated tests where practical and quick manual visual checks.
- Rollout does not require special coordination: once ready, the new behavior can replace the previous background in both demo and main experiences.
- If issues arise (for example, severe visual confusion or regressions), it should be straightforward to revert to the previous level presentation while iterating on the tilemap approach.

## Follow-up Increments (Optional)

- Introduce terrain-dependent movement speeds and turning behavior based on tile types.
- Add collisions and destructible obstacles that are derived from the tilemap layout.
- Refine map generation to produce clearer, more interesting level structures (for example, roads, choke points, and open areas).
- Make enemy spawning respect the map layout, such as spawning outside the player’s visible area on valid terrain.

## PRD Entry (for docs/PRD.md)

- Increment ID: generated-tilemap-level-with-tank-above-terrain
- Title: Generated Tilemap Level With Tank Above Terrain
- Status: Proposed
- Increment Folder: docs/increments/generated-tilemap-level-with-tank-above-terrain/
- User Story: As a player trying out the game, I want a playable scene that uses a generated tilemap for the level and clearly renders my tank above that map, so that the world feels like real terrain and I can see my tank moving over it rather than on a blank background.
- Acceptance Criteria:
  - A map demo experience uses a visibly generated tilemap level with the tank clearly rendered above it.
  - The main tankismus experience uses a generated tilemap level instead of a blank or placeholder background.
  - In both experiences, the tank remains clearly visible above the tilemap during normal play.
  - Existing basic movement and firing behaviors remain intact.
  - It is straightforward to verify that maps are generated rather than static.
- Use Case Summary: Players can launch a map demo and the main game experience, both of which generate and render a tile-based level as the ground layer while keeping the tank clearly visible above it; repeated runs show that maps come from generation rather than a single fixed layout, and core controls continue to function as before.