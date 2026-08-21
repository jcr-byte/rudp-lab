# RUDP Lab

RUDP Lab is a learning-focused implementation of reliable transport over UDP.
The current protocol uses a Go-Back-N sliding window with sequence numbers,
cumulative acknowledgments, checksums, timeout-based retransmission, and
simulated packet loss. Setting the window size to 1 provides stop-and-wait
behavior for comparison.

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

### Window-size comparison

The controlled benchmark compares stop-and-wait behavior (window 1) with
Go-Back-N (window 4) using the same implementation and test conditions.

#### Test environment

| Field | Value |
|---|---|
| Date | 2026-08-21 |
| Implementation | dc90f77 |
| Operating system | Windows 11 |
| CPU | Intel Core i9 12900K |
| Go version | 1.26.4 |
| Payload per packet | 1,024 bytes |
| Transfer sizes | 100,000; 500,000; 1,000,000 bytes |
| Window sizes | 1 and 4 |
| Retransmission timeout | 100 ms |
| Maximum attempts without progress | 5 |
| Trials per configuration | 5 |
| Receiver output | received.bin (disk) |

#### Results

Window 4 delivered 2.02x-2.66x more aggregate goodput than window 1. It also
completed all 45 trials, while window 1 exhausted its retry limit three times
at 10% loss. All completed transfers produced matching SHA-256 hashes.

| Payload (bytes) | Loss | Window-1 successes | Window-1 failures | Window-1 goodput (B/s) | Window-4 successes | Window-4 failures | Window-4 goodput (B/s) | Improvement |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 100000 | 0 | 5 | 0 | 28196973 | 5 | 0 | 74841336 | 2.65x |
| 100000 | 0.05 | 5 | 0 | 91798 | 5 | 0 | 234808 | 2.56x |
| 100000 | 0.1 | 5 | 0 | 39134 | 5 | 0 | 86998 | 2.22x |
| 500000 | 0 | 5 | 0 | 32821967 | 5 | 0 | 87175446 | 2.66x |
| 500000 | 0.05 | 5 | 0 | 79287 | 5 | 0 | 166897 | 2.10x |
| 500000 | 0.1 | 4 | 1 | 40496 | 5 | 0 | 84911 | 2.10x |
| 1000000 | 0 | 5 | 0 | 34329012 | 5 | 0 | 86705303 | 2.53x |
| 1000000 | 0.05 | 5 | 0 | 88441 | 5 | 0 | 188470 | 2.13x |
| 1000000 | 0.1 | 3 | 2 | 44109 | 5 | 0 | 89111 | 2.02x |

Aggregate goodput uses successful trials only. This favors window 1 in groups
where its failed transfers are excluded, so failure counts remain important.

<details>
<summary>Raw window-4 trials</summary>

Loss is independently applied to outgoing data packets and acknowledgments.
`Packets` counts original packets; retransmissions are separate.

