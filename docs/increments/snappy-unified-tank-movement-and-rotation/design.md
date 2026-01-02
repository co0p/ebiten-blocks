Design: Snappy Unified Tank Movement and Rotation
===============================================

1. Context and Problem
----------------------

This increment introduces a unified, snappy movement and rotation model for all tanks in the game. The goal is to replace the current "instant velocity" behavior with a shared model that:

- Uses linear acceleration toward bounded maximum forward and backward speeds.
- Uses bounded rotational speed for smooth, continuous turning around each tank's visual center.
- Keeps movement direction aligned with the on-screen facing direction.
- Can be driven both by player input and, later, by AI using the same control signals.

Today, tank movement is implemented as follows:

- A control system for the player reads input actions and directly sets the tank's linear and angular velocity.
- A movement system integrates velocity into position and rotation for entities that have both transform and velocity data.
- A rendering system draws sprites at the transform position but does not yet apply rotation or a well-defined pivot.

This leads to instant changes in speed (no acceleration), limited ability to tune the "feel" of tanks, and a tight coupling between player input and the concrete velocity values used by movement.

This design responds to the increment's product-level definition in `increment.md` and follows the values and guardrails described in `CONSTITUTION.md` (lite mode, small and testable increments, ECS-style architecture, blackbox-oriented tests). It focuses on tank movement and rotation only; weapons, collision, and AI decision-making remain unchanged.


2. Proposed Solution (Technical Overview)
----------------------------------------

The solution introduces a clear separation of concerns between control, movement, and rendering, while staying within the existing ECS and scene architecture:

- **Control systems** (player now, AI later) are responsible only for expressing tank movement intent in a normalized form (e.g. forward/backward and left/right turn inputs), not for setting concrete velocities.
- The **movement system** becomes the single place that turns these intents plus the tank's current state and parameters into updated linear and angular velocity, applies acceleration/deceleration and caps, and then integrates position and rotation.
- The **render system** reads transform and sprite data and draws tanks rotated around their center, so they visually pivot correctly and move in the direction they face.

Any entity that carries the required movement-related components will automatically follow the same movement model, regardless of whether its control intents come from the player or from AI. Movement parameters (max speeds, acceleration, turn rate) are kept configurable so future increments can vary them per tank type without changing the underlying model.

At a high level, each update frame behaves as follows:

1. Control systems (player input, and eventually AI) compute normalized movement intents for tanks.
2. The movement system reads the intents and current transform/velocity and adjusts velocity toward target linear and angular speeds, respecting acceleration and caps.
3. The movement system integrates velocity to produce new position and rotation.
4. The render system draws sprites at the updated transform, applying rotation around each tank's center.


3. Scope and Non-Scope (Technical)
-----------------------------------

**In scope**

- Defining a shared tank movement model based on:
  - Linear acceleration toward bounded maximum forward and backward speeds.
  - Bounded rotational speed for smooth, continuous turning.
  - Movement direction aligned with the visual facing direction.
- Refactoring control logic so that player input adjusts current velocity indirectly via normalized movement intents, rather than directly overwriting velocity.
- Updating movement logic to apply acceleration/deceleration and caps for both linear and angular components.
- Updating rendering behavior so tanks visually rotate around their center and appear to move according to their facing direction.
- Ensuring that the model is expressed in a way that AI systems can later use by providing the same movement intents as the player.

**Out of scope**

- Terrain-dependent modifiers (e.g. friction, traction) or different behavior on different surfaces.
- Complex physics such as sliding, inertia beyond what is needed for a snappy feel, or continuous collision resolution; existing collision behavior is not expanded.
- AI decision-making logic itself; AI will only be expected to provide movement intents compatible with this model in future increments.
- Per-tank-type tuning beyond supporting parameterization; detailed balancing is deferred to later increments.


4. Architecture and Boundaries
-------------------------------

The design builds on the existing layered architecture:

- **Applications (binaries)** use the game package and Ebiten for the main loop.
- **Game layer** contains scenes, systems, components, and assets.
- **Engine-style packages** provide ECS, input, and scene abstractions.

This increment affects only the game and engine-style packages, and keeps Ebiten usage confined to the same places as before (game loop, scenes, rendering, input adapter, assets).

