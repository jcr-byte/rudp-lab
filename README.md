# RUDP Lab

RUDP Lab is a learning-focused implementation of reliable transport over UDP.
The current protocol uses a Go-Back-N sliding window with sequence numbers,
cumulative acknowledgments, checksums, timeout-based retransmission, and
simulated packet loss. Setting the window size to 1 provides stop-and-wait
behavior for comparison.

## Project goals

I built RUDP Lab to better understand how reliable transport protocols provide ordered and dependable data delivery over an unreliable network. By implementing reliability on top of UDP, I was able to explore sequence numbers,
acknowledgments, checksums, retransmission timeouts, sliding windows, and packet-loss recovery without relying on TCP to handle those concerns automatically.

This project also served as an introduction to Go. It gave me experience working with Go’s networking APIs, binary data encoding, error handling, testing, and concurrent sender-receiver behavior. RUDP Lab is an educational
protocol rather than a production-ready replacement for TCP. The projects purpose is to explore the concepts and engineering tradeoffs involved in reliable transport.

## Highlights

- Designed a custom binary packet format with sequence numbers, packet-type flags, checksums, and payload data
- Implemented Go-Back-N transmission using a configurable sliding window and cumulative acknowledgments
- Supported stop-and-wait behavior through a window size of one
- Added timeout-based retransmission and retry limits to recover from lost data packets and acknowledgments
- Built a reproducible packet-loss simulator using configurable loss rates and random seeds
- Implemented transfer completion using a FIN/ACK packet
- Collected goodput, retransmission, completion-time, and RTT statistics while applying Karn’s algorithm to RTT sampling
- Created unit and integration tests covering packet validation, lossy transfers, out-of-order packets, window retransmission, and retry exhaustion
- Benchmarked stop-and-wait against Go-Back-N, observing approximately 2.02×–2.66× higher aggregate goodput with a window size of four

## How it works

RUDP Lab transfers data between a sender and receiver using UDP datagrams. The sender divides the input into payloads of up to 1,024 bytes, assigns each packet a sequence number, calculates a checksum, and transmits packets within a
configurable sliding window.

The receiver validates each packet’s checksum and accepts it only if its sequence number matches the next expected sequence number. It then responds with a cumulative acknowledgment indicating the next packet it expects. Duplicate and
out-of-order packets are discarded but still trigger an acknowledgment of the current expected sequence number.

As acknowledgments arrive, the sender advances its window and transmits additional packets. If the sender does not receive an acknowledgment before the retransmission timeout, it resends every unacknowledged packet in the current
window. After all data is acknowledged, a FIN/ACK exchange completes the transfer.

### Packet format

Each packet contains a fixed 5-byte header followed by an optional payload.
Multibyte fields are encoded in big-endian byte order.

| Field | Size | Purpose |
|---|---:|---|
| Flag | 1 byte | Identifies the packet as DATA (`1`), ACK (`2`), or FIN (`3`) |
| Sequence number | 2 bytes | Orders DATA packets or indicates the next expected sequence number in an ACK |
| Checksum | 2 bytes | Detects accidental corruption in the header or payload |
| Payload | 0-1,024 bytes | Contains application data for DATA packets |

The checksum is calculated across the entire encoded packet while the checksum
field is set to zero. The receiver repeats this calculation and discards the
packet if the result does not match the stored checksum. ACK and FIN packets
use the same header format but do not contain payload data.

### Sequence numbers and acknowledgments

Each DATA packet is assigned a sequence number beginning at `1`. The receiver
tracks the next sequence number it expects and accepts only packets that arrive
in order. When a packet is accepted, its payload is written to the output and
the expected sequence number is incremented.

Acknowledgments are cumulative: the ACK sequence number identifies the next
DATA packet the receiver expects. For example, an ACK with sequence number `5`
confirms that all packets through sequence number `4` were received in order.

Duplicate and out-of-order packets are discarded. However, the receiver still
sends an ACK containing its current expected sequence number. This informs the
sender of the last point at which the transfer was complete and allows lost
data packets or acknowledgments to be recovered through retransmission.

### Go-Back-N sliding window

The sender uses a configurable Go-Back-N sliding window to keep multiple DATA
packets in flight without waiting for an individual ACK after every packet.
The window begins at the oldest unacknowledged packet and limits how many
packets may be outstanding at one time.

