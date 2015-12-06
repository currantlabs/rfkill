package rfkill

// Event is the representation of an rfkill event.
type Event struct {
	idx  int
	typ  int
	op   int
	soft bool
	hard bool
}

const eventSize = 8

// The following definitions must conform to those in linux kernel.
const (
	// OpAdd indicates a switch has been added.
	OpAdd = 0
	// OpDel indicates a switch has been removed.
	OpDel = 1
	// OpChange indicate a switch has been changed.
	OpChange = 2
	// OpChangeAll indicates all switches or all specific type switches have been changed.
	OpChangeAll = 3
)

// Index returns the index of the device.
func (e *Event) Index() int { return int(e.idx) }

// Type returns the type of the device.
func (e *Event) Type() Type { return Type(e.typ) }

// SoftBlocked returns true if the device is soft blocked.
func (e *Event) SoftBlocked() bool { return e.soft }

// HardBlocked returns true if the device is hard blocked.
func (e *Event) HardBlocked() bool { return e.hard }

// Blocked returns true if the device is either soft or hard blocked.
func (e *Event) Blocked() bool { return e.soft || e.hard }

func (e *Event) marshal() []byte {
	b := make([]byte, 8)
	b[0] = uint8(e.idx)
	b[1] = uint8(e.idx >> 8)
	b[2] = uint8(e.idx >> 16)
	b[3] = uint8(e.idx >> 24)
	b[4] = uint8(e.typ)
	b[5] = uint8(e.op)
	b[6] = uint8(btoi(e.soft))
	b[7] = uint8(btoi(e.hard))

	return b
}

func (e *Event) unmarshal(b []byte) error {
	*e = Event{
		idx:  int(b[0]) | (int(b[1]) << 8) | (int(b[2]) << 16) | (int(b[3]) << 24),
		typ:  int(b[4]),
		op:   int(b[5]),
		soft: itob(int(b[6])),
		hard: itob(int(b[7])),
	}
	return nil
}

func itob(i int) bool { return i != 0 }

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}
