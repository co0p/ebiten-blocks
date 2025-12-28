# tankismus

The followup AI supported iteration building a top down panzer game in golang, using the ebiten engine and ecs pattern.

## game design docs

### game overview

- **Genre**: Top-down arcade action / survival shooter  
- **Perspective**: Top-down, camera centered on the player tank  
- **Core Loop**: Move, dodge, and shoot through waves of enemy tanks while managing positioning and terrain advantages.  
- **Primary Mode**: Survival / Horde (endless or long-running session, high-score focused)  
- **Target Session Length**: Short, replayable runs (e.g. 5–20 minutes) with increasing difficulty over time.  
- **Platform**: Desktop (PC), keyboard-centric controls (future: optional gamepad).

### core concept

A fast paced top down action showcase game for the developer to tackle most problems in game design.  
The focus is on responsive tank controls, escalating enemy waves, and terrain-based movement and combat decisions.

### design pillars

The design revolves around the following design pillars:

- **minimal ui** – No real UI elements should be visible other than essential feedback (health, ammo, score, wave, or similar).
- **immersion** – Through fullscreen, screen shake, sound and light effects, and responsive controls.
- **retro** – Pixel graphics and retro sound effects.
- **action** – High frequency of shooting, explosions, and destruction.
- **escalation** – The longer you play, the harder and more extreme the game becomes (more enemies, new enemy types, higher intensity).
- **endurance** – Players are rewarded for surviving longer (e.g. score multipliers, unlocks, ..).

### features & mechanics

#### player tank

The player controls a single tank from a camera that remains centered on the vehicle. The tank can move forward and backward and rotate freely along a full 360° axis.

- **Movement**:
  - Tank-style movement (forward/back on facing direction, rotation of movement direction).
  - Movement speed depends on terrain (see “Terrain & obstacles”).
- **Combat**:
  - Primary cannon with cooldown.
  - Projectiles can destroy certain obstacles.
  - Projectiles can fly only a limitted amount
- **Health & death**:
  - The player tank has a health pool.
  - On reaching 0 health, the run ends (game over).

#### game mode: survival / horde

The core game mode is a **Survival / Horde** mode:

- The player must withstand **increasingly difficult waves** of enemy tanks.
- The goal is to **survive as long as possible** and/or **reach the highest score**.
- Difficulty escalates over time through:
  - Higher enemy counts per wave.
  - Stronger enemy stats (health, damage, speed, accuracy).
  - Smarter behaviors (better pathfinding, flanking, more aggressive pursuit).
  - Possible introduction of new enemy variants (TBD).

Optional clarifications to define later:

- Wave structure: discrete waves vs continuous spawning with ramping intensity.
- Between-wave downtime: brief pause for breathing / positioning vs no pause.
- Scoring: points per kill, combo/multiplier system, time survived, bonus for wave completion, etc.

#### terrain & obstacles

The game map contains various terrain types and obstacles that impact movement and combat.

- **Obstacles**:
  - Solid objects that block movement.
  - Certain obstacles can be destroyed by tank weapon fire (e.g. walls, crates).
  - Indestructible obstacles define permanent chokepoints and cover (TBD: which ones).
- **Terrain effects on movement speed**:
  - **Grass**: normal movement speed.
  - **Street/Road**: increased movement speed.
  - **Sand**: reduced movement speed.
  - **Water**: no movement possible; the tank cannot enter or cross water tiles.
- Optional future design hooks:
  - Different traction / turning behavior per terrain.
  - Explosive barrels or secondary hazards.

#### enemies

Enemy forces consist of continuous waves of hostile tanks that spawn **outside the player’s visible area** to preserve immersion and avoid “popping in”.

As the survival session progresses, enemy tanks scale in **difficulty** and **behavior complexity**:

- **Early waves**: slower, weaker tanks with simple chase or patrol behaviors.
- **Mid/late waves**: faster tanks, higher health/damage, more accurate and coordinated behaviors (e.g. circling, flanking, backing off when low on health – TBD).

Each enemy tank is defined by:

- **Field of view** – The angular area and distance used to detect the player.
- **Health value** – Total health before destruction.
- **Shooting cooldown** – Time between shots.
- **Shooting accuracy** – How closely shots track the player’s actual position.
- **Behavior state** (TBD, example: patrol, chase, attack, retreat).

Spawning rules to clarify (future detail):

- Spawn cadence (per time, per wave).
- Spawn locations (off-screen edges, specific spawn points).
- Maximum enemies active at once.

#### failure & victory conditions

- **Failure**: The survival run ends when the player’s tank health is fully depleted.
- **Victory / completion**:
  - For pure endless mode: no “win,” only survival time and score.
  - Optional: cap at a final wave or boss and define a win condition (TBD).

### interface & controls

The tank is controlled using **WASD** keys:

- `W` / `S`: Move forward / backward relative to tank facing.
- `A` / `D`: Rotate tank inkl. fixed turret left and right.

Clarifications / planned additions:

- **Firing**: Space key.
- **Aiming**:
  - Aim along tank facing direction only.
- **UI elements**:
  - Minimal HUD: health, score, wave, maybe ammo/cooldown indicators.
- **Feedback**:
  - Screen shake on hits/explosions.
  - Flash / sound cues for low health, damage taken, and wave starts.

### visual / audio style

Using pixel art and retro style sound effects.

Additional clarifications:

- **Resolution / scale**:
  - Target pixel resolution and scale factor (e.g. 320×180 scaled to fullscreen).
- **Palette**:
  - Limited color palette for retro feel (TBD).
- **Effects**:
  - Explosion animations, muzzle flashes, bullet trails.
  - Dynamic lighting / fake light cones if feasible.
- **Audio**:
  - Looped background track (chiptune / retro).
  - Distinct sounds for:
    - Firing cannon.
    - Enemy fire.
    - Hits and explosions.
    - Movement (treads on different terrain, optional).
  - Volume and mix: prioritize feedback (shots, hits) over ambient sounds.

## implementation details

### package layout

Tries to follow [Ardan Labs design philosophy on packaging](https://www.ardanlabs.com/blog/2017/02/design-philosophy-on-packaging.html).

### architecture

Tries to follow the **ECS pattern** (Entity-Component-System) for separation of data and behavior:

- **Entities**: tanks, projectiles, terrain objects.
- **Components**: position, movement, rendering, health, AI, etc.
- **Systems**: movement system, collision system, rendering system, AI system, spawning system.

### technical goals

- Serve as a **sandbox** for experimenting with:
  - ECS in Go.
  - Ebiten rendering and input.
  - Simple AI behaviors and spawning systems.
- Maintain **readable, educational code** suitable for blog posts and examples.

## references

- https://github.com/brotherhood-of-recursive-descent/tankism
- https://ebitengine.org/
- https://co0p.github.io/posts/ecs-animation/
- https://gameprogrammingpatterns.com/contents.html
- https://kenney.nl/assets/top-down-tanks-redux