Key architectural roles for this increment:

- **Components (data)**
  - Transform-like data tracking a tank's position, rotation, and scale.
  - Velocity-like data holding current linear and angular velocities.
  - New logical control and parameter data for tanks, such as normalized movement intents and tunable movement parameters.

- **Control systems**
  - Player control: reads high-level actions (e.g. move forward/backward, turn left/right) from the input adapter layer and translates them into normalized movement intents (e.g. throttle and turn values in a small numeric range).
  - Future AI control: will set the same movement intents based on AI logic, without manipulating transforms or velocities directly.

- **Movement system**
  - Operates on tanks that have transform, velocity, movement controls, and movement parameters.
  - Computes target linear and angular speeds from the control intents and tank parameters.
  - Adjusts current linear and angular velocity toward these targets at a bounded acceleration/deceleration rate.
  - Integrates velocity into updated position and rotation based on frame delta time.

- **Render system**
  - Uses transform and sprite information to draw rotated sprites such that the tank's visual center coincides with its logical position.
  - Applies rotation from the transform when drawing, so that the sprite's orientation matches the movement direction.

This separation preserves the following boundaries:

- Engine-style packages remain generic and free of game-specific tank logic.
- Game systems operate on ECS data and abstract input, without embedding Ebiten-specific details other than where already allowed (rendering and the input adapter).
- Input and AI remain consumers and producers of high-level actions and intents, not low-level velocity values.


5. Contracts and Data
----------------------

This increment introduces or formalizes several contracts at the data and behavior level.

### Control Intent

Control intent represents what the controlling agent (player or AI) wants a tank to do in the current frame, expressed in normalized form:

- **Throttle**: a scalar in the range [-1, +1], where:
  - +1 means "full forward".
  - 0 means "no linear movement".
  - -1 means "full backward".
- **Turn**: a scalar in the range [-1, +1], where:
  - +1 means "full turn right".
  - 0 means "no turning".
  - -1 means "full turn left".

Player input and AI both write to this intent; only the movement system reads it.

### Movement Parameters

Movement parameters define how a tank responds to control intents. At minimum, they include:

- Maximum forward speed.
- Maximum backward speed.
- Linear acceleration rate toward target speed.
- Linear deceleration rate when throttle is reduced toward zero.
- Maximum rotational speed (turn rate).
- Rotational acceleration/deceleration rate.

Initially, a single set of default parameters can be shared by all tanks; the design keeps parameters explicit so that future increments can override them per tank type.

### Movement Behavior Contract

Given control intent, parameters, and current state, the movement system must satisfy the following behavioral contract:

- For linear speed:
  - Throttle is mapped to a target linear speed in the range [−maxBackwardSpeed, +maxForwardSpeed].
  - Each update, the current linear speed moves toward this target at a bounded rate (acceleration/deceleration), such that it cannot change faster than the configured acceleration in one frame.
  - The resulting linear speed never exceeds the defined bounds.

- For rotation:
  - Turn is mapped to a target angular speed in the range [−maxTurnRate, +maxTurnRate].
  - Each update, current angular velocity moves toward this target at a bounded rotational acceleration/deceleration rate.
  - The resulting angular speed never exceeds the defined bounds.

- For position and rotation:
  - Position is updated as current position plus linear velocity multiplied by frame delta time, using the tank's rotation to define the direction of motion.
  - Rotation is updated as current rotation plus angular velocity multiplied by frame delta time.
  - Movement direction always matches the on-screen facing direction implied by rotation.

### Visual Pivot and Alignment

The visual contract for rendering is:

- The logical position stored in transform represents the tank's center in world space.
- When drawing, the sprite is rotated around its own center and then translated so that this center coincides with the transform position.
- As rotation changes, the sprite appears to pivot in place around its center rather than orbiting or wobbling.

### Machine-Readable Artifacts (Contracts)

The following JSON Schemas define the contracts for tank control intent and movement parameters at a data level.

