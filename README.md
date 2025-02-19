# Go Ethereum

**This is a personal repository for learning purposes only.**

## General Summary

I've chosen a series of EIPs as hands‑on projects to explore different aspects of the Ethereum protocol and the Geth client. Each EIP targets a distinct area of the system—ranging from opcode extensions to precompiled contract integration. Working on these has given me practical insights into EVM internals, state management, and protocol upgrade mechanisms. As I continue, I'll add more entries to this list.

**Note:** Each implemented EIP is maintained on its own branch.

---

### EIP‑7212: Precompiled Contract for secp256r1 Verification

- **Purpose:**  
  Adds native support for verifying ECDSA signatures using the secp256r1 (P‑256) curve via a precompile.

- **What I Did:**  
  I implemented the `p256Verify` precompile, which parses a fixed 160-byte input (message hash, r, s, public key x and y), validates the signature using Go’s `crypto/ecdsa`, and integrates it into a simulated fork called "Septimus" to see how upgrades are handled.

---

### EIP‑3855: PUSH0 Opcode

- **Purpose:**  
  Introduces a new opcode that pushes a zero value onto the stack without requiring any immediate data, streamlining contract bytecode and reducing gas costs.

- **What I Did:**  
  I implemented the `PUSH0` opcode by adding its definition to the opcode table and writing the execution logic (`opPush0`), then verified it through tests and benchmarks.

---

### EIP‑7881: Increase Gas Utilization Target

- **Purpose:**  
  Increase the gas utilization target from 50% to 75%.

- **What I Did:**  
  I updated the gas target calculations and base fee update rules according to the new EIP specification. This involved:

    - Introducing new elasticity constants (`DefaultGasTargetNumerator` and `DefaultGasTargetDenominator`) to compute a 75% target while keeping the maximum gas limit unchanged.
    - Modifying the base fee update formula with piecewise slopes (1/8 for below target and 3/8 for above target) to ensure the base fee changes remain within ±12.5%.
