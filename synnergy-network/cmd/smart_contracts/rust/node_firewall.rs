//! Basic node firewall allowing specific IP addresses.

use std::collections::HashSet;
use std::net::IpAddr;

/// Maintains a set of allowed IP addresses for a node.
pub struct NodeFirewall {
    allowed: HashSet<IpAddr>,
}

impl NodeFirewall {
    /// Creates a new firewall with no allowed addresses.
    pub fn new() -> Self {
        Self {
            allowed: HashSet::new(),
        }
    }

    /// Allows the given IP address.
    pub fn allow(&mut self, ip: IpAddr) {
        self.allowed.insert(ip);
    }

    /// Revokes the given IP address.
    pub fn revoke(&mut self, ip: &IpAddr) {
        self.allowed.remove(ip);
    }

    /// Returns `true` if the IP address is allowed.
    pub fn is_allowed(&self, ip: &IpAddr) -> bool {
        self.allowed.contains(ip)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::net::{IpAddr, Ipv4Addr};

    #[test]
    fn allows_and_revokes() {
        let mut fw = NodeFirewall::new();
        let ip = IpAddr::V4(Ipv4Addr::new(127, 0, 0, 1));
        fw.allow(ip);
        assert!(fw.is_allowed(&ip));
        fw.revoke(&ip);
        assert!(!fw.is_allowed(&ip));
    }
}
