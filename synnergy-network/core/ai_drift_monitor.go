package core

// ai_drift_monitor.go - simple sliding window drift detector for AI models.

import "sync"

// DriftMonitor tracks average scores and signals when deviation exceeds threshold.
type DriftMonitor struct {
	mu     sync.Mutex
	window int
	data   map[[32]byte][]float32
	base   map[[32]byte]float32
}

// NewDriftMonitor returns a monitor with the provided window size.
func NewDriftMonitor(window int) *DriftMonitor {
	return &DriftMonitor{
		window: window,
		data:   make(map[[32]byte][]float32),
		base:   make(map[[32]byte]float32),
	}
}

// Record adds a new score for the model and calculates drift.
// It returns the drift value and whether it exceeds 20% of the baseline.
func (d *DriftMonitor) Record(model [32]byte, score float32) (float64, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	arr := d.data[model]
	arr = append(arr, score)
	if len(arr) > d.window {
		arr = arr[1:]
	}
	d.data[model] = arr
	var sum float64
	for _, s := range arr {
		sum += float64(s)
	}
	avg := sum / float64(len(arr))
	if _, ok := d.base[model]; !ok {
		d.base[model] = float32(avg)
		return 0, false
	}
	drift := avg - float64(d.base[model])
	if abs(drift)/float64(d.base[model]) > 0.2 {
		return drift, true
	}
	return drift, false
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
