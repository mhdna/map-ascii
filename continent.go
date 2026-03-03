package mapascii

import (
	"fmt"
	"strings"
)

type Continent string

const (
	ContinentAfrica       Continent = "africa"
	ContinentAntarctica   Continent = "antarctica"
	ContinentAsia         Continent = "asia"
	ContinentEurope       Continent = "europe"
	ContinentNorthAmerica Continent = "north-america"
	ContinentSouthAmerica Continent = "south-america"
	ContinentOceania      Continent = "oceania"
)

const ContinentAliasAustralia = "australia"

var continentOrder = []Continent{
	ContinentAfrica,
	ContinentAntarctica,
	ContinentAsia,
	ContinentEurope,
	ContinentNorthAmerica,
	ContinentSouthAmerica,
	ContinentOceania,
}

var continentViewports = map[Continent]Viewport{
	ContinentAfrica: {
		MinLon: -20.0,
		MinLat: -36.0,
		MaxLon: 53.0,
		MaxLat: 38.0,
	},
	ContinentAntarctica: {
		MinLon: -180.0,
		MinLat: -90.0,
		MaxLon: 180.0,
		MaxLat: -55.0,
	},
	ContinentAsia: {
		MinLon: 25.0,
		MinLat: -10.0,
		MaxLon: 180.0,
		MaxLat: 82.0,
	},
	ContinentEurope: {
		MinLon: -25.0,
		MinLat: 34.0,
		MaxLon: 45.0,
		MaxLat: 72.0,
	},
	ContinentNorthAmerica: {
		MinLon: -170.0,
		MinLat: 5.0,
		MaxLon: -50.0,
		MaxLat: 83.0,
	},
	ContinentSouthAmerica: {
		MinLon: -92.0,
		MinLat: -56.0,
		MaxLon: -30.0,
		MaxLat: 15.0,
	},
	ContinentOceania: {
		MinLon: 110.0,
		MinLat: -50.0,
		MaxLon: 180.0,
		MaxLat: 10.0,
	},
}

func Continents() []Continent {
	out := make([]Continent, len(continentOrder))
	copy(out, continentOrder)
	return out
}

func ContinentNames() []string {
	names := make([]string, 0, len(continentOrder))
	for _, continent := range continentOrder {
		names = append(names, string(continent))
	}
	return names
}

func ContinentNamesCSV() string {
	return strings.Join(ContinentNames(), ", ")
}

func ParseContinent(raw string) (Continent, error) {
	normalized := normalizeContinentName(raw)
	if normalized == "" {
		return "", fmt.Errorf("continent must not be empty")
	}
	if normalized == ContinentAliasAustralia {
		return ContinentOceania, nil
	}

	continent := Continent(normalized)
	if _, ok := continentViewports[continent]; !ok {
		return "", fmt.Errorf("continent must be one of: %s", ContinentNamesCSV())
	}

	return continent, nil
}

func (c Continent) Viewport() (Viewport, error) {
	viewport, ok := continentViewports[c]
	if !ok {
		return Viewport{}, fmt.Errorf("continent must be one of: %s", ContinentNamesCSV())
	}

	return viewport, nil
}

func ViewportForContinent(raw string) (Viewport, error) {
	continent, err := ParseContinent(raw)
	if err != nil {
		return Viewport{}, err
	}

	return continent.Viewport()
}

func normalizeContinentName(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, "_", "-")
	normalized = strings.ReplaceAll(normalized, " ", "-")
	return normalized
}
