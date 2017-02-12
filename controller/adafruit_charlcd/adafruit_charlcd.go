/*
Package Adafruit_CharLCD allows controlling the HD44780 character LCD
controller when hooked up through an MCP23017 I2C-GPIO chip.
Currently the library is write-only and does not support reading from the display controller.

Resources

This library is based three other HD44780 libraries:
	Adafruit	https://github.com/adafruit/Adafruit-Raspberry-Pi-Python-Code/blob/master/Adafruit_CharLCD/Adafruit_CharLCD.py
	hwio		https://github.com/mrmorphic/hwio/blob/master/devices/hd44780/hd44780_i2c.go
	LiquidCrystal	https://github.com/arduino/Arduino/blob/master/libraries/LiquidCrystal/LiquidCrystal.cpp
*/
package adafruit_charlcd

import (
	"time"

	"github.com/golang/glog"
	"github.com/jjcinaz/embd"
)

type entryMode byte
type displayMode byte
type functionMode byte

// RowAddress defines the cursor (DDRAM) address of the first column of each row, up to 4 rows.
// You must use the RowAddress value that matches the number of columns on your character display
// for the SetCursor function to work correctly.
type RowAddress [4]byte

var (
	// RowAddress16Col are row addresses for a 16-column display
	RowAddress16Col RowAddress = [4]byte{0x00, 0x40, 0x10, 0x50}
	// RowAddress20Col are row addresses for a 20-column display
	RowAddress20Col RowAddress = [4]byte{0x00, 0x40, 0x14, 0x54}
)

// BacklightPolarity is used to set the polarity of the backlight switch, either positive or negative.
type BacklightPolarity bool

const (
	// Negative indicates that the backlight is active-low and must have a logical low value to enable.
	Negative BacklightPolarity = false
	// Positive indicates that the backlight is active-high and must have a logical high value to enable.
	Positive BacklightPolarity = true

	writeDelay = 37 * time.Microsecond
	pulseDelay = 1 * time.Microsecond
	clearDelay = 1520 * time.Microsecond

	// Initialize display
	lcdInit     byte = 0x33 // 00110011
	lcdInit4bit byte = 0x32 // 00110010

	// Commands
	lcdClearDisplay byte = 0x01 // 00000001
	lcdReturnHome   byte = 0x02 // 00000010
	lcdCursorShift  byte = 0x10 // 00010000
	lcdSetCGRamAddr byte = 0x40 // 01000000
	lcdSetDDRamAddr byte = 0x80 // 10000000

	// Cursor and display move flags
	lcdCursorMove  byte = 0x00 // 00000000
	lcdDisplayMove byte = 0x08 // 00001000
	lcdMoveLeft    byte = 0x00 // 00000000
	lcdMoveRight   byte = 0x04 // 00000100

	// Entry mode flags
	lcdSetEntryMode   entryMode = 0x04 // 00000100
	lcdEntryDecrement entryMode = 0x00 // 00000000
	lcdEntryIncrement entryMode = 0x02 // 00000010
	lcdEntryShiftOff  entryMode = 0x00 // 00000000
	lcdEntryShiftOn   entryMode = 0x01 // 00000001

	// Display mode flags
	lcdSetDisplayMode displayMode = 0x08 // 00001000
	lcdDisplayOff     displayMode = 0x00 // 00000000
	lcdDisplayOn      displayMode = 0x04 // 00000100
	lcdCursorOff      displayMode = 0x00 // 00000000
	lcdCursorOn       displayMode = 0x02 // 00000010
	lcdBlinkOff       displayMode = 0x00 // 00000000
	lcdBlinkOn        displayMode = 0x01 // 00000001

	// Function mode flags
	lcdSetFunctionMode functionMode = 0x20 // 00100000
	lcd4BitMode        functionMode = 0x00 // 00000000
	lcd8BitMode        functionMode = 0x10 // 00010000
	lcd1Line           functionMode = 0x00 // 00000000
	lcd2Line           functionMode = 0x08 // 00001000
	lcd5x8Dots         functionMode = 0x00 // 00000000
	lcd5x10Dots        functionMode = 0x04 // 00000100
)

