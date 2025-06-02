package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type WeatherResponse struct {
	Name    string `json:"name"`
	Main    struct {
		Temp     float64 `json:"temp"`
		Humidity int     `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
}

func getWeather(city, apiKey string) (*WeatherResponse, error) {
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", city, apiKey)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %s", resp.Status)
	}

	var data WeatherResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	city := r.URL.Query().Get("city")
	if city == "" {
		http.Error(w, "Missing city", http.StatusBadRequest)
		return
	}
	weather, err := getWeather(city, apiKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"name":      weather.Name,
		"temp":      weather.Main.Temp,
		"humidity":  weather.Main.Humidity,
		"wind":      weather.Wind.Speed,
		"condition": fmt.Sprintf("%s - %s", weather.Weather[0].Main, weather.Weather[0].Description),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/weather", weatherHandler)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Server started at http://localhost:" + port)
	http.ListenAndServe(":"+port, nil)
}
