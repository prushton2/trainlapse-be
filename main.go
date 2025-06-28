package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Request-Method", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	querystring, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, "Error reading querystring (is there a querystring?)", http.StatusBadRequest)
		io.WriteString(w, "{}")
		return
	}

	// i cant name this time
	timeWindow, err := strconv.ParseInt(querystring.Get("time"), 10, 32)
	if err != nil {
		http.Error(w, "Please supply a valid 32 bit integer to the query parameter time", http.StatusBadRequest)
		io.WriteString(w, "{}")
		return
	}

	// Clamp time to be within 0 - 24 hours
	if timeWindow < 0 {
		timeWindow = 0
	} else if timeWindow > 86400 {
		timeWindow = 86400
	}

	var response Response = Response{
		Size:     0,
		Elements: make([]StoredData, 0),
	}

	files, err := os.ReadDir("./data")
	if err != nil {
		fmt.Println("Error reading ./data directory:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for _, file := range files {

		dateCreated, err := strconv.ParseInt(strings.Split(file.Name(), ".")[0], 10, 64)
		if err != nil {
			fmt.Println("Error parsing file name ", file.Name(), " to int64, continuing. :", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			continue
		}

		if time.Now().Unix()-dateCreated > timeWindow {
			continue
		}

		readFile, err := os.OpenFile(fmt.Sprintf("./data/%s", file.Name()), os.O_RDONLY, 0644) //0644 is the file perms
		if err != nil {
			fmt.Println("Error opening file:", err)
			continue
		}

		body, err := io.ReadAll(readFile)
		if err != nil {
			fmt.Println("Error reading file:", err)
			readFile.Close()
			continue
		}

		readFile.Close()

		var fileData StoredData

		err = json.Unmarshal(body, &fileData)

		response.Size += 1
		response.Elements = append(response.Elements, fileData)
	}

	responseString, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Error marshaling response to http request: ", err)
		return
	}

	io.Writer.Write(w, responseString)
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Request-Method", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	// move back a minute
	twominsago := time.Now().Unix() - 60
	// this rounds it back another minute effectively
	twominsago = twominsago - twominsago%60

	file, err := os.OpenFile(fmt.Sprintf("./data/%d.json", twominsago), os.O_RDONLY, 0644)
	if err != nil {
		http.Error(w, fmt.Sprintf("File ./data/%d.json likely doesnt exist, failing healthcheck", twominsago), http.StatusInternalServerError)
		return
	}

	var unmarshaled StoredData

	body, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading file %s", file.Name()), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &unmarshaled)
	if err != nil {
		http.Error(w, fmt.Sprintf("Data inside file %s is likely corrupt", file.Name()), http.StatusInternalServerError)
		return
	}

	io.WriteString(w, "Healthcheck passed")
}

func getTrainUpdates() {
	for {
		time.Sleep(1 * time.Second)
		now := time.Now().Unix()
		if now%60 == 0 {
			getCurrentState(now)
			deleteOldStates(now)
		}
	}
}

func deleteOldStates(now int64) {
	files, err := os.ReadDir("./data")
	if err != nil {
		fmt.Println("Error reading ./data directory:", err)
		return
	}

	for _, file := range files {

		dateCreated, err := strconv.ParseInt(strings.Split(file.Name(), ".")[0], 10, 64)
		if err != nil {
			fmt.Println("Error parsing file name ", file.Name(), " to int64, continuing. :", err)
			continue
		}

		if dateCreated >= now-86400 {
			continue
		}

		filePath := "./data/" + file.Name()
		err = os.Remove(filePath)
		if err != nil {
			fmt.Println("Error deleting file:", filePath, err)
		}
	}
}

func getCurrentState(now int64) {
	resp, err := http.Get("https://api-v3.mbta.com/vehicles?filter[route_type]=2&include=trip")
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	var response APIResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		// this only really happens when there are no trains, so we store the lack of trains.
		fmt.Println("Error unmarshaling body, writing empty array to file instead\n\nerr: ", err, "\n\nbody: ", body)

		file, err := os.OpenFile(fmt.Sprintf("./data/%d.json", now), os.O_WRONLY|os.O_CREATE, 0644) //0644 is the file perms
		if err != nil {
			fmt.Println("Error creating or opening file:", err)
			return
		}

		n, err := file.Write([]byte("[]"))
		if err != nil {
			fmt.Println("Error writing ", len([]byte("[]")), " to file, ", n, " bytes written: ", err)
			return
		}

		return
	}

	// store it in a map so it can be O(1) indexed when iterating over vehicles
	var tripDataMap map[string]TripDataResponse = make(map[string]TripDataResponse)

	for _, trip := range response.Included {
		tripDataMap[trip.ID] = trip
	}

	// rewrite it to here so theres significantly less data to store
	var storedData StoredData = StoredData{
		Data:      make([]TrainInfo, 0),
		Timestamp: now,
	}

	for _, vehicle := range response.Data {
		var train TrainInfo = TrainInfo{}
		thisTrip, exists := tripDataMap[vehicle.Relationships.Trip.Data.ID]
		if !exists {
			fmt.Printf("Trip data missing for train info %v\n", vehicle)
			continue
		}

		train.Bearing = vehicle.Attributes.Bearing
		train.Headsign = thisTrip.Attributes.Headsign
		train.Label = vehicle.Attributes.Label
		train.Latitude = vehicle.Attributes.Latitude
		train.Longitude = vehicle.Attributes.Longitude
		train.Speed = vehicle.Attributes.Speed

		storedData.Data = append(storedData.Data, train)
	}

	// each file is one "moment" of train data
	file, err := os.OpenFile(fmt.Sprintf("./data/%d.json", now), os.O_WRONLY|os.O_CREATE, 0644) //0644 is the file perms
	if err != nil {
		fmt.Println("Error creating or opening file:", err)
		return
	}

	bytes, err := json.Marshal(storedData)
	if err != nil {
		fmt.Println("Error marshalling storedData.Data: ", err)
		return
	}
	n, err := file.Write(bytes)
	if err != nil {
		fmt.Println("Error writing ", len(bytes), " to file, ", n, " bytes written: ", err)
		return
	}
}

func main() {
	http.HandleFunc("/get", handler)
	http.HandleFunc("/healthcheck", healthcheck)

	go getTrainUpdates()

	fmt.Println("Server listening on http://localhost:3000")
	http.ListenAndServe(":3000", nil)
}
