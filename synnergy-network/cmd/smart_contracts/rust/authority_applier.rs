//! Authority application contract stub.
//!
//! This module mirrors the behaviour of the authority applier smart contract within the
//! Synnergy network. Opcode execution and gas accounting are delegated to Go components
//! (`opcode_dispatcher.go` and `gas_table.go`). Here we provide a minimal Rust version for
//! compilation and basic validation.

#[derive(Default)]
pub struct AuthorityApplier;

impl AuthorityApplier {
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
    fn execute_with_gas() {
        let contract = AuthorityApplier::new();
        assert!(contract.execute_opcode(1, 10).is_ok());
    }

    #[test]
    fn execute_without_gas() {
        let contract = AuthorityApplier::new();
        assert!(contract.execute_opcode(1, 0).is_err());
    }
}
