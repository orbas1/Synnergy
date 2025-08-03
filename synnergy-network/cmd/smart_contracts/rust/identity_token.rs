//! Identity token registry.

use std::collections::HashMap;
use std::time::{SystemTime, UNIX_EPOCH};

#[derive(Clone)]
pub struct Token {
    pub owner: String,
    pub issued: u64,
}

/// Maintains issued identity tokens.
pub struct IdentityToken {
    tokens: HashMap<String, Token>,
}

impl IdentityToken {
    /// Creates an empty registry.
    pub fn new() -> Self {
        Self {
            tokens: HashMap::new(),
        }
    }

    /// Issues a new token for an `id` and `owner`.
    pub fn issue<S: Into<String>>(&mut self, id: S, owner: S) {
        let now = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_secs();
        self.tokens.insert(
            id.into(),
            Token {
                owner: owner.into(),
                issued: now,
            },
        );
    }

    /// Revokes the token for the provided `id`.
    pub fn revoke(&mut self, id: &str) {
        self.tokens.remove(id);
    }

    /// Returns the owner for a token if it exists.
    pub fn owner_of(&self, id: &str) -> Option<&str> {
        self.tokens.get(id).map(|t| t.owner.as_str())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn issue_and_revoke() {
        let mut registry = IdentityToken::new();
        registry.issue("id1", "alice");
        assert_eq!(registry.owner_of("id1"), Some("alice"));
        registry.revoke("id1");
        assert_eq!(registry.owner_of("id1"), None);
    }
}