When a cumulative ACK arrives, the sender marks every earlier packet as
acknowledged, advances the beginning of the window, and transmits new packets
to fill the available space. This reduces the time spent waiting between
packets and improves goodput compared with stop-and-wait transmission.

If an ACK does not arrive before the retransmission timeout, the sender goes
back to the oldest unacknowledged packet and retransmits every outstanding
packet in the current window. Setting the window size to `1` uses the same
implementation but produces stop-and-wait behavior, providing a controlled
baseline for comparison.

### Timeouts and retransmission

The sender starts a configurable retransmission timer when packets are placed
in an empty window. Each cumulative ACK that advances the window represents
forward progress and resets the timer for the remaining outstanding packets.

If the timer expires before the sender receives an ACK that advances the
window, every unacknowledged packet in the current window is retransmitted.
The sender also tracks consecutive timeouts without progress and ends the
transfer with an error when the configured retry limit is reached. This
prevents it from retransmitting indefinitely when the receiver is unavailable
or packet loss remains too high.

RUDP Lab records round-trip time (RTT) samples for packets acknowledged without
retransmission. Samples from retransmitted packets are excluded according to
Karn's algorithm because the sender cannot determine whether the ACK
corresponds to the original transmission or a later attempt.

### Transfer completion

After every DATA packet has been acknowledged, the sender transmits a FIN
packet using the next sequence number. The receiver responds with an ACK for
the following sequence number, confirming that it received the FIN. If the
sender does not receive this ACK before the timeout, it retransmits the FIN up
to the configured retry limit.

After sending the final ACK, the receiver remains available for a short linger
period instead of closing its UDP socket immediately. If the final ACK was lost
and the sender retransmits the same FIN, the receiver can acknowledge it again.
The transfer ends once the sender receives the FIN acknowledgment and the
receiver's linger period expires.

## Loss simulation

RUDP Lab includes a configurable loss simulator for testing reliability under
unreliable network conditions. Before each outgoing UDP datagram is sent, the
simulator uses the configured loss rate to decide whether to drop it.
A dropped datagram is reported to the protocol as successfully written, which
reproduces the lack of delivery feedback normally provided by UDP.

Loss is applied independently to outgoing DATA and FIN packets on the sender
and outgoing ACK packets on the receiver. Each side accepts its own random seed,
allowing the same loss pattern to be reproduced across test and benchmark runs.

The simulator currently models packet loss only. It does not introduce delay,
corruption, duplication, or packet reordering.

## Project structure

```text
cmd/rudp/
  main.go                  Command-line sender and receiver
internal/packet/
  packet.go                Packet format, encoding, decoding, and checksums
  packet_test.go           Packet and checksum unit tests
internal/transport/
  send.go                  Go-Back-N sender and transfer metrics
  recv.go                  Receiver, cumulative ACKs, and FIN handling
  transport_test.go        End-to-end and transport behavior tests
internal/netsim/
  lossy.go                 Seeded packet-loss simulation
go.mod                     Go module definition
README.md                  Project documentation and benchmark results
```

The command-line application is kept separate from the protocol packages. The
`packet` package defines the wire format, `transport` implements reliable data
delivery, and `netsim` provides controlled packet loss for testing and
benchmarking.

## Getting started

### Requirements

- Go 1.26.4, as declared in `go.mod`
- Two terminal sessions for running the sender and receiver locally

Clone the repository and enter the project directory:

```powershell
git clone https://github.com/jcr-byte/rudp-lab.git
cd rudp-lab
```

### Run the receiver

Start the receiver first. By default, it listens on UDP port `9000` and writes
the completed transfer to `received.bin` in the current directory.

```powershell
go run ./cmd/rudp recv -addr :9000
```

The `-addr` flag can be changed to listen on a different local address or port.
Run `go run ./cmd/rudp recv -h` to view all receiver options.

### Run the sender

In a second terminal, send a 100,000-byte payload to the receiver with a
four-packet window:

```powershell
go run ./cmd/rudp send -addr 127.0.0.1:9000 -size 100000 -window 4
```

The sender generates a deterministic random payload of the requested size and
prints transfer statistics after completion, including elapsed time, packet
count, retransmissions, RTT measurements, and goodput. The sender and receiver
also print SHA-256 hashes that can be compared to verify that the received data
matches the original payload.

Run `go run ./cmd/rudp send -h` to view options for the payload size, window
size, retransmission timeout, retry limit, loss rate, and random seed.

### Simulate packet loss

Packet loss can be enabled independently on both sides. Start the receiver with
a 5% ACK loss rate and an explicit seed:

