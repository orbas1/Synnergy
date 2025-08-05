use std::collections::HashMap;

/// Simple in-memory oracle used for tests and demonstrations.
///
/// The real implementation would interface with the Synnergy VM via
/// custom opcodes defined in `opcode_dispatcher.go`.  For the purposes of
/// this repository we provide a lightweight, fully synchronous Rust
/// implementation that allows unit tests to exercise basic behaviour.
#[derive(Default)]
pub struct OracleUpdater {
    values: HashMap<String, i64>,
}

impl OracleUpdater {
    /// Create a new [`OracleUpdater`].
    pub fn new() -> Self {
        Self {
            values: HashMap::new(),
        }
    }

    /// Store a new oracle value for the given `key`.
    pub fn update(&mut self, key: impl Into<String>, value: i64) {
        self.values.insert(key.into(), value);
    }

    /// Fetch the most recent value for `key` if present.
    pub fn get(&self, key: &str) -> Option<i64> {
        self.values.get(key).copied()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn update_and_get() {
        let mut oracle = OracleUpdater::new();
        oracle.update("eth_usd", 3_200);
        assert_eq!(oracle.get("eth_usd"), Some(3_200));
        assert!(oracle.get("btc_usd").is_none());
    }
}