// ADAFRUIT_CHARLCD represents an HD44780-compatible character LCD controller.
type ADAFRUIT_CHARLCD struct {
	Connection
	eMode   entryMode
	dMode   displayMode
	fMode   functionMode
	rowAddr RowAddress
}

// NewI2C creates a new ADAFRUIT_CHARLCD connected by an I²C bus.
func NewI2C(
	i2c embd.I2CBus,
	addr byte,
	pinMap I2CPinMap,
	rowAddr RowAddress,
	modes ...ModeSetter,
) (*ADAFRUIT_CHARLCD, error) {
	return New(NewI2CConnection(i2c, addr, pinMap), rowAddr, modes...)
}

// New creates a new ADAFRUIT_CHARLCD connected by a Connection bus.
func New(bus Connection, rowAddr RowAddress, modes ...ModeSetter) (*ADAFRUIT_CHARLCD, error) {
	controller := &ADAFRUIT_CHARLCD{
		Connection: bus,
		eMode:      0x00,
		dMode:      0x00,
		fMode:      0x00,
		rowAddr:    rowAddr,
	}
	err := controller.lcdInit()
	if err != nil {
		return nil, err
	}
	err = controller.SetMode(append(DefaultModes, modes...)...)
	if err != nil {
		return nil, err
	}
	return controller, nil
}

func (controller *ADAFRUIT_CHARLCD) lcdInit() error {
	glog.V(2).Info("charlcd: initializing display")
	err := controller.WriteInstruction(lcdInit)
	if err != nil {
		return err
	}
	glog.V(2).Info("charlcd: initializing display in 4-bit mode")
	return controller.WriteInstruction(lcdInit4bit)
}

// DefaultModes are the default initialization modes for an ADAFRUIT_CHARLCD.
// ModeSetters passed in to a constructor will override these default values.
var DefaultModes []ModeSetter = []ModeSetter{
	FourBitMode,
	TwoLine,
	Dots5x8,
	EntryIncrement,
	EntryShiftOff,
	DisplayOn,
	CursorOff,
	BlinkOff,
}

// ModeSetter defines a function used for setting modes on an ADAFRUIT_CHARLCD.
// ModeSetters must be used with the SetMode function or in a constructor.
type ModeSetter func(*ADAFRUIT_CHARLCD)

// EntryDecrement is a ModeSetter that sets the ADAFRUIT_CHARLCD to entry decrement mode.
func EntryDecrement(hd *ADAFRUIT_CHARLCD) { hd.eMode &= ^lcdEntryIncrement }

// EntryIncrement is a ModeSetter that sets the ADAFRUIT_CHARLCD to entry increment mode.
func EntryIncrement(hd *ADAFRUIT_CHARLCD) { hd.eMode |= lcdEntryIncrement }

// EntryShiftOff is a ModeSetter that sets the ADAFRUIT_CHARLCD to entry shift off mode.
func EntryShiftOff(hd *ADAFRUIT_CHARLCD) { hd.eMode &= ^lcdEntryShiftOn }

// EntryShiftOn is a ModeSetter that sets the ADAFRUIT_CHARLCD to entry shift on mode.
func EntryShiftOn(hd *ADAFRUIT_CHARLCD) { hd.eMode |= lcdEntryShiftOn }

// DisplayOff is a ModeSetter that sets the ADAFRUIT_CHARLCD to display off mode.
func DisplayOff(hd *ADAFRUIT_CHARLCD) { hd.dMode &= ^lcdDisplayOn }

// DisplayOn is a ModeSetter that sets the ADAFRUIT_CHARLCD to display on mode.
func DisplayOn(hd *ADAFRUIT_CHARLCD) { hd.dMode |= lcdDisplayOn }

// CursorOff is a ModeSetter that sets the ADAFRUIT_CHARLCD to cursor off mode.
func CursorOff(hd *ADAFRUIT_CHARLCD) { hd.dMode &= ^lcdCursorOn }

// CursorOn is a ModeSetter that sets the ADAFRUIT_CHARLCD to cursor on mode.
func CursorOn(hd *ADAFRUIT_CHARLCD) { hd.dMode |= lcdCursorOn }

