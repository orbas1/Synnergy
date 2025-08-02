package core

import (
	"math"
	"sync"
)

// GeoRecord stores a geospatial data point.
type GeoRecord struct {
	ID       string
	Location Location
	Meta     map[string]string
}

// Geofence represents a polygon area used for geofencing.
type Geofence struct {
	Name   string
	Points []Location
}

// GeospatialNode provides geospatial processing and networking capabilities.
type GeospatialNode struct {
	*Node
	ledger  *Ledger
	mu      sync.RWMutex
	records map[string]GeoRecord
	fences  map[string]Geofence
}

// NewGeospatialNode creates a network node with ledger support for geospatial data.
func NewGeospatialNode(netCfg Config, ledCfg LedgerConfig) (*GeospatialNode, error) {
	n, err := NewNode(netCfg)
	if err != nil {
		return nil, err
	}
	led, err := NewLedger(ledCfg)
	if err != nil {
		_ = n.Close()
		return nil, err
	}
	g := &GeospatialNode{
		Node:    n,
		ledger:  led,
		records: make(map[string]GeoRecord),
		fences:  make(map[string]Geofence),
	}
	return g, nil
}

// RegisterGeoData stores a geospatial record and persists the location on the ledger.
func (g *GeospatialNode) RegisterGeoData(id string, lat, lon float64) error {
	g.mu.Lock()
	g.records[id] = GeoRecord{ID: id, Location: Location{Latitude: lat, Longitude: lon}}
	g.mu.Unlock()
	if g.ledger != nil {
		g.ledger.SetNodeLocation(NodeID(id), Location{Latitude: lat, Longitude: lon})
	}
	return nil
}

// TransformCoordinates converts between supported coordinate systems.
// Currently only "radians" is supported.
func (g *GeospatialNode) TransformCoordinates(lat, lon float64, to string) (float64, float64, error) {
	switch to {
	case "radians":
		return lat * math.Pi / 180, lon * math.Pi / 180, nil
	default:
		return lat, lon, nil
	}
}

// AddGeofence registers a polygon used for geofencing.
func (g *GeospatialNode) AddGeofence(name string, polygon [][2]float64) error {
	pts := make([]Location, len(polygon))
	for i, p := range polygon {
		pts[i] = Location{Latitude: p[0], Longitude: p[1]}
	}
	g.mu.Lock()
	g.fences[name] = Geofence{Name: name, Points: pts}
	g.mu.Unlock()
	return nil
}

// InGeofence checks whether the given coordinate is inside a named geofence.
func (g *GeospatialNode) InGeofence(name string, lat, lon float64) bool {
	g.mu.RLock()
	fence, ok := g.fences[name]
	g.mu.RUnlock()
	if !ok {
		return false
	}
	return pointInPolygon(Location{Latitude: lat, Longitude: lon}, fence.Points)
}

// QueryRegion returns the IDs of all records within the bounding box.
func (g *GeospatialNode) QueryRegion(minLat, maxLat, minLon, maxLon float64) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var ids []string
	for id, rec := range g.records {
		if rec.Location.Latitude >= minLat && rec.Location.Latitude <= maxLat &&
			rec.Location.Longitude >= minLon && rec.Location.Longitude <= maxLon {
			ids = append(ids, id)
		}
	}
	return ids
}

// Close shuts down the node and ledger.
func (g *GeospatialNode) Close() error {
	if g.ledger != nil {
		_ = g.ledger.Close()
	}
	return g.Node.Close()
}

// pointInPolygon returns true if p is inside the polygon defined by pts.
func pointInPolygon(p Location, pts []Location) bool {
	inside := false
	j := len(pts) - 1
	for i := 0; i < len(pts); i++ {
		xi, yi := pts[i].Longitude, pts[i].Latitude
		xj, yj := pts[j].Longitude, pts[j].Latitude
		intersect := ((yi > p.Latitude) != (yj > p.Latitude)) &&
			(p.Longitude < (xj-xi)*(p.Latitude-yi)/(yj-yi)+xi)
		if intersect {
			inside = !inside
		}
		j = i
	}
	return inside
}
