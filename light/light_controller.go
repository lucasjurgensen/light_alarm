package light

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
)

// Configuration
const (
	LED_PIN       = 18
	NUM_LEDS      = 380
	BRIGHTNESS    = 255 // Maximum is 255
	TEST_DELAY_MS = 500 // Delay in milliseconds
)

// Colors
var (
	RED   = [3]uint8{255, 0, 0}
	GREEN = [3]uint8{0, 255, 0}
	BLUE  = [3]uint8{0, 0, 255}
	WHITE = [3]uint8{255, 255, 255}
	BLACK = [3]uint8{0, 0, 0}
)

type LightController struct {
	pin        int
	numLeds    int
	brightness uint8
    device     *ws2811.WS2811
}

// NewLightController creates a new light controller instance
func NewLightController(pin int, numLeds int, brightness uint8, device *ws2811.WS2811) *LightController {
	return &LightController{
		pin:        pin,
		numLeds:    numLeds,
		brightness: brightness,
		device:     device,
	}
}

// Initialize sets up the LED device
func (lc *LightController) Initialize() (*ws2811.WS2811, error) {
    opts := ws2811.DefaultOptions
    opts.Channels[0].GpioPin = lc.pin
    opts.Channels[0].LedCount = lc.numLeds
    opts.Channels[0].Brightness = int(lc.brightness) // Convert uint8 to int

    dev, err := ws2811.MakeWS2811(&opts)
    if err != nil {
        return nil, fmt.Errorf("failed to create WS2811: %v", err)
    }

    if err := dev.Init(); err != nil {
        return nil, fmt.Errorf("failed to initialize WS2811: %v", err)
    }

    lc.device = dev

    return dev, nil
}

func (lc *LightController) SunriseAlarm() error {
	fmt.Println("Starting alarm!")

	// Set the color to white
    lc.device.SetBrightness(0, 0)
	lc.SetColor(WHITE)

	for i := 25; i <= 250; i += 25 {
		fmt.Printf("Set brightness to %d\n", i)
		lc.device.SetBrightness(0, i)
		lc.SetColor(WHITE)
		time.Sleep(2 * time.Second)
	}
	time.Sleep(10 * time.Minute)
    lc.device.SetBrightness(0, 0)
	fmt.Println("All done!")
	return nil
}


// SetColor sets all LEDs to the specified color
func (lc *LightController) SetColor(color [3]uint8) error {
    for i := 0; i < lc.numLeds; i++ {
        // Convert RGB to uint32 color value
        colorVal := uint32(color[0])<<16 | uint32(color[1])<<8 | uint32(color[2])
        lc.device.Leds(0)[i] = colorVal
    }
    return lc.device.Render()
}

// Clear turns off all LEDs
func (lc *LightController) Clear() error {
    return lc.SetColor(BLACK)
}

// Add a mutex to prevent multiple simultaneous tests
var testMutex sync.Mutex
var isRunning bool
var cancelTest chan struct{}

func init() {
    cancelTest = make(chan struct{})
}

func CancelTest() {
    if isRunning {
        cancelTest <- struct{}{}
    }
}

func (lc *LightController) TestLights() error {
    // Use atomic operation to check if already running
    if isRunning {
        return fmt.Errorf("light test already in progress")
    }

    testMutex.Lock()
    defer testMutex.Unlock()

    isRunning = true
    defer func() { isRunning = false }()

    signalChan := make(chan os.Signal, 1)
    signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
    // defer func() {
    //  lc.device.Fini()
    //  fmt.Println("NeoPixel test complete. Exiting...")
    // }()

    fmt.Println("Starting Comprehensive NeoPixel Test")

    colorFillTest(lc.device, RED, "Red")
    // colorFillTest(device, GREEN, "Green")
    // colorFillTest(device, BLUE, "Blue")
    // colorFillTest(device, WHITE, "White")

    pixelScanTest(lc.device, RED, "Red")
    // pixelScanTest(device, GREEN, "Green")
    // pixelScanTest(device, BLUE, "Blue")
    // pixelScanTest(device, WHITE, "White")

    // brightnessTest(device)

    turnOffStrip(lc.device)

    return nil
}

func colorFillTest(device *ws2811.WS2811, color [3]uint8, name string) {
    fmt.Printf("\n--- Color Fill Test: %s ---\n", name)
    fillStrip(device, color)
    device.Render()
    time.Sleep(TEST_DELAY_MS * time.Millisecond)
}

func pixelScanTest(device *ws2811.WS2811, color [3]uint8, name string) {
    fmt.Printf("\n--- Pixel Scan Test: %s ---\n", name)
    for i := 0; i < NUM_LEDS; i++ {
        select {
        case <-cancelTest:
            return
        default:
            setPixelColor(device, i, color)
            device.Render()
            time.Sleep(1 * time.Millisecond)
            setPixelColor(device, i, BLACK)
            device.Render()
        }
    }
}

func brightnessTest(device *ws2811.WS2811) {
    fmt.Println("\n--- Brightness Test ---")
    brightnessLevels := []uint8{25, 128, 255}
    for _, level := range brightnessLevels {
        fmt.Printf("\nSetting brightness to %d\n", level)
        device.SetBrightness(0, int(level))
        fillStrip(device, WHITE)
        device.Render()
        time.Sleep(TEST_DELAY_MS * time.Millisecond)
    }
    device.SetBrightness(0, BRIGHTNESS) // Reset brightness
}

func turnOffStrip(device *ws2811.WS2811) {
    fmt.Println("\n--- Turning Off LEDs ---")
    fillStrip(device, BLACK)
    device.Render()
}

func fillStrip(device *ws2811.WS2811, color [3]uint8) {
    for i := 0; i < NUM_LEDS; i++ {
        setPixelColor(device, i, color)
    }
}

func setPixelColor(device *ws2811.WS2811, index int, color [3]uint8) {
    device.Leds(0)[index] = uint32(color[0])<<16 | uint32(color[1])<<8 | uint32(color[2])
}

// Add a function to check if test is running
func IsTestRunning() bool {
    return isRunning
}