# Synnergy Configuration Guide

Configuration files are located in `cmd/config/`.  `default.yaml` is always
loaded first and provides sane defaults for development.  You can override these
values by creating additional yaml files and setting the `SYNN_ENV` environment
variable to the basename of the file (for example `prod` to load `prod.yaml`).

Every file follows the structure described below:

```yaml
network:
  id: synnergy-mainnet
  chain_id: 1215
  max_peers: 50
  genesis_file: config/genesis.json
  rpc_enabled: true
  p2p_port: 30303
  listen_addr: "/ip4/0.0.0.0/tcp/4001"
  discovery_tag: synnergy-mesh
  bootstrap_peers:
    - "/ip4/127.0.0.1/tcp/4001/p2p/QmPeerID"

consensus:
  type: pos
  block_time_ms: 3000
  validators_required: 3

vm:
  max_gas_per_block: 8000000
  opcode_debug: false

storage:
  db_path: ./data/db
  prune: true

logging:
  level: info
  file: logs/synnergy.log
```

`bootstrap.yaml` is a template you can copy when running a dedicated bootstrap
node.  It is identical to `default.yaml` but sets a different `discovery_tag`
and usually runs with a higher peer limit and no `bootstrap_peers` configured.
