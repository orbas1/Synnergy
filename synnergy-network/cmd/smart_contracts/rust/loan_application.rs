//! Simple loan application state machine.

/// Possible states for a loan application.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum LoanStatus {
    Pending,
    Approved,
    Rejected,
}

/// Represents a loan application.
pub struct LoanApplication {
    pub applicant: String,
    pub amount: u64,
    status: LoanStatus,
}

impl LoanApplication {
    /// Creates a new loan application in the `Pending` state.
    pub fn new<S: Into<String>>(applicant: S, amount: u64) -> Self {
        Self {
            applicant: applicant.into(),
            amount,
            status: LoanStatus::Pending,
        }
    }

    /// Returns the current status of the application.
    pub fn status(&self) -> LoanStatus {
        self.status
    }

    /// Marks the application as approved.
    pub fn approve(&mut self) {
        self.status = LoanStatus::Approved;
    }

    /// Marks the application as rejected.
    pub fn reject(&mut self) {
        self.status = LoanStatus::Rejected;
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn state_transitions() {
        let mut app = LoanApplication::new("bob", 1000);
        assert_eq!(app.status(), LoanStatus::Pending);
        app.approve();
        assert_eq!(app.status(), LoanStatus::Approved);
        app.reject();
        assert_eq!(app.status(), LoanStatus::Rejected);
    }
}