**Control Intent (conceptual schema)**

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://example.com/schemas/control-intent.json",
  "title": "ControlIntent",
  "type": "object",
  "properties": {
    "throttle": {
      "type": "number",
      "minimum": -1.0,
      "maximum": 1.0
    },
    "turn": {
      "type": "number",
      "minimum": -1.0,
      "maximum": 1.0
    }
  },
  "required": ["throttle", "turn"],
  "additionalProperties": false
}
```

**Movement Parameters (conceptual schema)**

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://example.com/schemas/movement-params.json",
  "title": "MovementParams",
  "type": "object",
  "properties": {
    "maxForwardSpeed": { "type": "number", "minimum": 0 },
    "maxBackwardSpeed": { "type": "number", "minimum": 0 },
    "linearAcceleration": { "type": "number", "minimum": 0 },
    "linearDeceleration": { "type": "number", "minimum": 0 },
    "maxTurnRate": { "type": "number", "minimum": 0 },
    "angularAcceleration": { "type": "number", "minimum": 0 },
    "angularDeceleration": { "type": "number", "minimum": 0 }
  },
  "required": [
    "maxForwardSpeed",
    "maxBackwardSpeed",
    "linearAcceleration",
    "linearDeceleration",
    "maxTurnRate",
    "angularAcceleration",
    "angularDeceleration"
  ],
  "additionalProperties": false
}
```


6. Testing and Safety Net
--------------------------

In keeping with the lite constitution and preference for blackbox-style tests, tests should focus on observable movement behavior through the public movement and control interfaces, not on internal helper functions.

**Unit and system-level tests**

- Verify linear acceleration and speed caps:
  - Given a tank at rest with a constant forward throttle intent, repeated updates should increase its linear speed smoothly and monotonically toward the configured max forward speed, never exceeding it.
  - Given a tank at rest with a constant backward throttle intent, repeated updates should decrease its linear speed smoothly and monotonically toward the negative of the configured max backward speed, never exceeding the bound.
  - Given a tank moving at max forward speed, further forward throttle should not increase speed beyond the cap.

- Verify deceleration and quick responsiveness:
  - Given a tank moving forward and throttle suddenly set to zero, repeated updates should bring its linear speed back toward zero quickly (snappy stop) according to the deceleration rate.
  - Given a tank switching from forward to backward throttle, its speed should pass smoothly through zero and approach the backward cap without overshooting or oscillation.

- Verify rotational behavior and caps:
  - Given a tank with a constant turn intent, repeated updates should adjust angular velocity smoothly toward the configured max turn rate and then maintain it without exceeding the bound.
  - Given a tank rotating with non-zero angular velocity and turn intent set to zero, updates should bring angular velocity back toward zero quickly according to angular deceleration.

- Verify consistency across tanks:
  - Two tanks with identical parameters and control intents should exhibit the same movement over time (same position and rotation deltas for the same sequence of updates).

**Integration-style tests**

- Given a simple world with one or more tanks and a sequence of control intents applied over several frames, verify that their transforms evolve in a way consistent with the movement model (e.g. following an arc when turning and moving forward, staying in place when only turning, etc.).
- Verify that transition from no movement to movement and back again remains controllable and predictable, with no pathological oscillations or drift when intents are neutral.

These tests should run as part of the normal `go test` flow and remain fast. They should not require graphical output or Ebiten; instead, they should operate purely on ECS-level data and movement logic.


7. CI/CD and Rollout
----------------------

The design does not introduce any new external dependencies or build steps. Existing CI expectations remain sufficient:

- Building the project.
- Running automated tests (including new movement tests).

 Rollout considerations:

- Introduce the new movement behavior in a way that is compatible with the existing scene and ECS structure, so that the game can still be run and tested locally via usual commands.
- Once the increment is complete, all tanks should use the new model and no legacy instant-velocity logic should remain; do not keep runtime switches that toggle back to the old behavior.
- If serious regressions in controllability or feel are discovered, the movement logic can be reverted in a follow-up change by restoring the prior mapping from input to velocity, but this should be done explicitly rather than via hidden fallbacks.

No special deployment or release process is required; changes move through the same build-and-run workflow as other increments.


8. Observability and Operations
--------------------------------

In line with the lite constitution, observability for this increment focuses on clear, localized feedback:

- **Visual feedback**
  - Manual play sessions remain the primary way to evaluate movement feel: moving straight, rotating in place, circling obstacles, and performing quick forward/backward or left/right changes.
  - A debug overlay can optionally display current speed and angular velocity for a selected tank, along with target speeds derived from control intents, to visually confirm acceleration curves, caps, and snappy stopping behavior.

