use std::time::{SystemTime, UNIX_EPOCH};

/// Maintains a log of security events for auditing purposes.
#[derive(Default)]
pub struct SecurityAuditor {
    logs: Vec<AuditLog>,
}

/// A single security audit log entry.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct AuditLog {
    pub timestamp: u64,
    pub message: String,
}

impl SecurityAuditor {
    /// Create an empty [`SecurityAuditor`].
    pub fn new() -> Self {
        Self { logs: Vec::new() }
    }

    /// Record a new audit `message` with the current timestamp.
    pub fn log(&mut self, message: impl Into<String>) {
        let ts = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs();
        self.logs.push(AuditLog {
            timestamp: ts,
            message: message.into(),
        });
    }

    /// Return all audit logs.
    pub fn logs(&self) -> &[AuditLog] {
        &self.logs
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn log_records_message() {
        let mut auditor = SecurityAuditor::new();
        auditor.log("issue");
        assert_eq!(auditor.logs().len(), 1);
    }
}