| Commit | Window | Payload (bytes) | Loss | Sender seed | Receiver seed | Timeout (ms) | Elapsed (ms) | Packets | Retransmissions | RTT samples | Mean RTT (ms) | Goodput (B/s) | Hash match | Result |
|:---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|:---:|:---:|
| dc90f77 | 4 | 100000 | 0 | 1 | 101 | 100 | 1.5414 | 99 | 0 | 99 | 0.062278 | 64876087 | true | success |
| dc90f77 | 4 | 100000 | 0 | 2 | 102 | 100 | 1.5345 | 99 | 0 | 99 | 0.03608 | 65167807 | true | success |
| dc90f77 | 4 | 100000 | 0 | 3 | 103 | 100 | 1.0282 | 99 | 0 | 99 | 0.041543 | 97257343 | true | success |
| dc90f77 | 4 | 100000 | 0 | 4 | 104 | 100 | 1.561 | 99 | 0 | 99 | 0.06307 | 64061499 | true | success |
| dc90f77 | 4 | 100000 | 0 | 5 | 105 | 100 | 1.0157 | 99 | 0 | 99 | 0.041038 | 98454268 | true | success |
| dc90f77 | 4 | 100000 | 0.05 | 1 | 101 | 100 | 107.69 | 99 | 4 | 90 | 0.218655 | 928591 | true | success |
| dc90f77 | 4 | 100000 | 0.05 | 2 | 102 | 100 | 607.1736 | 99 | 23 | 72 | 0.074833 | 164698 | true | success |
| dc90f77 | 4 | 100000 | 0.05 | 3 | 103 | 100 | 704.7746 | 99 | 28 | 71 | 0.021743 | 141889 | true | success |
| dc90f77 | 4 | 100000 | 0.05 | 4 | 104 | 100 | 505.5015 | 99 | 20 | 75 | 0.10877 | 197823 | true | success |
| dc90f77 | 4 | 100000 | 0.05 | 5 | 105 | 100 | 204.2562 | 99 | 8 | 86 | 0.07388 | 489581 | true | success |
| dc90f77 | 4 | 100000 | 0.1 | 1 | 101 | 100 | 1006.1756 | 99 | 39 | 61 | 0.045226 | 99386 | true | success |
| dc90f77 | 4 | 100000 | 0.1 | 2 | 102 | 100 | 1410.5458 | 99 | 50 | 55 | 0.02033 | 70895 | true | success |
| dc90f77 | 4 | 100000 | 0.1 | 3 | 103 | 100 | 1417.0449 | 99 | 55 | 56 | 0.128387 | 70569 | true | success |
| dc90f77 | 4 | 100000 | 0.1 | 4 | 104 | 100 | 1307.763 | 99 | 47 | 56 | 0.009016 | 76466 | true | success |
| dc90f77 | 4 | 100000 | 0.1 | 5 | 105 | 100 | 605.7485 | 99 | 24 | 71 | 0.016402 | 165085 | true | success |
| dc90f77 | 4 | 500000 | 0 | 1 | 101 | 100 | 5.7977 | 490 | 0 | 490 | 0.046286 | 86241096 | true | success |
| dc90f77 | 4 | 500000 | 0 | 2 | 102 | 100 | 5.7151 | 490 | 0 | 490 | 0.044565 | 87487533 | true | success |
| dc90f77 | 4 | 500000 | 0 | 3 | 103 | 100 | 6.3348 | 490 | 0 | 490 | 0.048568 | 78929090 | true | success |
| dc90f77 | 4 | 500000 | 0 | 4 | 104 | 100 | 5.1625 | 490 | 0 | 490 | 0.038986 | 96852300 | true | success |
| dc90f77 | 4 | 500000 | 0 | 5 | 105 | 100 | 5.6677 | 490 | 0 | 490 | 0.045221 | 88219207 | true | success |
| dc90f77 | 4 | 500000 | 0.05 | 1 | 101 | 100 | 2951.8258 | 490 | 116 | 369 | 0.229292 | 169387 | true | success |
| dc90f77 | 4 | 500000 | 0.05 | 2 | 102 | 100 | 3240.5435 | 490 | 128 | 353 | 0.13619 | 154295 | true | success |
| dc90f77 | 4 | 500000 | 0.05 | 3 | 103 | 100 | 2935.5639 | 490 | 113 | 369 | 0.120487 | 170325 | true | success |
| dc90f77 | 4 | 500000 | 0.05 | 4 | 104 | 100 | 2726.5755 | 490 | 106 | 367 | 0.060237 | 183380 | true | success |
| dc90f77 | 4 | 500000 | 0.05 | 5 | 105 | 100 | 3124.8106 | 490 | 124 | 373 | 0.077542 | 160010 | true | success |
| dc90f77 | 4 | 500000 | 0.1 | 1 | 101 | 100 | 6152.2412 | 490 | 241 | 267 | 0.098287 | 81271 | true | success |
| dc90f77 | 4 | 500000 | 0.1 | 2 | 102 | 100 | 6552.3874 | 490 | 260 | 263 | 0.087536 | 76308 | true | success |
| dc90f77 | 4 | 500000 | 0.1 | 3 | 103 | 100 | 5238.7683 | 490 | 208 | 298 | 0.066795 | 95442 | true | success |
| dc90f77 | 4 | 500000 | 0.1 | 4 | 104 | 100 | 5250.9445 | 490 | 208 | 293 | 0.139219 | 95221 | true | success |
| dc90f77 | 4 | 500000 | 0.1 | 5 | 105 | 100 | 6248.4203 | 490 | 245 | 301 | 0.118595 | 80020 | true | success |
| dc90f77 | 4 | 1000000 | 0 | 1 | 101 | 100 | 10.3308 | 978 | 0 | 978 | 0.041725 | 96797925 | true | success |
| dc90f77 | 4 | 1000000 | 0 | 2 | 102 | 100 | 11.2074 | 978 | 0 | 978 | 0.043168 | 89226761 | true | success |
| dc90f77 | 4 | 1000000 | 0 | 3 | 103 | 100 | 9.8682 | 978 | 0 | 978 | 0.039197 | 101335603 | true | success |
| dc90f77 | 4 | 1000000 | 0 | 4 | 104 | 100 | 10.8346 | 978 | 0 | 978 | 0.043273 | 92296901 | true | success |
| dc90f77 | 4 | 1000000 | 0 | 5 | 105 | 100 | 15.4256 | 978 | 0 | 978 | 0.062553 | 64827300 | true | success |
| dc90f77 | 4 | 1000000 | 0.05 | 1 | 101 | 100 | 5234.5926 | 978 | 208 | 757 | 0.046963 | 191037 | true | success |
| dc90f77 | 4 | 1000000 | 0.05 | 2 | 102 | 100 | 6145.6394 | 978 | 244 | 732 | 0.072542 | 162717 | true | success |
| dc90f77 | 4 | 1000000 | 0.05 | 3 | 103 | 100 | 4530.1711 | 978 | 180 | 780 | 0.033762 | 220742 | true | success |
| dc90f77 | 4 | 1000000 | 0.05 | 4 | 104 | 100 | 4863.0785 | 978 | 190 | 759 | 0.1355 | 205631 | true | success |
| dc90f77 | 4 | 1000000 | 0.05 | 5 | 105 | 100 | 5755.9966 | 978 | 228 | 744 | 0.117334 | 173732 | true | success |
| dc90f77 | 4 | 1000000 | 0.1 | 1 | 101 | 100 | 11382.2228 | 978 | 443 | 567 | 0.099185 | 87856 | true | success |
| dc90f77 | 4 | 1000000 | 0.1 | 2 | 102 | 100 | 11692.457 | 978 | 464 | 561 | 0.095094 | 85525 | true | success |
| dc90f77 | 4 | 1000000 | 0.1 | 3 | 103 | 100 | 10571.2377 | 978 | 417 | 604 | 0.066031 | 94596 | true | success |
| dc90f77 | 4 | 1000000 | 0.1 | 4 | 104 | 100 | 11284.6585 | 978 | 440 | 573 | 0.064783 | 88616 | true | success |
| dc90f77 | 4 | 1000000 | 0.1 | 5 | 105 | 100 | 11179.4299 | 978 | 444 | 612 | 0.074684 | 89450 | true | success |

