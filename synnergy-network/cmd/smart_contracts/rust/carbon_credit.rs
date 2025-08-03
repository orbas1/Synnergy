//! Carbon credit contract stub.
//!
//! Provides a minimal Rust implementation of the carbon credit smart contract used in
//! Synnergy. Actual opcode dispatching and gas calculations occur within Go components. The
//! Rust code verifies gas usage and supports basic compilation.

#[derive(Default)]
pub struct CarbonCredit;

impl CarbonCredit {
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
    fn executes_with_gas() {
        let contract = CarbonCredit::new();
        assert!(contract.execute_opcode(3, 4).is_ok());
    }

    #[test]
    fn rejects_zero_gas() {
        let contract = CarbonCredit::new();
        assert!(contract.execute_opcode(3, 0).is_err());
    }
}
