package geo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type GAddress struct {
	Street  string  `json:"street"`
	City    string  `json:"city"`
	State   string  `json:"state"`
	Country string  `json:"country"`
	Zipcode string  `json:"zipcode"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

func (g GAddress) String() string {
	return fmt.Sprintf("%#v", g)
}

func NewGAddress(addr json.RawMessage) (GAddress, error) {
	var add GAddress
	if err := json.Unmarshal(addr, &add); err != nil {
		return GAddress{}, err
	}

	return add, nil
}

type Location struct {
	Latitude  float64
	Longitude float64
}

type GeoCoder interface {
	GeoCode(addr GAddress) Location
}

// TODO: refactor geo actor to not have address and only have inbox and recieve
// that way it can just be running and address can be passed to it

type GeoActor struct {
}

func NewGeoActor() GeoActor {
	return GeoActor{}
}

func (a *GeoActor) GeoCode(address GAddress) (*Location, error) {
	geoUrl, err := a.buildUrl(address)

	if err != nil {
		return nil, fmt.Errorf("failed to build geocode url %v", err)
	}

	req, err := http.NewRequest("POST", geoUrl, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to create geocode request %v", err)
	}

	client := http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("error sending geocode request %v", err)
	}

	var jsonRes map[string]interface{}
	decoder := json.NewDecoder(res.Body)
	defer res.Body.Close()
	if err := decoder.Decode(&jsonRes); err != nil {
		return nil, fmt.Errorf("failed to decode resposne body %v", err)
	}

	results, ok := jsonRes["results"].([]interface{})
	if !ok {
		return nil, errors.New("malformed response could not read results")
	}

	data, ok := results[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("malformed response could not read data")
	}

	geometry := data["geometry"].(map[string]interface{})
	loc := geometry["location"].(map[string]interface{})
	lat := loc["lat"].(float64)
	lng := loc["lng"].(float64)

	return &Location{Latitude: lat, Longitude: lng}, nil
}

func (a *GeoActor) buildUrl(addr GAddress) (string, error) {
	key := os.Getenv("GEOCODING_KEY")
	ep := os.Getenv("GEOCODE_EP")

	if key == "" || ep == "" {
		return "", errors.New("env variables are empty, need key and endpoint")
	}

	base, err := url.Parse(ep)
	if err != nil {
		return "", fmt.Errorf("failed to parse endpoint: %w", err)
	}

	params := url.Values{}
	params.Add("address", fmt.Sprintf("%v,%v,%v", addr.Street, addr.City, addr.State))
	params.Add("key", key)
	base.RawQuery = params.Encode()

	return base.String(), nil
}
