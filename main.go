package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"light_alarm/light"
)

type DaySchedule struct {
	Day     string `json:"day"`
	Start   int    `json:"start"`   // Start time in minutes from midnight
    End     int    `json:"end"`     // End time in minutes from midnight
    Enabled bool   `json:"enabled"` // Boolean indicating if schedule is enabled
}

var (
	schedules = map[string]DaySchedule{
		"Monday":    {Day: "Monday", Start: 420, End: 480, Enabled: true},
		"Tuesday":   {Day: "Tuesday", Start: 420, End: 480, Enabled: true},
		"Wednesday": {Day: "Wednesday", Start: 420, End: 480, Enabled: true},
		"Thursday":  {Day: "Thursday", Start: 420, End: 480, Enabled: true},
		"Friday":    {Day: "Friday", Start: 420, End: 480, Enabled: true},
		"Saturday":  {Day: "Saturday", Start: 420, End: 480, Enabled: true},
		"Sunday":    {Day: "Sunday", Start: 420, End: 480, Enabled: true},
	}
	mu     sync.Mutex
	led_mu sync.Mutex
)

func main() {
	// Serve static files (CSS, JS)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Serve the web interface
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/index.html")
	})

    http.HandleFunc("/api/schedules", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		loadFromFile("schedules.json")
		scheduleList := make([]DaySchedule, 0, len(schedules))
		for _, schedule := range schedules {
			scheduleList = append(scheduleList, schedule)
		}
		json.NewEncoder(w).Encode(scheduleList)
	})
	// API: Save schedule
	http.HandleFunc("/api/schedules/save", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
		var newSchedules []DaySchedule
		if err := json.NewDecoder(r.Body).Decode(&newSchedules); err != nil {
			http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
			return
		}
        mu.Lock()
        for _, schedule := range newSchedules {
           schedules[schedule.Day] = schedule
        }
		mu.Unlock()
		saveToFile("schedules.json", schedules)
		w.WriteHeader(http.StatusOK)
	})

	// Prepare light controller
	lightController := light.NewLightController(18, 380, 255)
	lightController.Initialize()

	// API: Trigger light test
	http.HandleFunc("/api/test-lights", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		if light.IsTestRunning() {
			http.Error(w, "Light test already in progress", http.StatusConflict)
			return
		}

		lightController.TestLights()

		w.WriteHeader(http.StatusOK)
	})

	// Load schedules from file (if exists)
	loadFromFile("schedules.json")

	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}

// Save schedules to a file
func saveToFile(filename string, data interface{}) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error saving to file:", err)
		return
	}
	defer file.Close()
	fmt.Printf("JURGENSENx00 -- data: %v\n", data)

	json.NewEncoder(file).Encode(data)
}

// Load schedules from a file
func loadFromFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("No existing schedule file found. Using default schedules.")
		return
	}
	defer file.Close()
	json.NewDecoder(file).Decode(&schedules)
	fmt.Printf("JURGENSENx00 -- schedules: %v\n", schedules)
}