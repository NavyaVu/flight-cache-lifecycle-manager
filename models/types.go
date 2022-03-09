package models

import "time"

type OpenSearchResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore float64 `json:"max_score"`
		Hits     []struct {
			Index  string  `json:"_index"`
			Type   string  `json:"_type"`
			ID     string  `json:"_id"`
			Score  float64 `json:"_score"`
			Source Result  `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

//TODO add time stamp prop
type Result struct {
	Routes           map[string]Route   `json:"routes"`
	Segments         map[string]Segment `json:"segments"`
	Combinations     []Combination      `json:"combinations"`
	Ancillaries      []Ancillary        `json:"ancillaries"`
	AdditionalParams map[string]string  `json:"additionalParams,omitempty"`
	TimeStamp        time.Time          `json:"timestamp"`
}

type Route struct {
	Id                       string            `json:"id"`
	Stops                    int8              `json:"stops"`
	ElapsedFlyingTimeMinutes int               `json:"elapsedFlyingTimeMinutes"`
	SegmentIDs               []string          `json:"segmentIDs"`
	AdditionalParams         map[string]string `json:"additionalParams,omitempty"`
}

type Segment struct {
	Id                  string            `json:"id"`
	Origin              string            `json:"origin"`
	OriginTerminal      string            `json:"originTerminal"`
	Destination         string            `json:"destination"`
	DestinationTerminal string            `json:"destinationTerminal"`
	DepartureTime       string            `json:"departureTime"`
	ArrivalTime         string            `json:"arrivalTime"`
	FlightNumber        string            `json:"flightNumber"`
	AirplaneType        string            `json:"airplaneType"`
	MarketingCarrier    string            `json:"marketingCarrier"`
	OperationCarrier    string            `json:"operatingCarrier"`
	AdditionalParams    map[string]string `json:"additionalParams,omitempty"`
}

type Combination struct {
	TotalFareAmount  float64           `json:"totalFareAmount"`
	TotalTaxAmount   float64           `json:"totalTaxAmount"`
	Fares            []TfmFare         `json:"fares"`
	RouteIDs         []string          `json:"routeIDs"`
	TariffType       string            `json:"tariffType"`
	AdditionalParams map[string]string `json:"additionalParams,omitempty"`
}

type TfmFare struct {
	PaxId        string        `json:"paxId"`
	PaxType      string        `json:"paxType"`
	FareAmount   float64       `json:"fareAmount"`
	TaxAmount    float64       `json:"taxAmount"`
	FareProducts []FareProduct `json:"fareProducts"`
	Vcc          string        `json:"vcc"`
}

type FareProduct struct {
	SegmentID        string            `json:"segmentID"`
	CabinProduct     string            `json:"cabinProduct"`
	FareBase         string            `json:"fareBase"`
	AncillaryIDs     []string          `json:"ancillaryIDs"`
	AdditionalParams map[string]string `json:"additionalParams,omitempty"`
}

type Ancillary struct {
	Id               string            `json:"id"`
	Type             string            `json:"type"`
	AdditionalParams map[string]string `json:"additionalParams,omitempty"`
}