// BlinkOff is a ModeSetter that sets the ADAFRUIT_CHARLCD to cursor blink off mode.
func BlinkOff(hd *ADAFRUIT_CHARLCD) { hd.dMode &= ^lcdBlinkOn }

// BlinkOn is a ModeSetter that sets the ADAFRUIT_CHARLCD to cursor blink on mode.
func BlinkOn(hd *ADAFRUIT_CHARLCD) { hd.dMode |= lcdBlinkOn }

// FourBitMode is a ModeSetter that sets the ADAFRUIT_CHARLCD to 4-bit bus mode.
func FourBitMode(hd *ADAFRUIT_CHARLCD) { hd.fMode &= ^lcd8BitMode }

// EightBitMode is a ModeSetter that sets the ADAFRUIT_CHARLCD to 8-bit bus mode.
func EightBitMode(hd *ADAFRUIT_CHARLCD) { hd.fMode |= lcd8BitMode }

// OneLine is a ModeSetter that sets the ADAFRUIT_CHARLCD to 1-line display mode.
func OneLine(hd *ADAFRUIT_CHARLCD) { hd.fMode &= ^lcd2Line }

// TwoLine is a ModeSetter that sets the ADAFRUIT_CHARLCD to 2-line display mode.
func TwoLine(hd *ADAFRUIT_CHARLCD) { hd.fMode |= lcd2Line }

// Dots5x8 is a ModeSetter that sets the ADAFRUIT_CHARLCD to 5x8-pixel character mode.
func Dots5x8(hd *ADAFRUIT_CHARLCD) { hd.fMode &= ^lcd5x10Dots }

// Dots5x10 is a ModeSetter that sets the ADAFRUIT_CHARLCD to 5x10-pixel character mode.
func Dots5x10(hd *ADAFRUIT_CHARLCD) { hd.fMode |= lcd5x10Dots }

// EntryIncrementEnabled returns true if entry increment mode is enabled.
func (hd *ADAFRUIT_CHARLCD) EntryIncrementEnabled() bool { return hd.eMode&lcdEntryIncrement > 0 }

// EntryShiftEnabled returns true if entry shift mode is enabled.
func (hd *ADAFRUIT_CHARLCD) EntryShiftEnabled() bool { return hd.eMode&lcdEntryShiftOn > 0 }

// DisplayEnabled returns true if the display is on.
func (hd *ADAFRUIT_CHARLCD) DisplayEnabled() bool { return hd.dMode&lcdDisplayOn > 0 }

// CursorEnabled returns true if the cursor is on.
func (hd *ADAFRUIT_CHARLCD) CursorEnabled() bool { return hd.dMode&lcdCursorOn > 0 }

// BlinkEnabled returns true if the cursor blink mode is enabled.
func (hd *ADAFRUIT_CHARLCD) BlinkEnabled() bool { return hd.dMode&lcdBlinkOn > 0 }

// EightBitModeEnabled returns true if 8-bit bus mode is enabled and false if 4-bit
// bus mode is enabled.
func (hd *ADAFRUIT_CHARLCD) EightBitModeEnabled() bool { return hd.fMode&lcd8BitMode > 0 }

// TwoLineEnabled returns true if 2-line display mode is enabled and false if 1-line
// display mode is enabled.
func (hd *ADAFRUIT_CHARLCD) TwoLineEnabled() bool { return hd.fMode&lcd2Line > 0 }

// Dots5x10Enabled returns true if 5x10-pixel characters are enabled.
func (hd *ADAFRUIT_CHARLCD) Dots5x10Enabled() bool { return hd.fMode&lcd5x8Dots > 0 }

// SetModes modifies the entry mode, display mode, and function mode with the
// given mode setter functions.
func (hd *ADAFRUIT_CHARLCD) SetMode(modes ...ModeSetter) error {
	for _, m := range modes {
		m(hd)
	}
	functions := []func() error{
		func() error { return hd.setDisplayMode() },
		func() error { return hd.setFunctionMode() },
		func() error { return hd.setEntryMode() },
	}
	for _, f := range functions {
		err := f()
		if err != nil {
			return err
		}
	}
	return nil
}

