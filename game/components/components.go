package components

import "github.com/co0p/tankismus/pkg/ecs"

// Type IDs used with the generic ECS world.
//
// Note: these values are used as bit positions in a uint64 mask,
// so the total number of distinct component types must stay below 64.
const (
	TypeTransform ecs.ComponentType = iota
	TypeVelocity
	TypePlayerTag
	TypeEnemyTag
	TypeHealth
	TypeSprite
	TypeCollider
	TypeProjectile
	TypeControlIntent
	TypeMovementParams
	TypeRenderOrder
)

// Transform represents position, rotation and uniform scale.
type Transform struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Rotation float64 `json:"rotation"`
	Scale    float64 `json:"scale"`
}

func (Transform) Type() ecs.ComponentType { return TypeTransform }

// Velocity represents linear and angular velocity.
type Velocity struct {
	VX      float64 `json:"vx"`
	VY      float64 `json:"vy"`
	Angular float64 `json:"angular"`
}

func (Velocity) Type() ecs.ComponentType { return TypeVelocity }

// PlayerTag marks the player entity.
type PlayerTag struct {
	IsPlayer bool `json:"is_player"`
}

func (PlayerTag) Type() ecs.ComponentType { return TypePlayerTag }

// EnemyTag marks enemy entities.
type EnemyTag struct {
	IsEnemy bool `json:"is_enemy"`
}

func (EnemyTag) Type() ecs.ComponentType { return TypeEnemyTag }

// Health represents hit points.
type Health struct {
	Current float64 `json:"current"`
	Max     float64 `json:"max"`
}

func (Health) Type() ecs.ComponentType { return TypeHealth }

// Sprite identifies which sprite to render for an entity.
type Sprite struct {
	SpriteID string `json:"sprite_id"`
}

func (Sprite) Type() ecs.ComponentType { return TypeSprite }

// Collider is a simple axis-aligned bounding box.
type Collider struct {
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
	OffsetX float64 `json:"offset_x"`
	OffsetY float64 `json:"offset_y"`
}

func (Collider) Type() ecs.ComponentType { return TypeCollider }

// Projectile marks projectile entities.
type Projectile struct {
	Speed    float64 `json:"speed"`
	Lifetime float64 `json:"lifetime"`
	Damage   float64 `json:"damage"`
}

func (Projectile) Type() ecs.ComponentType { return TypeProjectile }

// ControlIntent represents normalized control input for a tank.
// Throttle and Turn are expected to be in the range [-1, 1].
type ControlIntent struct {
	Throttle float64 `json:"throttle"`
	Turn     float64 `json:"turn"`
}

func (ControlIntent) Type() ecs.ComponentType { return TypeControlIntent }

// MovementParams defines how a tank responds to control intent.
// All values are non-negative configuration parameters.
type MovementParams struct {
	MaxForwardSpeed     float64 `json:"maxForwardSpeed"`
	MaxBackwardSpeed    float64 `json:"maxBackwardSpeed"`
	LinearAcceleration  float64 `json:"linearAcceleration"`
	LinearDeceleration  float64 `json:"linearDeceleration"`
	MaxTurnRate         float64 `json:"maxTurnRate"`
	AngularAcceleration float64 `json:"angularAcceleration"`
	AngularDeceleration float64 `json:"angularDeceleration"`
}

func (MovementParams) Type() ecs.ComponentType { return TypeMovementParams }

// RenderOrder represents a simple z-order / render layer for sprite-bearing
// entities. Lower values are drawn before higher values. When absent, systems
// should treat the entity as using the default layer.
type RenderOrder struct {
	Z int `json:"z"`
}

func (RenderOrder) Type() ecs.ComponentType { return TypeRenderOrder }
