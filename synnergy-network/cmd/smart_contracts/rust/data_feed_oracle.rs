//! Data feed oracle contract stub.
//!
//! Supplies external data to the Synnergy network. The comprehensive opcode and gas logic is
//! maintained in Go, while this Rust implementation guarantees basic validation and
//! successful compilation.

#[derive(Default)]
pub struct DataFeedOracle;

impl DataFeedOracle {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn execute_opcode(&self, opcode: u8, gas: u64) -> Result<(), String> {
        if gas == 0 {
            return Err("insufficient gas".into());
        }
        let _ = opcode;
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn oracle_executes_with_gas() {
        let oracle = DataFeedOracle::new();
        assert!(oracle.execute_opcode(8, 9).is_ok());
    }

    #[test]
    fn oracle_rejects_zero_gas() {
        let oracle = DataFeedOracle::new();
        assert!(oracle.execute_opcode(8, 0).is_err());
    }
}
