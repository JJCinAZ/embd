package embd

import (
	"fmt"

	"github.com/golang/glog"
)

const (
	CapNormal int = 1 << iota
	CapI2C
	CapUART
	CapSPI
	CapGPMC
	CapLCD
	CapPWM
)

type PinDesc struct {
	N    int
	IDs  []string
	Caps int
}

type PinMap []*PinDesc

func (m PinMap) Lookup(k interface{}) (*PinDesc, bool) {
	switch key := k.(type) {
	case int:
		for i := range m {
			if m[i].N == key {
				return m[i], true
			}
		}
	case string:
		for i := range m {
			for j := range m[i].IDs {
				if m[i].IDs[j] == key {
					return m[i], true
				}
			}
		}
	}

	return nil, false
}

type pin interface {
	Close() error
}

type gpioDriver struct {
	pinMap          PinMap
	initializedPins map[int]pin
}

func newGPIODriver(pinMap PinMap) *gpioDriver {
	return &gpioDriver{
		pinMap:          pinMap,
		initializedPins: map[int]pin{},
	}
}

func (io *gpioDriver) lookupKey(key interface{}) (*PinDesc, bool) {
	return io.pinMap.Lookup(key)
}

func (io *gpioDriver) digitalPin(key interface{}) (*digitalPin, error) {
	pd, found := io.lookupKey(key)
	if !found {
		return nil, fmt.Errorf("gpio: could not find pin matching %q", key)
	}

	n := pd.N

	p, ok := io.initializedPins[n]
	if ok {
		dp, ok := p.(*digitalPin)
		if !ok {
			return nil, fmt.Errorf("gpio: sorry, pin %q is already initialized for a different mode", key)
		}
		return dp, nil
	}

	if pd.Caps&CapNormal == 0 {
		return nil, fmt.Errorf("gpio: sorry, pin %q cannot be used for digital io", key)
	}

	if pd.Caps != CapNormal {
		glog.Infof("gpio: pin %q is not a dedicated digital io pin. please refer to the system reference manual for more details", key)
	}

	dp := newDigitalPin(n)
	io.initializedPins[n] = dp

	return dp, nil
}

func (io *gpioDriver) DigitalPin(key interface{}) (DigitalPin, error) {
	return io.digitalPin(key)
}

func (io *gpioDriver) Close() error {
	for _, p := range io.initializedPins {
		if err := p.Close(); err != nil {
			return err
		}
	}

	return nil
}