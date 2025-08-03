//! Simple health monitor tracking metrics.

use std::collections::HashMap;

/// Maintains a set of health metrics.
pub struct HealthMonitor {
    metrics: HashMap<String, f64>,
}

impl HealthMonitor {
    /// Creates a new `HealthMonitor` with no metrics.
    pub fn new() -> Self {
        Self {
            metrics: HashMap::new(),
        }
    }

    /// Updates or inserts a metric with the provided value.
    pub fn update<S: Into<String>>(&mut self, metric: S, value: f64) {
        self.metrics.insert(metric.into(), value);
    }

    /// Returns the value of a metric if it exists.
    pub fn value(&self, metric: &str) -> Option<f64> {
        self.metrics.get(metric).copied()
    }

    /// Returns `true` if the metric exists and is within the range `[min, max]`.
    pub fn within(&self, metric: &str, min: f64, max: f64) -> bool {
        self.metrics
            .get(metric)
            .map_or(false, |v| *v >= min && *v <= max)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn monitor_metrics() {
        let mut m = HealthMonitor::new();
        m.update("cpu", 0.5);
        assert!(m.within("cpu", 0.0, 0.8));
        assert!(!m.within("cpu", 0.0, 0.4));
        assert_eq!(m.value("cpu"), Some(0.5));
    }
}
