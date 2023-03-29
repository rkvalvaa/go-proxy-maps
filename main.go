package main

import (
	"fmt"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const googleMapsBaseURL = "https://maps.googleapis.com/maps/api/"

var requestCache *cache.Cache

func main() {
	requestCache = cache.New(5*time.Minute, 10*time.Minute)

	http.Handle("/", tollbooth.LimitFuncHandler(createRateLimiter(), handleRequest))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func createRateLimiter() *limiter.Limiter {
	rateLimiter := tollbooth.NewLimiter(10, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})
	rateLimiter.SetIPLookups([]string{"RemoteAddr", "X-Forwarded-For", "X-Real-IP"})
	rateLimiter.SetMethods([]string{"GET"})
	return rateLimiter
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if apiKey == "" {
		http.Error(w, "API key not found", http.StatusInternalServerError)
		return
	}

	cacheKey := r.URL.String()
	data, found := requestCache.Get(cacheKey)
	if found {
		fmt.Fprint(w, data)
		return
	}

	url := fmt.Sprintf("%s%s&key=%s", googleMapsBaseURL, r.URL.Path[1:], apiKey)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Error fetching data from Google Maps API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading Google Maps API response, might be missing correct inputs", http.StatusInternalServerError)
		return
	}

	// Check for errors in the API response
	if resp.StatusCode != http.StatusOK {
		var apiError map[string]interface{}
		if err := json.Unmarshal(body, &apiError); err != nil {
			http.Error(w, "Error processing Google Maps API response", http.StatusInternalServerError)
			return
		}
		errorMessage := "An error occurred while processing the request"
		if status, ok := apiError["status"]; ok {
			errorMessage = fmt.Sprintf("%s: %s", status, apiError["error_message"])
		}