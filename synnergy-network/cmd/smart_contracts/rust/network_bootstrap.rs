//! Tracks nodes participating in bootstrap phase.

use std::collections::HashSet;

/// Registry for nodes joining the network during bootstrap.
pub struct NetworkBootstrap {
    nodes: HashSet<String>,
}

impl NetworkBootstrap {
    /// Creates an empty registry.
    pub fn new() -> Self {
        Self {
            nodes: HashSet::new(),
        }
    }

    /// Adds a node identifier. Returns `true` if the node was newly inserted.
    pub fn add_node<S: Into<String>>(&mut self, id: S) -> bool {
        self.nodes.insert(id.into())
    }

    /// Returns `true` if the node identifier exists in the registry.
    pub fn has_node(&self, id: &str) -> bool {
        self.nodes.contains(id)
    }

    /// Returns the number of registered nodes.
    pub fn node_count(&self) -> usize {
        self.nodes.len()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn add_and_check_nodes() {
        let mut nb = NetworkBootstrap::new();
        assert!(nb.add_node("node1"));
        assert!(!nb.add_node("node1"));
        assert!(nb.has_node("node1"));
        assert_eq!(nb.node_count(), 1);
    }
}