</details>

#### Window-4 summary

Aggregate goodput is the sum of delivered payload bytes divided by the sum of
elapsed time for all successful trials in the group.

| Commit | Window | Payload (bytes) | Loss | Successful trials | Median elapsed (ms) | Mean retransmissions | Aggregate goodput (B/s) |
|:---|---:|---:|---:|---:|---:|---:|---:|
| dc90f77 | 4 | 100000 | 0 | 5 | 1.5345 | 0 | 74841336 |
| dc90f77 | 4 | 100000 | 0.05 | 5 | 505.5015 | 16.6 | 234808 |
| dc90f77 | 4 | 100000 | 0.1 | 5 | 1307.763 | 43 | 86998 |
| dc90f77 | 4 | 500000 | 0 | 5 | 5.7151 | 0 | 87175446 |
| dc90f77 | 4 | 500000 | 0.05 | 5 | 2951.8258 | 117.4 | 166897 |
| dc90f77 | 4 | 500000 | 0.1 | 5 | 6152.2412 | 232.4 | 84911 |
| dc90f77 | 4 | 1000000 | 0 | 5 | 10.8346 | 0 | 86705303 |
| dc90f77 | 4 | 1000000 | 0.05 | 5 | 5234.5926 | 210 | 188470 |
| dc90f77 | 4 | 1000000 | 0.1 | 5 | 11284.6585 | 441.6 | 89111 |

<details>
<summary>Raw window-1 trials</summary>

These controlled stop-and-wait trials use the same executable and settings as
the Go-Back-N trials, except the window size is 1. A failed transfer has no
final sender statistics because the CLI exits as soon as the retry limit is
reached.

