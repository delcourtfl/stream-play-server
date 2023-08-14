package main

import (
	"log"
	"syscall"
	"unsafe"
	"fmt"

	"golang.org/x/sys/windows"
	"errors"
)

var (
	user32    			= syscall.NewLazyDLL("user32.dll")
	sendInput 			= user32.NewProc("SendInput")
	mapVirtualKey		= user32.NewProc("MapVirtualKeyW")
	findWindow          = user32.NewProc("FindWindowW")
	showWindow          = user32.NewProc("ShowWindow")
	setForegroundWindow = user32.NewProc("SetForegroundWindow")
)

//////////////////////////////////////////////////////////////////////////////////////////
////////////////////// KEYBOARD //////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////

const (
	inputKeyboard    = 1
	SW_SHOWNORMAL    = 1
	KEYEVENTF_KEYDOWN     = 0x0000
	KEYEVENTF_EXTENDEDKEY = 0x0001
	KEYEVENTF_KEYUP       = 0x0002
	KEYEVENTF_SCANCODE    = 0x0008
)

type keyboardInput struct {
	wVk         uint16
	wScan       uint16
	dwFlags     uint32
	time        uint32
	dwExtraInfo uint64
}

type input struct {
	inputType uint32
	ki        keyboardInput
	padding   uint64
}

/**
 * sendKey sends a keyboard input event with the specified key code and event type to a target window.
 * It finds the target window by its title, shows it, brings it to the front, and sends the input event.
 *
 * @param keyCode The virtual key code for the key.
 * @param eventType true for key down event, false for key up event.
 * @param title The title of the target window.
 */
func sendKey(keyCode uint16, eventType bool, title string) {

	hwnd, _, _ := findWindow.Call(
		uintptr(unsafe.Pointer(nil)),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))))
	if hwnd == 0 {
		fmt.Println("Window not found")
		return
	}
	// Show the window
	_, _, _ = showWindow.Call(hwnd, SW_SHOWNORMAL)
	// Bring the window to the front
	_, _, _ = setForegroundWindow.Call(hwnd)
	/////////

	var i input
	i.inputType = 1 //INPUT_KEYBOARD
	i.ki.wVk = keyCode // virtual key code
	extended := false
	i.ki.wScan, extended = getScanCode(keyCode)
	if eventType {
		if extended {
			i.ki.dwFlags = KEYEVENTF_EXTENDEDKEY
		} else {
			i.ki.dwFlags = 0
		}
	} else {
		if extended {
			i.ki.dwFlags = KEYEVENTF_KEYUP | KEYEVENTF_EXTENDEDKEY
		} else {
			i.ki.dwFlags = KEYEVENTF_KEYUP
		}
	}
	i.ki.time = 0
	i.ki.dwExtraInfo = 0

	fmt.Printf("Virtual Key Code: 0x%X, Scan Code: 0x%X\n", keyCode, i.ki.wScan)

	ret, _, err := sendInput.Call(
		uintptr(1),
		uintptr(unsafe.Pointer(&i)),
		uintptr(unsafe.Sizeof(i)),
	)
	log.Printf("ret: %v error: %v", ret, err)
}

/**
 * getScanCode retrieves the scan code for the specified virtual key code.
 * It calls the "MapVirtualKey" function to map the virtual key code to a scan code.
 *
 * @param virtualKeyCode The virtual key code to map.
 * @return scanCode The scan code corresponding to the virtual key code.
 * @return needExtendedFlag A boolean indicating if the extended flag is needed for the key.
 */
func getScanCode(virtualKeyCode uint16) (uint16, bool) {
	ret, _, err := mapVirtualKey.Call(uintptr(virtualKeyCode), 0)
	if ret == 0 {
		log.Println(err)
		log.Println("problem scan")
		return 0, false
	}
	needExtendedFlag := false
	if virtualKeyCode >= 37 && virtualKeyCode <= 40 {
		needExtendedFlag = true
	}
	return uint16(ret), needExtendedFlag
}

//////////////////////////////////////////////////////////////////////////////////////////
////////////////////// CONTROLLER ////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////

// Bits that correspond to the Xbox 360 controller buttons.
const (
	Xbox360ControllerButtonUp            = 0
	Xbox360ControllerButtonDown          = 1
	Xbox360ControllerButtonLeft          = 2
	Xbox360ControllerButtonRight         = 3
	Xbox360ControllerButtonStart         = 4
	Xbox360ControllerButtonBack          = 5
	Xbox360ControllerButtonLeftThumb     = 6
	Xbox360ControllerButtonRightThumb    = 7
	Xbox360ControllerButtonLeftShoulder  = 8
	Xbox360ControllerButtonRightShoulder = 9
	Xbox360ControllerButtonGuide         = 10
	// no 11 ?
	Xbox360ControllerButtonA             = 12
	Xbox360ControllerButtonB             = 13
	Xbox360ControllerButtonX             = 14
	Xbox360ControllerButtonY             = 15
)