func (hd *ADAFRUIT_CHARLCD) setEntryMode() error {
	return hd.WriteInstruction(byte(lcdSetEntryMode | hd.eMode))
}

func (hd *ADAFRUIT_CHARLCD) setDisplayMode() error {
	return hd.WriteInstruction(byte(lcdSetDisplayMode | hd.dMode))
}

func (hd *ADAFRUIT_CHARLCD) setFunctionMode() error {
	return hd.WriteInstruction(byte(lcdSetFunctionMode | hd.fMode))
}

// DisplayOff sets the display mode to off.
func (hd *ADAFRUIT_CHARLCD) DisplayOff() error {
	DisplayOff(hd)
	return hd.setDisplayMode()
}

// DisplayOn sets the display mode to on.
func (hd *ADAFRUIT_CHARLCD) DisplayOn() error {
	DisplayOn(hd)
	return hd.setDisplayMode()
}

// CursorOff turns the cursor off.
func (hd *ADAFRUIT_CHARLCD) CursorOff() error {
	CursorOff(hd)
	return hd.setDisplayMode()
}

// CursorOn turns the cursor on.
func (hd *ADAFRUIT_CHARLCD) CursorOn() error {
	CursorOn(hd)
	return hd.setDisplayMode()
}

// BlinkOff sets cursor blink mode off.
func (hd *ADAFRUIT_CHARLCD) BlinkOff() error {
	BlinkOff(hd)
	return hd.setDisplayMode()
}

// BlinkOn sets cursor blink mode on.
func (hd *ADAFRUIT_CHARLCD) BlinkOn() error {
	BlinkOn(hd)
	return hd.setDisplayMode()
}

// ShiftLeft shifts the cursor and all characters to the left.
func (hd *ADAFRUIT_CHARLCD) ShiftLeft() error {
	return hd.WriteInstruction(lcdCursorShift | lcdDisplayMove | lcdMoveLeft)
}

// ShiftRight shifts the cursor and all characters to the right.
func (hd *ADAFRUIT_CHARLCD) ShiftRight() error {
	return hd.WriteInstruction(lcdCursorShift | lcdDisplayMove | lcdMoveRight)
}

// Home moves the cursor and all characters to the home position.
func (hd *ADAFRUIT_CHARLCD) Home() error {
	err := hd.WriteInstruction(lcdReturnHome)
	time.Sleep(clearDelay)
	return err
}

// Clear clears the display and mode settings sets the cursor to the home position.
func (hd *ADAFRUIT_CHARLCD) Clear() error {
	err := hd.WriteInstruction(lcdClearDisplay)
	if err != nil {
		return err
	}
	time.Sleep(clearDelay)
	// have to set mode here because clear also clears some mode settings
	return hd.SetMode()
}

// SetCursor sets the input cursor to the given position.
func (hd *ADAFRUIT_CHARLCD) SetCursor(col, row int) error {
	return hd.SetDDRamAddr(byte(col) + hd.lcdRowOffset(row))
}

func (hd *ADAFRUIT_CHARLCD) lcdRowOffset(row int) byte {
	// Offset for up to 4 rows
	if row > 3 {
		row = 3
	}
	return hd.rowAddr[row]
}

// SetDDRamAddr sets the input cursor to the given address.
func (hd *ADAFRUIT_CHARLCD) SetDDRamAddr(value byte) error {
	return hd.WriteInstruction(lcdSetDDRamAddr | value)
}

// WriteInstruction writes a byte to the bus with register select in data mode.
func (hd *ADAFRUIT_CHARLCD) WriteChar(value byte) error {
	return hd.Write(true, value)
}

// WriteInstruction writes a byte to the bus with register select in command mode.
func (hd *ADAFRUIT_CHARLCD) WriteInstruction(value byte) error {
	return hd.Write(false, value)
}

// Close closes the underlying Connection.
func (hd *ADAFRUIT_CHARLCD) Close() error {
	return hd.Connection.Close()
}

// Connection abstracts the different methods of communicating with an ADAFRUIT_CHARLCD.
type Connection interface {
	// Write writes a byte to the ADAFRUIT_CHARLCD controller with the register select
	// flag either on or off.
	Write(rs bool, data byte) error

	// BacklightOff turns the optional backlight off.
	BacklightOff() error

	// BacklightOn turns the optional backlight on.
	BacklightOn() error

	// Close closes all open resources.
	Close() error
}

