//! Tracks environmentally certified entities.

use std::collections::HashSet;

/// Maintains a registry of certified entities.
pub struct GreenCertifier {
    certified: HashSet<String>,
}

impl GreenCertifier {
    /// Creates an empty registry.
    pub fn new() -> Self {
        Self {
            certified: HashSet::new(),
        }
    }

    /// Certifies the specified entity.
    pub fn certify<S: Into<String>>(&mut self, entity: S) {
        self.certified.insert(entity.into());
    }

    /// Revokes the certification for an entity.
    pub fn revoke(&mut self, entity: &str) {
        self.certified.remove(entity);
    }

    /// Returns `true` if the entity is certified.
    pub fn is_certified(&self, entity: &str) -> bool {
        self.certified.contains(entity)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn certification_flow() {
        let mut cert = GreenCertifier::new();
        cert.certify("farm");
        assert!(cert.is_certified("farm"));
        cert.revoke("farm");
        assert!(!cert.is_certified("farm"));
    }
}
