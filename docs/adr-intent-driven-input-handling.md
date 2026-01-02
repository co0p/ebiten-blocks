# ADR: Intent-Driven Input Handling

**Status**: Accepted  \
**Date**: 2026-01-02

## Context

Originally, player input in Tankismus was wired directly into physics state:

- `pkg/input` exposed high-level actions (move forward/backward, turn left/right, fire) that were mapped to concrete keys.
- `InputMovementSystem` read those actions and **directly set** the player tank's linear and angular velocity.
- `MovementSystem` only integrated velocity into position/rotation.

This had several drawbacks:

- **Tight coupling between input and physics**: player input bypassed any shared movement model; changing the movement "feel" required adjusting both input and movement logic.
- **Limited extensibility to AI**: AI would have to manipulate velocity and rotation directly, duplicating logic instead of reusing the same control surface as the player.
- **Testing friction**: tests that needed to simulate input were coupled to Ebiten key state and a global `pkg/input` implementation, making them brittle and harder to reason about.

At the same time, the "Snappy Unified Tank Movement and Rotation" increment defined a movement model that should be driven by **normalized control intents** (throttle/turn) and **movement parameters**, not by instantaneous, unbounded velocities. That design implied a clearer separation between:

- Device/input-layer concerns (keys, actions, input polling), and
- Game-domain control intent (what a tank *wants* to do this frame), and
- The movement system (how that intent translates into motion under constraints).

## Decision

We introduce an **intent-driven input handling** approach with two key abstractions:

1. **Input manager interface at the adapter layer (`pkg/input`)**
2. **Control intent component at the game domain layer (`game/components`)**

### 1. Input Manager Interface (`pkg/input`)

`pkg/input` now exposes a small interface that represents the source of high-level actions:

```go
type Manager interface {
    Poll()
    IsActionDown(Action) bool
    AnyKeyPressed() bool
}
```

Two implementations are provided:

- **Ebiten-backed manager** (production)
  - Polls Ebiten's keyboard state and updates internal action state.
  - Used by default in the running game.

- **Test manager** (tests)
  - Holds an in-memory `map[Action]bool` state.
  - `Poll` is a no-op; tests modify `State` directly.
  - Allows tests to simulate arbitrary input sequences without relying on Ebiten.

A global `SetManager` function switches the active manager, enabling tests to install a `TestManager` while production code continues to use the default Ebiten-backed manager.

### 2. Control Intent Component (`game/components`)

We introduce a `ControlIntent` ECS component representing the **per-entity movement intent**:

```go
type ControlIntent struct {
    Throttle float64 // [-1, 1]
    Turn     float64 // [-1, 1]
}
```

Semantics:

- `Throttle` in [-1, 1]: -1 = full reverse, 0 = neutral, +1 = full forward.
- `Turn` in [-1, 1]: -1 = full left, 0 = neutral, +1 = full right.

This component is written by control systems (player, later AI) and read by the movement system.

### Control Flow Per Frame

With these abstractions, a typical frame proceeds as follows in the run scene:

1. `pkg/input.Poll()` is called, delegating to the current `Manager` implementation.
2. `InputMovementSystem`:
   - Reads `input.IsActionDown` for movement-related actions.
   - Computes normalized `Throttle` and `Turn` values.
   - Writes them into the entity's `ControlIntent` component.
3. `MovementSystem`:
   - Reads `ControlIntent`, `MovementParams`, `Transform`, and `Velocity`.
   - Computes target linear and angular speeds from intent and parameters.
   - Adjusts current velocities with acceleration/deceleration and caps.
   - Integrates velocity into position and rotation.

AI and other non-player controllers will later produce `ControlIntent` values in exactly the same format, reusing the same movement model without needing to know about input devices or keys.

## Rationale

This decision aligns Tankismus's input and movement design with widely recommended game architecture practices:

- **Decouple input devices from game logic**:
  - Instead of scattering direct key checks throughout the game, a single adapter layer (`pkg/input`) translates device state into semantic actions.
  - The rest of the code depends on actions and intents, not keys or Ebiten APIs.

- **Use intents/commands rather than raw input**:
  - By representing what the player/AI *wants* the tank to do (throttle/turn), the movement system can own how that intent is realized (acceleration curves, caps, responsiveness).
  - This makes it easier to tune "feel" without touching input code and to keep movement behavior consistent across different controllers.

- **Improve testability and portability**:
  - The `Manager` interface allows tests to simulate input without Ebiten.
  - Systems that operate on `ControlIntent` and movement components are pure ECS logic, independent of the concrete input backend.

These ideas are consistent with patterns described in game development literature, where input handling is often modeled as a pipeline from **device → actions → commands/intents → game state changes**, instead of wiring devices directly to physics or game state.

## Consequences

### Positive

- **Single movement model**: All controllers (player now, AI later) drive tanks through the same `ControlIntent`/`MovementParams` interface, so behavior stays consistent and centralised in `MovementSystem`.
- **Easier tuning**: Changing acceleration, caps, or responsiveness does not require changing input or AI code; it only affects the movement system and parameters.
- **Testable input paths**: Tests can install a `TestManager` and validate that input systems correctly set `ControlIntent` and that the movement system responds appropriately.
- **Reduced Ebiten coupling**: Most game logic (systems, components, scenes) no longer depends on Ebiten directly; only the input adapter, rendering, and top-level game loop do.

### Negative / Trade-offs

- **More indirection**: There are additional layers (actions → intent → movement) compared to directly setting velocity, which may be more complex to follow initially.
- **Global manager**: Using a global `Manager` is pragmatic but still a global; care is needed in tests to reset it between cases.
- **Extra components**: ECS entities need additional components (`ControlIntent`, `MovementParams`) for movement, which mildly increases boilerplate when creating entities.

## Alternatives Considered

1. **Keep direct velocity control from input systems**
   - Simpler wiring, but tightly couples input and movement behavior, making it harder to tune or reuse for AI.
   - Tests remain entangled with Ebiten-specific input state.

2. **Hard-code input handling inside movement system**
   - Movement system would read keys directly and update velocities.
   - Violates separation of concerns and further entangles Ebiten with game-domain logic.

3. **Command objects without normalized intent**
   - Use discrete command objects (e.g. "StartMoveForward", "StopTurnLeft") instead of continuous throttle/turn values.
   - Could work, but the continuous control model (throttle/turn in [-1, 1]) maps more naturally to the tank movement design and tuning requirements for this project.

## References

These sources advocate similar separations between devices, actions, and game-domain intent, and influenced this decision:

- **Game Programming Patterns – Input Handling (Robert Nystrom)**
  - Describes decoupling input devices from game logic using command objects and centralized input handling.
- **Unity Input System – Input Actions**
  - Unity's input system introduces *actions* as an abstraction between physical inputs (keys, gamepads) and gameplay logic, allowing the same gameplay code to work across devices.
- **Godot Engine – Input and Action Mapping**
  - Godot's input map lets developers define high-level actions and bind them to multiple physical inputs, encouraging code to respond to actions rather than keys.
- **General ECS and game architecture practices**
  - Many ECS-based engines and frameworks promote separating input gathering, intent/command generation, and state mutation into distinct systems to keep code testable and portable.

These references are conceptual rather than prescriptive; the concrete `Manager` and `ControlIntent` designs in Tankismus are tailored to this codebase and the snappy tank movement increment.
