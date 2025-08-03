use std::collections::HashMap;

/// Tracks pension contributions for participants and allows withdrawals
/// once a predefined release epoch has been reached.
pub struct PensionPlan {
    balances: HashMap<String, u64>,
    release_epoch: u64,
}

impl PensionPlan {
    /// Create a new plan with a given `release_epoch`.
    pub fn new(release_epoch: u64) -> Self {
        Self {
            balances: HashMap::new(),
            release_epoch,
        }
    }

    /// Record a contribution for `participant`.
    pub fn contribute(&mut self, participant: impl Into<String>, amount: u64) {
        let entry = self.balances.entry(participant.into()).or_insert(0);
        *entry += amount;
    }

    /// Attempt to withdraw funds for `participant` at `current_epoch`.
    /// Returns the amount withdrawn, or `None` if the funds are still locked.
    pub fn withdraw(&mut self, participant: &str, current_epoch: u64) -> Option<u64> {
        if current_epoch < self.release_epoch {
            return None;
        }
        self.balances.remove(participant)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn contributions_release_after_epoch() {
        let mut plan = PensionPlan::new(10);
        plan.contribute("alice", 100);
        assert!(plan.withdraw("alice", 5).is_none());
        assert_eq!(plan.withdraw("alice", 10), Some(100));
    }
}