type Xbox360Controller struct {
	emulator            *Emulator
	handle              uintptr
	connected           bool
	notificationHandler uintptr
}

type Xbox360ControllerReport struct {
	native    xusb_report
	Capture   bool
	Assistant bool
}

type xusb_report struct {
	WButtons      uint16 `json:"wButtons"`
	BLeftTrigger  uint8 `json:"bLeftTrigger"`
	BRightTrigger uint8 `json:"bRightTrigger"`
	SThumbLX  int16   `json:"sThumbLX"`
	SThumbLY  int16   `json:"sThumbLY"`
	SThumbRX  int16   `json:"sThumbRX"`
	SThumbRY  int16   `json:"sThumbRY"`
}

type Emulator struct {
	handle      uintptr
	onVibration func(vibration Vibration)
}

type Vibration struct {
	LargeMotor byte
	SmallMotor byte
}

const (
	VIGEM_ERROR_NONE                        = 0x20000000
	VIGEM_ERROR_BUS_NOT_FOUND               = 0xE0000001
	VIGEM_ERROR_NO_FREE_SLOT                = 0xE0000002
	VIGEM_ERROR_INVALID_TARGET              = 0xE0000003
	VIGEM_ERROR_REMOVAL_FAILED              = 0xE0000004
	VIGEM_ERROR_ALREADY_CONNECTED           = 0xE0000005
	VIGEM_ERROR_TARGET_UNINITIALIZED        = 0xE0000006
	VIGEM_ERROR_TARGET_NOT_PLUGGED_IN       = 0xE0000007
	VIGEM_ERROR_BUS_VERSION_MISMATCH        = 0xE0000008
	VIGEM_ERROR_BUS_ACCESS_FAILED           = 0xE0000009
	VIGEM_ERROR_CALLBACK_ALREADY_REGISTERED = 0xE0000010
	VIGEM_ERROR_CALLBACK_NOT_FOUND          = 0xE0000011
	VIGEM_ERROR_BUS_ALREADY_CONNECTED       = 0xE0000012
	VIGEM_ERROR_BUS_INVALID_HANDLE          = 0xE0000013
	VIGEM_ERROR_XUSB_USERINDEX_OUT_OF_RANGE = 0xE0000014

	VIGEM_ERROR_MAX = VIGEM_ERROR_XUSB_USERINDEX_OUT_OF_RANGE + 1
)

var (
	client = windows.NewLazyDLL("ViGEmClient.dll")

	procAlloc                            = client.NewProc("vigem_alloc")
	procFree                             = client.NewProc("vigem_free")
	procConnect                          = client.NewProc("vigem_connect")
	procDisconnect                       = client.NewProc("vigem_disconnect")
	procTargetAdd                        = client.NewProc("vigem_target_add")
	procTargetFree                       = client.NewProc("vigem_target_free")
	procTargetRemove                     = client.NewProc("vigem_target_remove")
	procTargetX360Alloc                  = client.NewProc("vigem_target_x360_alloc")
	procTargetX360RegisterNotification   = client.NewProc("vigem_target_x360_register_notification")
	procTargetX360UnregisterNotification = client.NewProc("vigem_target_x360_unregister_notification")
	procTargetX360Update                 = client.NewProc("vigem_target_x360_update")
)

type VigemError struct {
	code uint
}

func NewVigemError(rawCode uintptr) *VigemError {
	code := uint(rawCode)

	if code == VIGEM_ERROR_NONE {
		return nil
	}

	return &VigemError{code}
}

func (err *VigemError) Error() string {
	switch err.code {
	case VIGEM_ERROR_BUS_NOT_FOUND:
		return "bus not found"
	case VIGEM_ERROR_NO_FREE_SLOT:
		return "no free slot"
	case VIGEM_ERROR_INVALID_TARGET:
		return "invalid target"
	case VIGEM_ERROR_REMOVAL_FAILED:
		return "removal failed"
	case VIGEM_ERROR_ALREADY_CONNECTED:
		return "already connected"
	case VIGEM_ERROR_TARGET_UNINITIALIZED:
		return "target uninitialized"
	case VIGEM_ERROR_TARGET_NOT_PLUGGED_IN:
		return "target not plugged in"
	case VIGEM_ERROR_BUS_VERSION_MISMATCH:
		return "bus version mismatch"
	case VIGEM_ERROR_BUS_ACCESS_FAILED:
		return "bus access failed"
	case VIGEM_ERROR_CALLBACK_ALREADY_REGISTERED:
		return "callback already registered"
	case VIGEM_ERROR_CALLBACK_NOT_FOUND:
		return "callback not found"
	case VIGEM_ERROR_BUS_ALREADY_CONNECTED:
		return "bus already connected"
	case VIGEM_ERROR_BUS_INVALID_HANDLE:
		return "bus invalid handle"
	case VIGEM_ERROR_XUSB_USERINDEX_OUT_OF_RANGE:
		return "xusb userindex out of range"
	default:
		return "invalid code returned by ViGEm"
	}
}

