# RUDP Lab

RUDP Lab is a learning-focused implementation of reliable transport over UDP.
The current protocol uses stop-and-wait delivery with sequence numbers,
acknowledgments, checksums, timeout-based retransmission, and simulated packet
loss.

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
| Date | 2026-08-19 |
| Git commit | f5ac09f |
| Protocol | Stop-and-wait |
| Operating system | Windows 11 |
| CPU | Intel Core i9 12900K |
| Go version | 1.26.4 |
| Payload per packet | 1,024 bytes |
| Retransmission timeout | 500 ms |
| Maximum attempts per packet | 5 |
| Receiver output | received.bin (disk) |

### Stop-and-wait baseline

One row per every trial. Loss is the configured probability on each
direction, not the combined probability that an entire data/ACK exchange fails.

| Payload (bytes) | Loss | Sender seed | Receiver seed | Elapsed (ms) | Packets | Retransmissions | Mean RTT (ms) | Goodput (B/s) | Hash match | Result |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|:---:|:---:|
| 10000 | 0 | 1 | 101 | 0.515 | 11 | 0 | 0.046818 | 19417476 | true | success |
| 10000 | 0 | 2 | 102 | 0.5195 | 11 | 0 | 0.047227 | 19249278 | true | success |
| 10000 | 0 | 3 | 103 | 0.5086 | 11 | 0 | 0.046236 | 19661817 | true | success |
| 10000 | 0 | 4 | 104 | 0.5086 | 11 | 0 | 0.046236 | 19661817 | true | success |
| 10000 | 0 | 5 | 105 | 1.0709 | 11 | 0 | 0.097354 | 9337940 | true | success |
| 10000 | 0.05 | 1 | 101 | 0.5089 | 11 | 0 | 0.046263 | 19650226 | true | success |
| 10000 | 0.05 | 2 | 102 | 1.0477 | 11 | 0 | 0.047481 | 9544717 | true | success |
| 10000 | 0.05 | 3 | 103 | 0.5161 | 11 | 0 | 0.046918 | 19376090 | true | success |
| 10000 | 0.05 | 4 | 104 | 1002.4446 | 11 | 2 | 0.136777 | 9976 | true | success |
| 10000 | 0.05 | 5 | 105 | 1001.1956 | 11 | 2 | 0.0513 | 9988 | true | success |
| 10000 | 0.10 | 1 | 101 | 1001.7262 | 11 | 2 | 0.056666 | 9983 | true | success |
| 10000 | 0.10 | 2 | 102 | 500.58 | 11 | 1 | 0 | 19977 | true | success |
| 10000 | 0.10 | 3 | 103 | 0.515 | 11 | 0 | 0.046818 | 19417476 | true | success |
| 10000 | 0.10 | 4 | 104 | 2002.5718 | 11 | 4 | 0 | 4994 | true | success |
| 10000 | 0.10 | 5 | 105 | 1001.12 | 11 | 2 | 0.05282 | 9989 | true | success |
| 50000 | 0 | 1 | 101 | 1.527 | 50 | 0 | 0.03054 | 32743942 | true | success |
| 50000 | 0 | 2 | 102 | 2.5756 | 50 | 0 | 0.051512 | 19412952 | true | success |
| 50000 | 0 | 3 | 103 | 2.0839 | 50 | 0 | 0.041678 | 23993474 | true | success |
| 50000 | 0 | 4 | 104 | 2.2511 | 50 | 0 | 0.034234 | 22211363 | true | success |
| 50000 | 0 | 5 | 105 | 1.5502 | 50 | 0 | 0.010234 | 32253903 | true | success |
| 50000 | 0.05 | 1 | 101 | 1508.3898 | 50 | 3 | 0.125568 | 33148 | true | success |
| 50000 | 0.05 | 2 | 102 | 3007.8079 | 50 | 6 | 0.10681 | 16623 | true | success |
| 50000 | 0.05 | 3 | 103 | 1003.1277 | 50 | 2 | 0.045841 | 49844 | true | success |
| 50000 | 0.05 | 4 | 104 | 4510.3388 | 50 | 9 | 0.113245 | 11086 | true | success |
| 50000 | 0.05 | 5 | 105 | 2004.124 | 50 | 4 | 0.064276 | 24949 | true | success |
| 50000 | 0.10 | 1 | 101 | 5507.8416 | 50 | 11 | 0.02761 | 9078 | true | success |
| 50000 | 0.10 | 2 | 102 | 7009.9013 | 50 | 14 | 0.027562 | 7133 | true | success |
| 50000 | 0.10 | 3 | 103 | 2004.5725 | 50 | 4 | 0.057736 | 24943 | true | success |
| 50000 | 0.10 | 4 | 104 | 7013.9172 | 50 | 14 | 0.12464 | 7129 | true | success |
| 50000 | 0.10 | 5 | 105 | 6508.6694 | 50 | 13 | 0.083369 | 7682 | true | success |
| 100000 | 0 | 1 | 101 | 3.2116 | 99 | 0 | 0.03244 | 31137128 | true | success |
| 100000 | 0 | 2 | 102 | 2.9741 | 99 | 0 | 0.030041 | 33623617 | true | success |
| 100000 | 0 | 3 | 103 | 3.0857 | 99 | 0 | 0.031168 | 32407557 | true | success |
| 100000 | 0 | 4 | 104 | 3.0758 | 99 | 0 | 0.025957 | 32511867 | true | success |
| 100000 | 0 | 5 | 105 | 3.0803 | 99 | 0 | 0.031114 | 32464370 | true | success |
| 100000 | 0.05 | 1 | 101 | 3020.3045 | 99 | 6 | 0.178055 | 33109 | true | success |
| 100000 | 0.05 | 2 | 102 | 7522.9981 | 99 | 15 | 0.059596 | 13293 | true | success |
| 100000 | 0.05 | 3 | 103 | 5511.0841 | 99 | 11 | 0.07502 | 18145 | true | success |
| 100000 | 0.05 | 4 | 104 | 6016.1385 | 99 | 12 | 0.099703 | 16622 | true | success |
| 100000 | 0.05 | 5 | 105 | 5010.4543 | 99 | 10 | 0.068527 | 19958 | true | success |
| 100000 | 0.10 | 1 | 101 | 12526.2599 | 99 | 25 | 0.125206 | 7983 | true | success |
| 100000 | 0.10 | 2 | 102 | 13523.9829 | 99 | 27 | 0.06598 | 7394 | true | success |
| 100000 | 0.10 | 3 | 103 | 12520.9087 | 99 | 25 | 0.094002 | 7987 | true | success |
| 100000 | 0.10 | 4 | 104 | 13019.795 | 99 | 26 | 0.08634 | 7681 | true | success |
| 100000 | 0.10 | 5 | 105 | 12027.3102 | 99 | 24 | 0.165241 | 8314 | true | success |

### Stop-and-wait summary

Summarize repeated trials after the raw baseline is complete.

| Payload (bytes) | Loss | Successful trials | Mean elapsed (ms) | Mean retransmissions | Mean RTT (ms) | Mean goodput (B/s) |
|---:|---:|---:|---:|---:|---:|---:|
| 10000 | 0 | 5 | 0.625 | 0 | 0.0568 | 17465666 |
| 10000 | 0.05 | 5 | 401.143 | 0.8 | 0.0657 | 9718199 |
| 10000 | 0.10 | 5 | 901.303 | 1.8 | 0.0313 | 3892484 |
| 50000 | 0 | 5 | 1.998 | 0 | 0.0336 | 26123127 |
| 50000 | 0.05 | 5 | 2406.758 | 4.8 | 0.0911 | 27130 |
| 50000 | 0.10 | 5 | 5608.98 | 11.2 | 0.0642 | 11193 |
| 100000 | 0 | 5 | 3.086 | 0 | 0.0301 | 32428908 |
| 100000 | 0.05 | 5 | 5416.196 | 10.8 | 0.0962 | 20225 |
| 100000 | 0.10 | 5 | 12723.651 | 25.4 | 0.1074 | 7872 |

### Go-Back-N comparison
