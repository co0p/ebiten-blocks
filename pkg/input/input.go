package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Action represents a high-level input action like moving or shooting.
type Action string

const (
	ActionMoveForward  Action = "move_forward"
	ActionMoveBackward Action = "move_backward"
	ActionTurnLeft     Action = "turn_left"
	ActionTurnRight    Action = "turn_right"
	ActionFire         Action = "fire"
)

// mapping from actions to ebiten keys.
var actionBindings = map[Action][]ebiten.Key{
	ActionMoveForward:  {ebiten.KeyW},
	ActionMoveBackward: {ebiten.KeyS},
	ActionTurnLeft:     {ebiten.KeyA},
	ActionTurnRight:    {ebiten.KeyD},
	ActionFire:         {ebiten.KeySpace},
}

// Manager abstracts input management so production code can use Ebiten-backed
// input while tests can install a fake implementation.
type Manager interface {
	Poll()
	IsActionDown(Action) bool
	AnyKeyPressed() bool
}

// ebitenManager uses Ebiten's keyboard state as the input source.
type ebitenManager struct {
	state map[Action]bool
}

func newEbitenManager() *ebitenManager {
	return &ebitenManager{
		state: make(map[Action]bool),
	}
}

func (m *ebitenManager) Poll() {
	for action, keys := range actionBindings {
		pressed := false
		for _, k := range keys {
			if ebiten.IsKeyPressed(k) {
				pressed = true
				break
			}
		}
		m.state[action] = pressed
	}
}

func (m *ebitenManager) IsActionDown(a Action) bool {
	return m.state[a]
}

func (m *ebitenManager) AnyKeyPressed() bool {
	return len(inpututil.PressedKeys()) > 0
}

// TestManager is a simple in-memory Manager suitable for tests.
type TestManager struct {
	State map[Action]bool
}

// NewTestManager constructs a TestManager with an empty state map.
func NewTestManager() *TestManager {
	return &TestManager{State: make(map[Action]bool)}
}

func (m *TestManager) Poll() {}

func (m *TestManager) IsActionDown(a Action) bool {
	return m.State[a]
}

func (m *TestManager) AnyKeyPressed() bool {
	for _, down := range m.State {
		if down {
			return true
		}
	}
	return false
}

var (
	defaultManager Manager = newEbitenManager()
	manager        Manager = defaultManager
)

// SetManager replaces the current input manager. Passing nil restores the
// default Ebiten-backed manager. This is primarily intended for tests.
func SetManager(m Manager) {
	if m == nil {
		manager = defaultManager
		return
	}
	manager = m
}

// Poll updates the current action state via the active Manager.
func Poll() {
	manager.Poll()
}

// IsActionDown reports whether the given action is currently active.
func IsActionDown(a Action) bool {
	return manager.IsActionDown(a)
}

// AnyKeyPressed reports whether any key was pressed in the current frame.
// This is useful for simple "press any key" screens while still keeping
// Ebiten-specific details inside the input package.
func AnyKeyPressed() bool {
	return manager.AnyKeyPressed()
}
