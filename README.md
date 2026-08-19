# RUDP Lab

RUDP Lab is a learning-focused implementation of reliable transport over UDP.
The current protocol uses stop-and-wait delivery with sequence numbers,
acknowledgments, checksums, timeout-based retransmission, and simulated packet
loss. A Go-Back-N sliding window will be added after the stop-and-wait baseline
is recorded.

## Benchmarks

### Measurement notes

- **Goodput** is delivered application payload divided by transfer time. It
  excludes protocol headers and retransmitted bytes.
- **Throughput** is the total amount of data transmitted over time, including
  protocol overhead and retransmissions. RUDP Lab does not currently report it.
- Packet loss is applied independently to outgoing data packets and outgoing
  acknowledgments.
- Sender timing covers data transmission through FIN acknowledgment. Address
  resolution, socket setup, and result logging are excluded.
- RTT samples from retransmitted packets are excluded, following Karn's
  algorithm.

### Test environment

Record the environment once for each benchmark session.

| Field | Value |
|---|---|
| Date | |
| Git commit | |
| Protocol | Stop-and-wait |
| Operating system | |
| CPU | |
| Go version | |
| Payload per packet | 1,024 bytes |
| Retransmission timeout | 500 ms |
| Maximum attempts per packet | 5 |
| Receiver output | |

### Stop-and-wait baseline

Add one row for every trial. Loss is the configured probability on each
direction, not the combined probability that an entire data/ACK exchange fails.

| Payload (bytes) | Loss | Sender seed | Receiver seed | Elapsed (ms) | Packets | Retransmissions | Mean RTT (ms) | Goodput (B/s) | Hash match | Result |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|:---:|:---:|

### Stop-and-wait summary

Summarize repeated trials after the raw baseline is complete.

| Payload (bytes) | Loss | Successful trials | Mean elapsed (ms) | Mean retransmissions | Mean RTT (ms) | Mean goodput (B/s) |
|---:|---:|---:|---:|---:|---:|---:|

### Go-Back-N comparison

Run the same payload sizes, loss rates, and seed pairs after Go-Back-N is
implemented. Keep the stop-and-wait data above unchanged so the comparison uses
the original baseline.

