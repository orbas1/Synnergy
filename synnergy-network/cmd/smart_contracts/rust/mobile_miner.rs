//! Mobile miner stub.

/// Represents a miner operating from a mobile device.
pub struct MobileMiner {
    pub hash_rate: u64,
    mined_blocks: u64,
}

impl MobileMiner {
    /// Creates a new miner with the specified hash rate.
    pub fn new(hash_rate: u64) -> Self {
        Self {
            hash_rate,
            mined_blocks: 0,
        }
    }

    /// Records the mining of a block.
    pub fn mine_block(&mut self) {
        self.mined_blocks += 1;
    }

    /// Returns the total number of mined blocks.
    pub fn mined_blocks(&self) -> u64 {
        self.mined_blocks
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn mining_increments() {
        let mut miner = MobileMiner::new(10);
        miner.mine_block();
        miner.mine_block();
        assert_eq!(miner.mined_blocks(), 2);
    }
}
