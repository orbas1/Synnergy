//! Cross-chain protocol contract stub.
//!
//! Provides the structural foundation for cross-chain communication within Synnergy. The
//! full opcode execution is delegated to Go infrastructure; here we only handle gas checks
//! and ensure compilation.

#[derive(Default)]
pub struct CrossChainProtocol;

impl CrossChainProtocol {
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
    fn protocol_works_with_gas() {
        let proto = CrossChainProtocol::new();
        assert!(proto.execute_opcode(7, 2).is_ok());
    }

    #[test]
    fn protocol_fails_without_gas() {
        let proto = CrossChainProtocol::new();
        assert!(proto.execute_opcode(7, 0).is_err());
    }
}
