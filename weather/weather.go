package weather

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// API endpoints for Mountain View and San Francisco
var endpoints = []string{
	"https://api.weather.gov/gridpoints/MTR/93,86/forecast",  // Mountain View
	"https://api.weather.gov/gridpoints/MTR/86,106/forecast", // San Francisco
}

// ForecastPeriod represents a single period in NOAA's response
type ForecastPeriod struct {
	StartTime                string `json:"startTime"`
	ProbabilityOfPrecipitation struct {
		Value *float64 `json:"value"`
	} `json:"probabilityOfPrecipitation"`
}

// NOAAResponse represents the structure of NOAA's API response
type NOAAResponse struct {
	Properties struct {
		Periods []ForecastPeriod `json:"periods"`
	} `json:"properties"`
}

// fetchForecast fetches and parses the NOAA forecast JSON
func fetchForecast(url string) ([]ForecastPeriod, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "GoWeatherApp (youremail@example.com)") // Required by NOAA API

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch data: status code " + resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var noaaResp NOAAResponse
	err = json.Unmarshal(body, &noaaResp)
	if err != nil {
		return nil, err
	}

	return noaaResp.Properties.Periods, nil
}

// getMaxRainProbability returns the highest chance of rain for today (0-100)
func getMaxRainProbability(periods []ForecastPeriod) int {
	today := time.Now().Format("2006-01-02") // Get today's date in YYYY-MM-DD format
	maxProbability := 0.0

	for _, period := range periods {
		if strings.HasPrefix(period.StartTime, today) && period.ProbabilityOfPrecipitation.Value != nil {
			if *period.ProbabilityOfPrecipitation.Value > maxProbability {
				maxProbability = *period.ProbabilityOfPrecipitation.Value
			}
		}
	}

	return int(maxProbability) // Convert to integer (0-100)
}

// GetMaxRainProbability fetches NOAA weather data and returns the max rain probability as an int (0-100)
func GetMaxRainProbability() int {
	maxRain := 0

	for _, url := range endpoints {
		periods, err := fetchForecast(url)
		if err != nil {
			continue // Skip errors and continue checking the other city
		}

		cityMax := getMaxRainProbability(periods)
		if cityMax > maxRain {
			maxRain = cityMax
		}
	}

	return maxRain
}
