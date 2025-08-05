//! Simple fault tolerance checker.

/// Provides utility to verify whether the number of faulty nodes
/// is within the tolerated threshold.
pub struct FaultToleranceChecker {
    threshold: f64,
}

impl FaultToleranceChecker {
    /// Creates a new checker with the allowed faulty fraction.
    ///
    /// # Panics
    /// Panics if `threshold` is not between 0.0 and 1.0 inclusive.
    pub fn new(threshold: f64) -> Self {
        assert!(
            (0.0..=1.0).contains(&threshold),
            "threshold must be between 0 and 1"
        );
        Self { threshold }
    }

    /// Returns `true` if `faulty_nodes` is within the tolerated
    /// fraction of `total_nodes`.
    pub fn is_tolerated(&self, faulty_nodes: usize, total_nodes: usize) -> bool {
        if total_nodes == 0 {
            return true;
        }
        (faulty_nodes as f64) <= self.threshold * (total_nodes as f64)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn tolerates_within_threshold() {
        let checker = FaultToleranceChecker::new(0.33);
        assert!(checker.is_tolerated(1, 4));
        assert!(!checker.is_tolerated(2, 4));
    }
}
