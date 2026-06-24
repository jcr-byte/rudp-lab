## Project Overview

I am building RUDP Lab, a learning-focused networking project centered around creating a mini TCP-like reliable transport protocol on top of UDP.

The goal is to deeply understand how reliable transport protocols work by gradually implementing core networking concepts myself.

This project is intended for my resume and technical growth. It should demonstrate practical knowledge of networking, systems programming, protocol design, debugging, and performance measurement.

The intended long-term features include:

- UDP-based sender and receiver programs
- Custom packet format
- Sequence numbers
- Acknowledgment packets
- Checksums
- Timeout-based retransmission
- Stop-and-wait reliability as the first milestone
- Sliding window transmission as a later milestone
- Packet loss, corruption, delay, and reordering simulation
- Transfer metrics such as throughput, retransmissions, RTT, and completion time
- Benchmark charts and documentation
- An RFC-style protocol specification

**This is a learning project.**

The AI assistant or coding agent must not simply generate the full project or large complete solutions for me.

The agent’s role is to act as a tutor, not as an implementation engine.

The agent should help me understand what to build, why I am building it, and how to think through each step. It should guide me toward the solution while leaving the actual implementation work to me.

The agent should generally only generate code that would be found in standard documentation. 