| Commit | Window | Payload (bytes) | Loss | Sender seed | Receiver seed | Timeout (ms) | Elapsed (ms) | Packets | Retransmissions | RTT samples | Mean RTT (ms) | Goodput (B/s) | Hash match | Result |
|:---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|:---:|:---|
| dc90f77 | 1 | 100000 | 0 | 1 | 101 | 100 | 3.666 | 99 | 0 | 99 | 0.03703 | 27277687 | true | success |
| dc90f77 | 1 | 100000 | 0 | 2 | 102 | 100 | 3.632 | 99 | 0 | 99 | 0.036686 | 27533040 | true | success |
| dc90f77 | 1 | 100000 | 0 | 3 | 103 | 100 | 3.6606 | 99 | 0 | 99 | 0.036975 | 27317926 | true | success |
| dc90f77 | 1 | 100000 | 0 | 4 | 104 | 100 | 3.6967 | 99 | 0 | 99 | 0.03734 | 27051154 | true | success |
| dc90f77 | 1 | 100000 | 0 | 5 | 105 | 100 | 3.0771 | 99 | 0 | 99 | 0.031081 | 32498131 | true | success |
| dc90f77 | 1 | 100000 | 0.05 | 1 | 101 | 100 | 607.8216 | 99 | 6 | 93 | 0.048687 | 164522 | true | success |
| dc90f77 | 1 | 100000 | 0.05 | 2 | 102 | 100 | 1510.8045 | 99 | 15 | 86 | 0.045865 | 66190 | true | success |
| dc90f77 | 1 | 100000 | 0.05 | 3 | 103 | 100 | 1110.1933 | 99 | 11 | 88 | 0.042122 | 90074 | true | success |
| dc90f77 | 1 | 100000 | 0.05 | 4 | 104 | 100 | 1211.0899 | 99 | 12 | 88 | 0.059735 | 82570 | true | success |
| dc90f77 | 1 | 100000 | 0.05 | 5 | 105 | 100 | 1006.806 | 99 | 10 | 90 | 0.047135 | 99324 | true | success |
| dc90f77 | 1 | 100000 | 0.1 | 1 | 101 | 100 | 2514.9856 | 99 | 25 | 77 | 0.024694 | 39762 | true | success |
| dc90f77 | 1 | 100000 | 0.1 | 2 | 102 | 100 | 2716.5107 | 99 | 27 | 77 | 0.038887 | 36812 | true | success |
| dc90f77 | 1 | 100000 | 0.1 | 3 | 103 | 100 | 2514.3424 | 99 | 25 | 82 | 0.045573 | 39772 | true | success |
| dc90f77 | 1 | 100000 | 0.1 | 4 | 104 | 100 | 2616.6914 | 99 | 26 | 83 | 0.047634 | 38216 | true | success |
| dc90f77 | 1 | 100000 | 0.1 | 5 | 105 | 100 | 2413.9937 | 99 | 24 | 81 | 0.033532 | 41425 | true | success |
| dc90f77 | 1 | 500000 | 0 | 1 | 101 | 100 | 15.4301 | 490 | 0 | 490 | 0.030438 | 32404197 | true | success |
| dc90f77 | 1 | 500000 | 0 | 2 | 102 | 100 | 14.8557 | 490 | 0 | 490 | 0.028222 | 33657115 | true | success |
| dc90f77 | 1 | 500000 | 0 | 3 | 103 | 100 | 14.5102 | 490 | 0 | 490 | 0.029612 | 34458519 | true | success |
| dc90f77 | 1 | 500000 | 0 | 4 | 104 | 100 | 16.3264 | 490 | 0 | 490 | 0.032282 | 30625245 | true | success |
| dc90f77 | 1 | 500000 | 0 | 5 | 105 | 100 | 15.0461 | 490 | 0 | 490 | 0.030706 | 33231203 | true | success |
| dc90f77 | 1 | 500000 | 0.05 | 1 | 101 | 100 | 6442.2871 | 490 | 64 | 438 | 0.028793 | 77612 | true | success |
| dc90f77 | 1 | 500000 | 0.05 | 2 | 102 | 100 | 6754.9558 | 490 | 67 | 432 | 0.054883 | 74020 | true | success |
| dc90f77 | 1 | 500000 | 0.05 | 3 | 103 | 100 | 6044.8768 | 490 | 60 | 437 | 0.037137 | 82715 | true | success |
| dc90f77 | 1 | 500000 | 0.05 | 4 | 104 | 100 | 6144.3807 | 490 | 61 | 433 | 0.038775 | 81375 | true | success |
| dc90f77 | 1 | 500000 | 0.05 | 5 | 105 | 100 | 6144.5153 | 490 | 61 | 434 | 0.042206 | 81373 | true | success |
| dc90f77 | 1 | 500000 | 0.1 | 1 | 101 | 100 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | giving up on seq 382 |
| dc90f77 | 1 | 500000 | 0.1 | 2 | 102 | 100 | 13176.9072 | 490 | 131 | 385 | 0.03739 | 37945 | true | success |
| dc90f77 | 1 | 500000 | 0.1 | 3 | 103 | 100 | 11768.913 | 490 | 117 | 402 | 0.040722 | 42485 | true | success |
| dc90f77 | 1 | 500000 | 0.1 | 4 | 104 | 100 | 12474.6836 | 490 | 124 | 395 | 0.039106 | 40081 | true | success |
| dc90f77 | 1 | 500000 | 0.1 | 5 | 105 | 100 | 11966.9831 | 490 | 119 | 391 | 0.032714 | 41782 | true | success |
| dc90f77 | 1 | 1000000 | 0 | 1 | 101 | 100 | 30.1004 | 978 | 0 | 978 | 0.029204 | 33222150 | true | success |
| dc90f77 | 1 | 1000000 | 0 | 2 | 102 | 100 | 28.0046 | 978 | 0 | 978 | 0.027075 | 35708419 | true | success |
| dc90f77 | 1 | 1000000 | 0 | 3 | 103 | 100 | 28.8658 | 978 | 0 | 978 | 0.028989 | 34643072 | true | success |
| dc90f77 | 1 | 1000000 | 0 | 4 | 104 | 100 | 29.5297 | 978 | 0 | 978 | 0.028612 | 33864211 | true | success |
| dc90f77 | 1 | 1000000 | 0 | 5 | 105 | 100 | 29.1489 | 978 | 0 | 978 | 0.028116 | 34306612 | true | success |
| dc90f77 | 1 | 1000000 | 0.05 | 1 | 101 | 100 | 11589.6338 | 978 | 115 | 881 | 0.048507 | 86284 | true | success |
| dc90f77 | 1 | 1000000 | 0.05 | 2 | 102 | 100 | 12700.7879 | 978 | 126 | 872 | 0.049733 | 78735 | true | success |
| dc90f77 | 1 | 1000000 | 0.05 | 3 | 103 | 100 | 9071.4308 | 978 | 90 | 896 | 0.039208 | 110236 | true | success |
| dc90f77 | 1 | 1000000 | 0.05 | 4 | 104 | 100 | 11786.8338 | 978 | 117 | 873 | 0.037865 | 84840 | true | success |
| dc90f77 | 1 | 1000000 | 0.05 | 5 | 105 | 100 | 11385.9491 | 978 | 113 | 874 | 0.037736 | 87828 | true | success |
| dc90f77 | 1 | 1000000 | 0.1 | 1 | 101 | 100 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | giving up on seq 382 |
| dc90f77 | 1 | 1000000 | 0.1 | 2 | 102 | 100 | N/A | N/A | N/A | N/A | N/A | N/A | N/A | giving up on seq 726 |
| dc90f77 | 1 | 1000000 | 0.1 | 3 | 103 | 100 | 21132.7815 | 978 | 210 | 813 | 0.043494 | 47320 | true | success |
| dc90f77 | 1 | 1000000 | 0.1 | 4 | 104 | 100 | 24447.5139 | 978 | 243 | 786 | 0.044726 | 40904 | true | success |
| dc90f77 | 1 | 1000000 | 0.1 | 5 | 105 | 100 | 22432.4444 | 978 | 223 | 794 | 0.036502 | 44578 | true | success |

