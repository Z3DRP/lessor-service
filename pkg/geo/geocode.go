package geo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

type Location struct {
	Latitude  float64
	Longitude float64
}

type AddressEncoder interface {
	EncodeAddress(addr GAddress) string
}

type GeoCoder interface {
	AddressEncoder
	GeoCode(addr GAddress) Location
}

// TODO: refactor geo actor to not have address and only have inbox and recieve
// that way it can just be running and address can be passed to it

type GeoActor struct {
}

func NewGeoActor() GeoActor {
	return GeoActor{}
}

func (a *GeoActor) encodeAddress(addr GAddress) string {
	params := url.Values{}
	params.Add("address", fmt.Sprintf("%v, %v, %v", addr.Street, addr.City, addr.State))
	return params.Encode()
}

func (a *GeoActor) GeoCode(address GAddress) (*Location, error) {
	geoUrl, err := a.buildUrl()

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
	addrs := a.encodeAddress(addr)

	if key == "" || ep == "" {
		return "", errors.New("env variables are empty, need key and endpoint")
	}

	gUrl, err := url.JoinPath(ep, addrs)

	if err != nil {
		return "", err
	}

	param := url.Values{}
	param.Add("key", key)

	return url.JoinPath(gUrl, param.Encode())
}
