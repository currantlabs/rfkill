package rfkill

import (
	"fmt"
	"syscall"
	"time"
)

// RFKill is the control device of wireless switches.
type RFKill struct {
	fd   int             // fd of the opened control device.
	sw   map[int]*Switch // current states of switches.
	fn   func(e Event)   // event handler registered with Listen().
	d    time.Duration   // time interval between event polling.
	c    chan bool       // signaling between updates.
	done chan struct{}
}

// Open returns a rfkill control device.
func Open() (*RFKill, error) {
	fd, err := syscall.Open("/dev/rfkill", syscall.O_RDWR|syscall.O_NONBLOCK, 0777)
	if err != nil {
		return nil, fmt.Errorf("can't open /dev/rfkill. %v", err)
	}

	r := &RFKill{
		fd:   fd,
		sw:   map[int]*Switch{},
		done: make(chan struct{}),
		c:    make(chan bool),
	}

	go func() {
		for {
			select {
			case <-r.c:
			case <-r.done:
				return
			}
			r.poll()
			r.c <- true
		}
	}()
	return r, nil
}

// Close closes the rfkill control device.
func (r *RFKill) Close() {
	close(r.done)
	syscall.Close(r.fd)
}

// Listen registers the provided event handler. When events are drain, it waits for d before next poll.
func (r *RFKill) Listen(fn func(e Event), d time.Duration) {
	r.fn = fn
	r.d = d
	go r.poller()
}

// Switches returns all switches which pass the filter.
func (r *RFKill) Switches(f Filter) ([]*Switch, error) {

	// signal the poller to drain events and update the switches satates.
	r.c <- true
	<-r.c

	ss := make([]*Switch, 0, len(r.sw))
	for _, s := range r.sw {
		if f(s) {
			ss = append(ss, s)
		}
	}
	return ss, nil
}

// Filter is a function, which passes switches with specified condition.
type Filter func(s *Switch) bool

// Any is a filter which passes any switch.
func Any() Filter { return func(s *Switch) bool { return true } }

// WithType is a filter which passes switches with the specified type.
func WithType(t Type) Filter { return func(s *Switch) bool { return s.Type() == t } }

// WithTypeName is a filter which passes switches with the specified type.
func WithTypeName(n string) Filter { return func(s *Switch) bool { return s.Type().String() == n } }

// WithIndex is a filter which passes switches with the specified index.
func WithIndex(i int) Filter { return func(s *Switch) bool { return s.Index() == i } }

// WithName is a filter which passes switches with the specified name.
func WithName(n string) Filter { return func(s *Switch) bool { return s.Name() == n } }

func (r *RFKill) send(e Event) error {
	switch n, err := syscall.Write(r.fd, e.marshal()); {
	case err != nil:
		return fmt.Errorf("can't write to /dev/rfkill: %s", err)
	case n != eventSize:
		return fmt.Errorf("wrong Event size: %d", n)
	}
	return nil
}

func (r *RFKill) handle(e *Event) error {
	switch e.op {
	case OpAdd:
		r.sw[e.idx] = &Switch{Event: *e, r: r}
	case OpDel:
		delete(r.sw, e.idx)
	case OpChange:
		d, ok := r.sw[e.idx]
		if !ok {
			return fmt.Errorf("changing non-existing switch idx: %d", e.idx)
		}
		d.Event = *e
	case OpChangeAll:
		for _, d := range r.sw {
			if d.typ == e.typ {
				d.Event = *e
			}
		}
	default:
		return fmt.Errorf("unknown Event op: %d", e.op)
	}
	return nil
}

func (r *RFKill) poll() error {
	var e Event
	b := make([]byte, 256)
	for {
		switch n, err := syscall.Read(r.fd, b); {
		case err == syscall.EAGAIN:
			return nil
		case err != nil:
			return fmt.Errorf("can't read /dev/rfkill: %v", err)
		case n != eventSize:
			return fmt.Errorf("wrong Event size: %d", n)
		}
		e.unmarshal(b)

		// Update switches states accordingly.
		if err := r.handle(&e); err != nil {
			return fmt.Errorf("can't handle event: %s", err)
		}

		// Notifiy switch spicific event handler, if registered.
		if s, ok := r.sw[e.Index()]; ok && s.fn != nil {
			go s.fn(e)
		}

		// Notifiy switch spicific event handler, if registered.
		if r.fn != nil {
			go r.fn(e)
		}
	}
}

func (r *RFKill) poller() {
	for {
		select {
		case r.c <- true:
		case <-r.done:
			return
		}
		<-r.c
		time.Sleep(r.d)
	}
}
