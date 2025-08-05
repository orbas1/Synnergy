use std::collections::HashMap;

/// Minimal replication service that stores a copy of data for each node.
#[derive(Default)]
pub struct ReplicationService {
    nodes: Vec<String>,
    storage: HashMap<String, Vec<u8>>,
}

impl ReplicationService {
    /// Create a new service replicating across `nodes`.
    pub fn new(nodes: Vec<String>) -> Self {
        Self {
            nodes,
            storage: HashMap::new(),
        }
    }

    /// Replicate `data` for the given `id` across all nodes.
    pub fn replicate(&mut self, id: impl Into<String>, data: Vec<u8>) {
        let id = id.into();
        for node in &self.nodes {
            let key = format!("{id}@{node}");
            self.storage.insert(key, data.clone());
        }
    }

    /// Number of stored replicas.
    pub fn replicas(&self) -> usize {
        self.storage.len()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn replicate_to_all_nodes() {
        let nodes = vec!["n1".into(), "n2".into()];
        let mut svc = ReplicationService::new(nodes);
        svc.replicate("file1", vec![1, 2, 3]);
        assert_eq!(svc.replicas(), 2);
    }
}
