package core

import (
	"fmt"
	"sync"
)

// Location represents a geographic coordinate pair in decimal degrees.
type Location struct {
	Latitude  float64
	Longitude float64
}

var (
	geoMu  sync.RWMutex
	geoMap = make(map[NodeID]Location)
)

// RegisterLocation records the geolocation for a node in memory and, if a
// ledger is available, persists it there via SetNodeLocation.
func RegisterLocation(id NodeID, lat, lon float64) {
	geoMu.Lock()
	geoMap[id] = Location{Latitude: lat, Longitude: lon}
	geoMu.Unlock()
	if l := CurrentLedger(); l != nil {
		l.SetNodeLocation(id, Location{Latitude: lat, Longitude: lon})
	}
}

// GetLocation retrieves the stored location for a node.
func GetLocation(id NodeID) (Location, bool) {
	geoMu.RLock()
	loc, ok := geoMap[id]
	geoMu.RUnlock()
	return loc, ok
}

// ListLocations returns a copy of all tracked node locations.
func ListLocations() map[NodeID]Location {
	geoMu.RLock()
	out := make(map[NodeID]Location, len(geoMap))
	for id, loc := range geoMap {
		out[id] = loc
	}
	geoMu.RUnlock()
	return out
}

// NodesInRadius returns all node IDs within the given rectangular bounding box
// defined by min/max latitude and longitude. It is a simple filter and does not
// perform complex geospatial calculations.
func NodesInRadius(minLat, maxLat, minLon, maxLon float64) []NodeID {
	geoMu.RLock()
	defer geoMu.RUnlock()
	var ids []NodeID
	for id, loc := range geoMap {
		if loc.Latitude >= minLat && loc.Latitude <= maxLat &&
			loc.Longitude >= minLon && loc.Longitude <= maxLon {
			ids = append(ids, id)
		}
	}
	return ids
}

// PrettyLocation formats a location for human readable output.
func PrettyLocation(loc Location) string {
	return fmt.Sprintf("%.5f,%.5f", loc.Latitude, loc.Longitude)
}
