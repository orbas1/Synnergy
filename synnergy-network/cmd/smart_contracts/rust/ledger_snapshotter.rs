//! Maintains ledger snapshots.

/// Stores arbitrary textual representations of ledger state.
pub struct LedgerSnapshotter {
    snapshots: Vec<String>,
}

impl LedgerSnapshotter {
    /// Creates a new snapshotter.
    pub fn new() -> Self {
        Self {
            snapshots: Vec::new(),
        }
    }

    /// Stores a new snapshot.
    pub fn take_snapshot<S: Into<String>>(&mut self, data: S) {
        self.snapshots.push(data.into());
    }

    /// Returns the latest snapshot if available.
    pub fn latest(&self) -> Option<&str> {
        self.snapshots.last().map(|s| s.as_str())
    }

    /// Returns the number of snapshots held.
    pub fn count(&self) -> usize {
        self.snapshots.len()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn snapshots_work() {
        let mut s = LedgerSnapshotter::new();
        s.take_snapshot("state1");
        s.take_snapshot("state2");
        assert_eq!(s.latest(), Some("state2"));
        assert_eq!(s.count(), 2);
    }
}