</details>

#### Window-1 summary

Aggregate goodput includes successful trials only. Retry-exhaustion failures
are counted separately and are not replaced with new seeds.

| Commit | Window | Payload (bytes) | Loss | Successful trials | Protocol failures | Median elapsed (ms) | Mean retransmissions | Aggregate goodput (B/s) |
|:---|---:|---:|---:|---:|---:|---:|---:|---:|
| dc90f77 | 1 | 100000 | 0 | 5 | 0 | 3.6606 | 0 | 28196973 |
| dc90f77 | 1 | 100000 | 0.05 | 5 | 0 | 1110.1933 | 10.8 | 91798 |
| dc90f77 | 1 | 100000 | 0.1 | 5 | 0 | 2514.9856 | 25.4 | 39134 |
| dc90f77 | 1 | 500000 | 0 | 5 | 0 | 15.0461 | 0 | 32821967 |
| dc90f77 | 1 | 500000 | 0.05 | 5 | 0 | 6144.5153 | 62.6 | 79287 |
| dc90f77 | 1 | 500000 | 0.1 | 4 | 1 | 12220.8334 | 122.75 | 40496 |
| dc90f77 | 1 | 1000000 | 0 | 5 | 0 | 29.1489 | 0 | 34329012 |
| dc90f77 | 1 | 1000000 | 0.05 | 5 | 0 | 11589.6338 | 112.2 | 88441 |
| dc90f77 | 1 | 1000000 | 0.1 | 3 | 2 | 22432.4444 | 225.33 | 44109 |