/**
 * NewEmulator initializes a new ViGEm emulator and returns a handle to it.
 * The `onVibration` callback will be called when vibration (motor) feedback is received.
 *
 * @param onVibration The callback function to be called on vibration feedback.
 * @return emulator Pointer to the Emulator struct or nil on failure.
 * @return error An error (if any) encountered during initialization.
 */
func NewEmulator(onVibration func(vibration Vibration)) (*Emulator, error) {
	handle, _, err := procAlloc.Call()

	if !errors.Is(err, windows.ERROR_SUCCESS) {
		return nil, err
	}

	libErr, _, err := procConnect.Call(handle)

	if !errors.Is(err, windows.ERROR_SUCCESS) {
		return nil, err
	}
	if err := NewVigemError(libErr); err != nil {
		return nil, err
	}

	return &Emulator{handle, onVibration}, nil
}

/**
 * Close closes the ViGEm emulator and frees allocated resources.
 *
 * @param emulator Pointer to the Emulator struct to be closed.
 * @return error An error (if any) encountered during the closing process.
 */
func (e *Emulator) Close() error {
	procDisconnect.Call(e.handle)
	_, _, err := procFree.Call(e.handle)

	return err
}

/**
 * CreateXbox360Controller creates a new Xbox 360 controller for the specified emulator.
 *
 * @param emulator Pointer to the Emulator struct to associate the controller with.
 * @return controller Pointer to the Xbox360Controller struct or nil on failure.
 * @return error An error (if any) encountered during the controller creation.
 */
func (e *Emulator) CreateXbox360Controller() (*Xbox360Controller, error) {
	handle, _, err := procTargetX360Alloc.Call()

	if !errors.Is(err, windows.ERROR_SUCCESS) {
		return nil, err
	}

	notificationHandler := func(client, target uintptr, largeMotor, smallMotor, ledNumber byte) uintptr {
		return 0
	}
	callback := windows.NewCallback(notificationHandler)

	return &Xbox360Controller{e, handle, false, callback}, nil
}

/**
 * Close closes the Xbox 360 controller and frees allocated resources.
 *
 * @param controller Pointer to the Xbox360Controller struct to be closed.
 * @return error An error (if any) encountered during the closing process.
 */
func (c *Xbox360Controller) Close() error {
	_, _, err := procTargetFree.Call(c.handle)

	return err
}

/**
 * Connect connects the Xbox 360 controller to the emulator.
 *
 * @param controller Pointer to the Xbox360Controller struct to be connected.
 * @return error An error (if any) encountered during the connection process.
 */
func (c *Xbox360Controller) Connect() error {
	libErr, _, err := procTargetAdd.Call(c.emulator.handle, c.handle)

	if !errors.Is(err, windows.ERROR_SUCCESS) {
		return err
	}
	if err := NewVigemError(libErr); err != nil {
		return err
	}

	libErr, _, err = procTargetX360RegisterNotification.Call(c.emulator.handle, c.handle, c.notificationHandler)

	if !errors.Is(err, windows.ERROR_SUCCESS) {
		return err
	}
	if err := NewVigemError(libErr); err != nil {
		return err
	}

	c.connected = true

	return nil
}

/**
 * Disconnect disconnects the Xbox 360 controller from the emulator.
 *
 * @param controller Pointer to the Xbox360Controller struct to be disconnected.
 * @return error An error (if any) encountered during the disconnection process.
 */
func (c *Xbox360Controller) Disconnect() error {
	libErr, _, err := procTargetX360UnregisterNotification.Call(c.handle)

	if !errors.Is(err, windows.ERROR_SUCCESS) {
		return err
	}
	if err := NewVigemError(libErr); err != nil {
		return err
	}

	libErr, _, err = procTargetRemove.Call(c.emulator.handle, c.handle)

	if !errors.Is(err, windows.ERROR_SUCCESS) {
		return err
	}
	if err := NewVigemError(libErr); err != nil {
		return err
	}

	c.connected = false

	return nil
}

/**
 * Send sends a controller report to the Xbox 360 emulator.
 *
 * @param controller Pointer to the Xbox360Controller struct to send the report from.
 * @param report Pointer to the Xbox360ControllerReport struct containing the report data.
 * @return error An error (if any) encountered during the sending process.
 */
func (c *Xbox360Controller) Send(report *Xbox360ControllerReport) error {
	libErr, _, err := procTargetX360Update.Call(c.emulator.handle, c.handle, uintptr(unsafe.Pointer(&report.native)))

	if !errors.Is(err, windows.ERROR_SUCCESS) {
		return err
	}
	if err := NewVigemError(libErr); err != nil {
		return err
	}

	return nil
}
