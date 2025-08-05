//! Governance quorum utility.

/// Determines whether a vote meets a required quorum.
pub struct GovQuorum {
    required_percentage: f64,
}

impl GovQuorum {
    /// Creates a new quorum with a required percentage between 0.0 and 1.0.
    pub fn new(required_percentage: f64) -> Self {
        assert!((0.0..=1.0).contains(&required_percentage));
        Self {
            required_percentage,
        }
    }

    /// Returns `true` if the ratio of `votes_for` to `total_votes` meets the requirement.
    pub fn has_quorum(&self, votes_for: u64, total_votes: u64) -> bool {
        if total_votes == 0 {
            return false;
        }
        (votes_for as f64) / (total_votes as f64) >= self.required_percentage
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn quorum_check() {
        let q = GovQuorum::new(0.6);
        assert!(q.has_quorum(6, 10));
        assert!(!q.has_quorum(5, 10));
    }
}