- **Logging**
  - When debugging, the movement system can emit structured logs summarizing changes in speed, rotation, and control intents for a small number of frames (for example, when a special debug flag is active).
  - Logs should remain optional and should not be enabled in normal gameplay to avoid noise.

No external monitoring, dashboards, or alerts are required for this increment; issues are expected to be caught during local playtesting and automated tests.


9. Risks, Trade-offs, and Alternatives
--------------------------------------

**Risks**

- Movement may become too twitchy or too sluggish if acceleration and speed caps are not tuned carefully, which could harm the intended "snappy but controllable" feel.
- Changes to movement curves may subtly impact difficulty, collision behavior, or how easy it is to dodge enemies and projectiles.
- Introducing additional state (parameters and control intents) may initially make the code more complex until patterns settle.

**Trade-offs**

- The design favors a straightforward, parametric acceleration model over a more physically heavy approach, balancing expressiveness and simplicity.
- Tying all tanks to a single movement model enhances consistency and reduces code duplication, at the cost of requiring parameter tuning to differentiate tank types in future increments.

**Alternatives considered**

- Only capping instantaneous velocities without introducing acceleration/deceleration, preserving the current instant-response feel. This does not meet the smooth acceleration and rotation requirements from the increment.
- Implementing a more complete physics system with detailed friction and inertia. This is more complex than what the constitution and scope call for and would likely require additional design and testing increments.


10. Follow-up Work
-------------------

The following potential follow-ups are intentionally left for later increments:

- Introduce terrain-dependent modifiers (e.g. different acceleration or max speeds for different surfaces) built on top of the shared movement model.
- Define movement profiles per tank type (e.g. heavier but tougher tanks vs lighter, faster scouts) by overriding movement parameters.
- Implement AI systems that set control intents for tanks rather than manipulating transforms or velocities directly, providing unified behavior between player-controlled and AI-controlled tanks.
- Refine or extend the movement model if playtesting reveals the need for subtler curves (e.g. non-linear acceleration) or specific behaviors like drift.


11. References
--------------

- Project constitution and development principles.
- Increment description for snappy unified tank movement and rotation.
- Existing architecture documentation describing the ECS, scene management, and Ebiten encapsulation.


12. Architecture Diagrams (Machine-Readable)
--------------------------------------------

The following Mermaid diagrams capture the core structure of this increment within the existing architecture.

**Movement and Rendering Flow (Container/Component View)**

```mermaid
flowchart LR
    subgraph EbitenLoop[Ebiten Frame Loop]
        U[Update()] --> SceneUpdate[Active Scene.Update(dt)]
        D[Draw()] --> SceneDraw[Active Scene.Draw(screen)]
    end

    subgraph RunScene[Run Scene]
        SceneUpdate --> InputPoll[Input Adapter\n(poll actions)]
        InputPoll --> ControlSystems[Control Systems\n(player, future AI)]
        ControlSystems -->|write intents| TankIntents[(Tank Control Intents)]
        TankIntents --> MovementSystem[Movement System]
        MovementSystem --> ECSWorld[(ECS World)]
        ECSWorld --> MovementSystem
        SceneDraw --> RenderSystem[Render System]
        RenderSystem --> ECSWorld
        RenderSystem --> Assets[Assets]
    end

    InputPoll --> EbitenInput[Ebiten Input State]
    RenderSystem --> EbitenDraw[Ebiten Draw APIs]
    Assets --> EbitenImages[Ebiten Images]
```

This diagram emphasizes that control systems only write intents, the movement system is solely responsible for applying them to ECS state with acceleration and caps, and the render system consumes the resulting transform for drawing.


13. Summary
-----------

This design establishes a unified, snappy movement and rotation model for all tanks by cleanly separating control, movement, and rendering responsibilities. Control systems (player now, AI later) produce normalized intents; the movement system converts these intents into bounded linear and angular motion using acceleration and deceleration; and the render system draws tanks rotated around their center based on updated transforms. The approach respects the existing ECS architecture, stays within lite-mode expectations, and is testable via blackbox-style system tests focused on observable movement behavior.
