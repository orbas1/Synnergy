/// Simple time locked wallet used for demonstration purposes.
/// Funds can only be withdrawn after the `unlock_timestamp` has been
/// reached.  This is a synchronous, in-memory model and does not interact
/// with the Synnergy VM.
pub struct TimeLockedWallet {
    pub balance: u64,
    pub unlock_timestamp: u64,
}

impl TimeLockedWallet {
    /// Create a wallet locked until `unlock_timestamp`.
    pub fn new(unlock_timestamp: u64) -> Self {
        Self {
            balance: 0,
            unlock_timestamp,
        }
    }

    /// Deposit `amount` into the wallet.
    pub fn deposit(&mut self, amount: u64) {
        self.balance += amount;
    }

    /// Withdraw the full balance if the lock has expired.
    pub fn withdraw(&mut self, current_timestamp: u64) -> Option<u64> {
        if current_timestamp < self.unlock_timestamp {
            return None;
        }
        let amount = self.balance;
        self.balance = 0;
        Some(amount)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn cannot_withdraw_before_unlock() {
        let mut w = TimeLockedWallet::new(100);
        w.deposit(50);
        assert!(w.withdraw(99).is_none());
        assert_eq!(w.withdraw(100), Some(50));
    }
}
