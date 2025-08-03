# Authority Node Type Manual

Authority nodes oversee sensitive operations within the Synnergy network. Every
authority node **must** register with a dedicated wallet address used for fee
distribution and grant payouts. Registration without a wallet is rejected.

Upon acceptance, the network releases a unique 32‑byte **job key** to the
authority. This secret unlocks encrypted job assignments stored in the node’s
local keystore. Jobs are distributed randomly to ensure decentralised voting
and workload balance.

## Roles and Capabilities

- **GovernmentNode** – may issue benefit tokens and apply monetary or fiscal
  policy to SYN‑10/11/12 tokens.
- **CentralBankNode** – the only role permitted to deploy the SYN‑10/11/12 token
  standard.
- **RegulatorNode** – can propose security upgrades which must then pass a
  community vote.
- **CreditorBankNode** – authorised to originate bill tokens and other regulated
  financial instruments.
- **StandardAuthorityNode** – participates in governance and validation but
  cannot perform the specialised actions above.

Disbursed grants pay participating authority nodes a 5% fee and require approval
from at least five authority nodes plus a broader set of normal nodes. All
authority nodes except the elected authority may validate ID tokens; identities
remain invalid until verification to prevent double voting.


Reversal transactions require signatures from at least three active authority
nodes. Their wallets receive proportional fees where applicable. Authority
approval is also mandatory for LoanPool proposals and other governance actions
to maintain decentralisation thresholds.