// I2CConnection implements Connection using an I²C bus.
type I2CConnection struct {
	I2C       embd.I2CBus
	Addr      byte
	PinMap    I2CPinMap
	Backlight bool
}

// I2CPinMap represents a mapping between the pins on an I²C port expander and
// the pins on the ADAFRUIT_CHARLCD controller.
type I2CPinMap struct {
	RS, RW, EN     byte
	D4, D5, D6, D7 byte
	Backlight      byte
	BLPolarity     BacklightPolarity
}

var (
	// MXXXXXPinMap is the standard pin mapping for a PCF8574-based I²C backpack.
	MCP230XXPinMap I2CPinMap = I2CPinMap{
		RS: 7, RW: 6, EN: 5,
		D4: 4, D5: 3, D6: 2, D7: 1,
		Backlight:  0,
		BLPolarity: Positive,
	}
)

// NewI2CConnection returns a new Connection based on an I²C bus.
func NewI2CConnection(i2c embd.I2CBus, addr byte, pinMap I2CPinMap) *I2CConnection {
	x := &I2CConnection{
		I2C:    i2c,
		Addr:   addr,
		PinMap: pinMap,
	}
	_ = x.I2C.WriteByteToReg(addr, 0x05, 0x00) // IOCON.BANK = 0 if it was in BANK1 mode
	_ = x.I2C.WriteByteToReg(addr, 0x0A, 0x20) // IOCON.BANK = 0, no auto-increment
	_ = x.I2C.WriteByteToReg(addr, 0x12, 0x40) // backlight off
	_ = x.I2C.WriteByteToReg(addr, 0x00, 0x00) // all A-pins to output
	_ = x.I2C.WriteByteToReg(addr, 0x01, 0x00) // all B-pins to output
	return x
}

// BacklightOff turns the optional backlight off.
func (conn *I2CConnection) BacklightOff() error {
	return conn.I2C.WriteByteToReg(conn.Addr, 0x12, 0x40) // backlight off
}

// BacklightOn turns the optional backlight on.
func (conn *I2CConnection) BacklightOn() error {
	return conn.I2C.WriteByteToReg(conn.Addr, 0x12, 0x00) // backlight on
}

// Write writes a register select flag and byte to the I²C connection.
func (conn *I2CConnection) Write(rs bool, data byte) error {
	var instructionHigh byte = 0x00
	instructionHigh |= ((data >> 4) & 0x01) << conn.PinMap.D4
	instructionHigh |= ((data >> 5) & 0x01) << conn.PinMap.D5
	instructionHigh |= ((data >> 6) & 0x01) << conn.PinMap.D6
	instructionHigh |= ((data >> 7) & 0x01) << conn.PinMap.D7

	var instructionLow byte = 0x00
	instructionLow |= (data & 0x01) << conn.PinMap.D4
	instructionLow |= ((data >> 1) & 0x01) << conn.PinMap.D5
	instructionLow |= ((data >> 2) & 0x01) << conn.PinMap.D6
	instructionLow |= ((data >> 3) & 0x01) << conn.PinMap.D7

	instructions := []byte{instructionHigh, instructionLow}
	for _, ins := range instructions {
		if rs {
			ins |= 0x01 << conn.PinMap.RS
		}
		glog.V(3).Infof("charlcd: writing to I2C: %#x", ins)
		err := conn.pulseEnable(ins)
		if err != nil {
			return err
		}
	}
	time.Sleep(writeDelay)
	return nil
}

func (conn *I2CConnection) pulseEnable(data byte) error {
	bytes := []byte{data, data | (0x01 << conn.PinMap.EN), data}
	for _, b := range bytes {
		time.Sleep(pulseDelay)
		err := conn.I2C.WriteByteToReg(conn.Addr, 0x13, b)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close closes the I²C connection.
func (conn *I2CConnection) Close() error {
	glog.V(2).Info("charlcd: closing I2C bus")
	return conn.I2C.Close()
}
