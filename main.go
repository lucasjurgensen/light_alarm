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
	Day   string `json:"day"`
	Start int    `json:"start"` // Start time in minutes from midnight
	End   int    `json:"end"`   // End time in minutes from midnight
}

var (
	schedules = map[string]DaySchedule{
		"Monday":    {Day: "Monday", Start: 420, End: 480},
		"Tuesday":   {Day: "Tuesday", Start: 420, End: 480},
		"Wednesday": {Day: "Wednesday", Start: 420, End: 480},
		"Thursday":  {Day: "Thursday", Start: 420, End: 480},
		"Friday":    {Day: "Friday", Start: 420, End: 480},
		"Saturday":  {Day: "Saturday", Start: 420, End: 480},
		"Sunday":    {Day: "Sunday", Start: 420, End: 480},
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

	// API: Load schedule
	http.HandleFunc("/api/schedules", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(schedules)
	})

	// API: Save schedule
	http.HandleFunc("/api/schedules/save", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
		var newSchedules map[string]DaySchedule
		if err := json.NewDecoder(r.Body).Decode(&newSchedules); err != nil {
			http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
			return
		}
		mu.Lock()
		schedules = newSchedules
		mu.Unlock()
		saveToFile("schedules.json", schedules)
		w.WriteHeader(http.StatusOK)
	})

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

		light.TestLights()

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
}
