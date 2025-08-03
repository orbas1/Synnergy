//! Bank node bridge contract stub.
//!
//! In the complete Synnergy network this contract enables secure communication between
//! bank nodes. Opcode dispatching and gas metering are handled externally in Go. This file
//! offers a lightweight Rust version to guarantee compilation and provide basic gas
//! validation.

#[derive(Default)]
pub struct BankNodeBridge;

impl BankNodeBridge {
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
    fn bridge_with_gas() {
        let bridge = BankNodeBridge::new();
        assert!(bridge.execute_opcode(2, 5).is_ok());
    }

    #[test]
    fn bridge_without_gas() {
        let bridge = BankNodeBridge::new();
        assert!(bridge.execute_opcode(2, 0).is_err());
    }
}
