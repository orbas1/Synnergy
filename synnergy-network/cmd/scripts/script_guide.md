# Synnergy Script Guide

The `start_synnergy_network.sh` script builds the `synnergy` CLI and boots the
core daemons using the commands provided by the CLI packages.  It also runs a
sample security command.

```bash
./start_synnergy_network.sh
```

This will:

1. Compile the CLI from `cmd/synnergy/main.go`.
2. Start the networking, consensus, replication and VM services.
3. Run a demo `~sec merkle` command.
4. Wait until the services exit (Ctrl+C to terminate).

Ensure that the Go toolchain is available in your `PATH` before running the
script.
