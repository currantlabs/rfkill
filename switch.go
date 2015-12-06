package rfkill

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

// Type is the type of switch where the definitions must conform to those in linux kernel
type Type int

const (
	// TypeAll is the type used to specify all the switches
	TypeAll = 0
	// TypeWLAN is the type of a 802.11 wireless switch
	TypeWLAN = 1
	// TypeBluetooth is the type of a bluetooth switch
	TypeBluetooth = 2
	// TypeUWB is the type of an ultra wideband switch
	TypeUWB = 3
	// TypeWimax is the type of a WiMax switch
	TypeWimax = 4
	// TypeWWAN is the type of a wireless WAN switch
	TypeWWAN = 5
	// TypeGPS is the type of a GPS switch
	TypeGPS = 6
	// TypeFM is the type of a FM radio switch
	TypeFM = 7
	// TypeNFC is the type of a NFC switch
	TypeNFC = 8
	// TypeUnknown is the type of an Unknown switch
	TypeUnknown = 9
)

var typString = map[Type]string{
	TypeAll:       "all",
	TypeWLAN:      "wlan",
	TypeBluetooth: "bluetooth",
	TypeUWB:       "uwb",
	TypeWimax:     "wimax",
	TypeWWAN:      "wwan",
	TypeGPS:       "gps",
	TypeFM:        "fm",
	TypeNFC:       "nfc",
	TypeUnknown:   "unknown",
}

func (t Type) String() string {
	if s, ok := typString[t]; ok {
		return s
	}
	return "unkown"
}

// Switch is a switch on a RF device.
type Switch struct {
	Event
	fn func(e Event)
	r  *RFKill
}

// Name returns the name of the device.
func (s *Switch) Name() string {
	b, err := ioutil.ReadFile(fmt.Sprintf("/sys/class/rfkill/rfkill%d/name", s.idx))
	if err != nil {
		log.Printf("failed to read name, %s", err)
	}
	return strings.TrimSpace(string(b))
}

// Index returns the index of the device.
func (s *Switch) Index() int { return int(s.idx) }

// Type returns the type of the device.
func (s *Switch) Type() Type { return Type(s.typ) }

// SoftBlocked returns true if the device is soft blocked.
func (s *Switch) SoftBlocked() bool { return s.soft }

// HardBlocked returns true if the device is hard blocked.
func (s *Switch) HardBlocked() bool { return s.hard }

// Blocked returns true if the device is either soft or hard blocked.
func (s *Switch) Blocked() bool { return s.soft || s.hard }

// Block blocks(soft) the device.
func (s *Switch) Block() error { return s.r.send(Event{op: OpChange, idx: s.idx, soft: true}) }

// Unblock unblocks (soft) the device.
func (s *Switch) Unblock() error { return s.r.send(Event{op: OpChange, idx: s.idx, soft: false}) }

// Listen registers the provided event handler. When events are drain, it waits for d before next poll.
func (s *Switch) Listen(fn func(e Event), d time.Duration) {
	s.fn = fn
	go func() {
		for {
			select {
			case s.r.c <- true:
			case <-s.r.done:
				return
			}
			<-s.r.c
			time.Sleep(d)
		}
	}()
}

func (s Switch) String() string {
	return fmt.Sprintf("%d: %s (%s), Soft blocked: %t, Hard blocked: %t",
		s.idx, s.Name(), Type(s.typ), s.soft, s.hard)
}