```powershell
go run ./cmd/rudp recv -addr :9000 -loss 0.05 -seed 101
```

Then run the sender with a 5% DATA/FIN loss rate and its own seed:

```powershell
go run ./cmd/rudp send -addr 127.0.0.1:9000 -size 100000 -window 4 -loss 0.05 -seed 1
```

Using the same configuration and seeds reproduces the same simulated loss
decisions, making failures and performance experiments easier to investigate.

## Testing

Run the complete unit and integration test suite from the repository root:

```powershell
go test ./...
```

The packet tests verify checksum calculation, corruption detection, and
encoding/decoding round trips. The transport tests exercise:

- End-to-end transfers with and without simulated packet loss
- One-byte payloads, exact 1,024-byte chunks, and multi-packet payloads
- Stop-and-wait operation with a window size of `1`
- Filling a sliding window before receiving an ACK
- Retransmitting the outstanding window after a timeout
- Stopping after the retry limit when no ACK is received
- Discarding out-of-order packets while returning the expected cumulative ACK

End-to-end tests compare the received bytes with the original payload rather
than checking only whether the sender and receiver exited successfully.

## Design decisions and tradeoffs

### Go-Back-N over Selective Repeat

Go-Back-N was chosen as the first sliding-window protocol because it adds
pipelining while keeping the sender and receiver state manageable. The
receiver only tracks the next expected sequence number, and the sender can
recover from loss by retransmitting the outstanding window. The tradeoff is
that packets received correctly after a missing packet must be sent again,
using more bandwidth than Selective Repeat under loss.

The window size is configurable, and setting it to `1` produces stop-and-wait
behavior through the same implementation. This made it possible to compare the
two transmission strategies without maintaining separate protocol paths.

### Cumulative ACKs and out-of-order packets

ACKs report the next sequence number expected by the receiver, confirming all
earlier packets at once. This reduces acknowledgment state and allows a single
ACK to advance the sender past multiple packets. To preserve that simplicity,
the receiver discards out-of-order packets instead of buffering them. This
keeps memory use and receiver logic small but may cause additional
retransmissions.

### Fixed retransmission timeout

The sender uses a configurable fixed timeout rather than dynamically adapting
it to measured RTT. A fixed value makes the retransmission behavior easier to
reason about and keeps benchmark configurations reproducible. However, a value
that is too short can cause unnecessary retransmissions, while one that is too
long delays recovery. A production transport protocol would normally estimate
an adaptive retransmission timeout as network conditions change.

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

## AI usage

I wrote the protocol implementation myself to develop a full understanding of reliable transport over UDP, beginning with stop-and-wait and progressing to Go-Back-N. During development, I used AI as a tutor and reviewer. It was used to ask
conceptual questions, challenge design decisions, and receive feedback on my work. AI also assisted with organizing and editing the project documentation. I remained responsible for implementing, testing, debugging, and
understanding the code.

## Scope and limitations

RUDP Lab is an educational implementation rather than a production transport
protocol. Its scope is intentionally smaller than TCP and other established
reliable transports:

- The sender uses a fixed retransmission timeout rather than estimating an
  adaptive timeout from network conditions.
- There is no congestion control, receiver flow control, or dynamic window
  sizing.
- The receiver discards out-of-order packets, and the protocol does not support
  selective acknowledgments or selective retransmission.
- Sequence-number wraparound is not handled, which limits the maximum transfer
  size supported by the 16-bit sequence field.
- The protocol has no connection handshake, authentication, encryption, or
  protection against malicious traffic.
- The additive checksum is intended to demonstrate corruption detection, not
  provide strong data-integrity guarantees.
- The receiver handles one transfer at a time, and the CLI transfers generated
  payloads rather than arbitrary input files.
- The network simulator models independent packet loss but not latency,
  corruption, duplication, bandwidth limits, or reordering.

## Future improvements

- Calculate an adaptive retransmission timeout using measured RTT and RTT
  variation.
- Implement Selective Repeat with receiver-side buffering to compare its
  behavior with Go-Back-N under packet loss.
- Extend the network simulator with delay, corruption, duplication, and
  reordering controls.
- Add file input and configurable output paths to make the CLI useful for
  arbitrary file transfers.
- Automate benchmark execution and generate charts from the collected results.
- Write an RFC-style protocol specification covering the wire format, sender
  and receiver state, acknowledgment semantics, and termination behavior.
