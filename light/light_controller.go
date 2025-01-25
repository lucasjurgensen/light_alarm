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

func TestLights() error {
	// Use atomic operation to check if already running
	if isRunning {
		return fmt.Errorf("light test already in progress")
	}

	testMutex.Lock()
	defer testMutex.Unlock()

	isRunning = true
	defer func() { isRunning = false }()

	// Initialize NeoPixel
	o := ws2811.DefaultOptions
	o.Channels[0].GpioPin = LED_PIN
	o.Channels[0].LedCount = NUM_LEDS
	o.Channels[0].Brightness = BRIGHTNESS

	device, err := ws2811.MakeWS2811(&o)
	if err != nil {
		fmt.Println("Error initializing NeoPixel:", err)
		return err
	}

	err = device.Init()
	if err != nil {
		fmt.Println("Error during device initialization:", err)
		return err
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	defer func() {
		device.Fini()
		fmt.Println("NeoPixel test complete. Exiting...")
	}()

	fmt.Println("Starting Comprehensive NeoPixel Test")

	colorFillTest(device, RED, "Red")
	// colorFillTest(device, GREEN, "Green")
	// colorFillTest(device, BLUE, "Blue")
	// colorFillTest(device, WHITE, "White")

	pixelScanTest(device, RED, "Red")
	// pixelScanTest(device, GREEN, "Green")
	// pixelScanTest(device, BLUE, "Blue")
	// pixelScanTest(device, WHITE, "White")

	// brightnessTest(device)

	turnOffStrip(device)

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
			time.Sleep(20 * time.Millisecond)
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
