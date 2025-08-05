use std::collections::HashMap;

/// Tracks which nodes are responsible for particular shards.
#[derive(Default)]
pub struct ShardingCoordinator {
    shards: HashMap<u32, Vec<String>>,
}

impl ShardingCoordinator {
    /// Create an empty coordinator.
    pub fn new() -> Self {
        Self {
            shards: HashMap::new(),
        }
    }

    /// Assign a `node` to a `shard`.
    pub fn assign_node(&mut self, shard: u32, node: impl Into<String>) {
        self.shards.entry(shard).or_default().push(node.into());
    }

    /// Retrieve the nodes assigned to `shard` if any.
    pub fn nodes_for(&self, shard: u32) -> Option<&[String]> {
        self.shards.get(&shard).map(|v| v.as_slice())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn assign_and_retrieve_nodes() {
        let mut coord = ShardingCoordinator::new();
        coord.assign_node(1, "n1");
        coord.assign_node(1, "n2");
        assert_eq!(coord.nodes_for(1).unwrap().len(), 2);
        assert!(coord.nodes_for(2).is_none());
    }
}
