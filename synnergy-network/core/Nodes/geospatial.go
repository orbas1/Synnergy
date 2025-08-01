package Nodes

// GeospatialNodeInterface extends NodeInterface with geospatial capabilities.
type GeospatialNodeInterface interface {
	NodeInterface
	RegisterGeoData(id string, lat, lon float64) error
	TransformCoordinates(lat, lon float64, to string) (float64, float64, error)
	AddGeofence(name string, polygon [][2]float64) error
	InGeofence(name string, lat, lon float64) bool
	QueryRegion(minLat, maxLat, minLon, maxLon float64) []string
}
