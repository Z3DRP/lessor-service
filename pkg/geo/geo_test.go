package geo

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/joho/godotenv"
)

func setup() {
	err := godotenv.Load("../../.env")
	if err != nil {
		panic("failed to load env")
	}
}

func TestGeocodeNoEnvFail(t *testing.T) {
	for _, test := range []struct {
		name        string
		input       GAddress
		locExpected *Location
		errExpected error
	}{
		{
			"valid address no env",
			GAddress{Street: "1425 Rock hill road", City: "Wood River", State: "Illinois"},
			nil,
			errors.New("failed to build geocode url env variables are empty, need key and endpoint"),
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", test.name, test.input), func(t *testing.T) {
			actor := NewGeoActor()
			location, err := actor.GeoCode(test.input)
			if got, want := err.Error(), test.errExpected.Error(); got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}

			if got, want := location, test.locExpected; got != want {
				t.Fatalf("location got= %v location want- %v", got, want)
			}

		})
	}
}

func TestGeocode(t *testing.T) {
	setup()
	for _, test := range []struct {
		name        string
		input       GAddress
		locExpected Location
		errExpected error
	}{
		{
			"valid address",
			GAddress{Street: "1425 Rock hill road", City: "Wood River", State: "Illinois"},
			Location{Latitude: 38.861574, Longitude: -90.07448586421482},
			nil,
		},
	} {
		t.Run(fmt.Sprintf("%s [%s]", test.name, test.input), func(t *testing.T) {
			const epsilon = 0.0002
			actor := NewGeoActor()
			location, err := actor.GeoCode(test.input)
			if got, want := err, test.errExpected; got != want {
				t.Fatalf("err=%v, want=%v", got, want)
			}

			if got, want := location, test.locExpected; math.Abs(got.Latitude-want.Latitude) > epsilon {
				t.Fatalf("latitude mismatch got= %v lat want- %v", got.Latitude, want.Latitude)
			}

			if got, want := location, test.locExpected; math.Abs(got.Longitude-want.Longitude) > epsilon {
				t.Fatalf("longitude mismatch got= %v, lng want= %v", got.Longitude, want.Longitude)
			}
		})
	}
}
