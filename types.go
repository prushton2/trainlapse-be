package main

type APIResponse struct {
	Data     []VehicleDataResponse `json:"data"`
	Included []TripDataResponse    `json:"included"`
	Jsonapi  ApiData               `json:"jsonapi"`
}

type VehicleDataResponse struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Attributes struct {
		Bearing         *int     `json:"bearing"`
		Carriages       []string `json:"carriages"`
		CurrentStatus   string   `json:"current_status"`
		CurrentStopSeq  int      `json:"current_stop_sequence"`
		DirectionID     int      `json:"direction_id"`
		Label           string   `json:"label"`
		Latitude        float64  `json:"latitude"`
		Longitude       float64  `json:"longitude"`
		OccupancyStatus *string  `json:"occupancy_status"`
		Revenue         string   `json:"revenue"`
		Speed           float64  `json:"speed"`
		UpdatedAt       string   `json:"updated_at"`
	} `json:"attributes"`
	Links struct {
		Self string `json:"self"`
	} `json:"links"`
	Relationships struct {
		Route struct {
			Data struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		} `json:"route"`
		Stop struct {
			Data struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		} `json:"stop"`
		Trip struct {
			Data struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		} `json:"trip"`
	} `json:"relationships"`
}

type TripDataResponse struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Attributes struct {
		BikesAllowed         int    `json:"bikes_allowed"`
		BlockID              string `json:"block_id"`
		DirectionID          int    `json:"direction_id"`
		Headsign             string `json:"headsign"`
		Name                 string `json:"name"`
		Revenue              string `json:"revenue"`
		WheelchairAccessible int    `json:"wheelchair_accessible"`
	} `json:"attributes"`
	Links struct {
		Self string `json:"self"`
	} `json:"links"`
	Relationships struct {
		Route struct {
			Data struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		} `json:"route"`
		RoutePattern struct {
			Data struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		} `json:"route_pattern"`
		Service struct {
			Data struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		} `json:"service"`
		Shape struct {
			Data struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		} `json:"shape"`
	} `json:"relationships"`
}

type ApiData struct {
	Version string `json:"version"`
}

type StoredData struct {
	Data      []TrainInfo `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

type TrainInfo struct {
	Bearing   *int    `json:"bearing"`
	Label     string  `json:"label"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Speed     float64 `json:"speed"`
	Headsign  string  `json:"headsign"`
}

type Response struct {
	Size     int          `json:"size"`
	Elements []StoredData `json:"elements"`
}
