# Reliable UDP File Transfer

A reliable file transfer protocol built **from scratch in Go** over UDP, implementing **Stop-and-Wait ARQ** and **Selective Repeat ARQ** with sliding windows, CRC32 error detection, concurrent ACK processing, and per-packet retransmission timers.

> **Goal:** Recreate reliability on top of an unreliable transport protocol (UDP) without relying on TCP.

---

## Overview

UDP provides fast, connectionless communication but does **not** guarantee:

- Reliable delivery
- Ordered delivery
- Duplicate suppression
- Error recovery

This project implements these reliability mechanisms in user space by building a transport protocol over UDP.

---

# Project Highlights

- Reliable file transfer over UDP
- Custom binary packet format
- CRC32 checksum based error detection
- Stop-and-Wait ARQ implementation
- Sliding Window protocol
- Selective Repeat ARQ
- Out-of-order packet buffering
- Independent ACK processing
- Per-packet retransmission timers
- Concurrent sender using goroutines
- Recovery from packet loss
- Recovery from ACK loss
- Unit tests for packet serialization and checksum validation

---

# At a Glance

| Metric                            |                          Value |
| --------------------------------- | -----------------------------: |
| Programming Language              |                             Go |
| Transport Protocol                |                            UDP |
| Reliability Protocols Implemented |                              2 |
| Synchronization Primitives        | 2 (Mutex + Condition Variable) |
| Concurrent Goroutines             |                              2 |
| Packet Integrity Algorithm        |                          CRC32 |
| Retransmission Strategy           |              Per-packet timers |
| Window Management                 |                 Sliding Window |
| Receiver Buffering                |                            Yes |
| Out-of-order Delivery Support     |                            Yes |
| Packet Loss Recovery              |                            Yes |
| ACK Loss Recovery                 |                            Yes |

---

# Features

## Reliable Data Transfer

Implemented reliability on top of UDP using:

- Stop-and-Wait ARQ
- Sliding Window
- Selective Repeat ARQ

Only lost packets are retransmitted, improving throughput compared to Go-Back-N.

---

## Custom Packet Format

Each packet contains:

```text
+---------------------------------------------------------------+
| SeqNum | AckNum | Flags | Length | CRC32 | Payload |
+---------------------------------------------------------------+
```

Fields:

- Sequence Number
- Acknowledgement Number
- Packet Flags
- Payload Length
- CRC32 Checksum
- Payload

---

## Error Detection

Every transmitted packet is protected using **CRC32**.

Receiver verifies integrity before processing.

Corrupted packets are rejected automatically.

---

## Sliding Window

The sender maintains a configurable transmission window.

```text
          Sender Window

Base                           NextSeq
 |                                 |
 v                                 v

+----+----+----+----+----+----+----+
| 10 | 11 | 12 | 13 | 14 | 15 | 16 |
+----+----+----+----+----+----+----+
        Outstanding packets
```

The sender continues transmitting until the window becomes full.

---

## Selective Repeat ARQ

Unlike Stop-and-Wait, multiple packets can be outstanding simultaneously.

Receiver:

- accepts out-of-order packets
- buffers them
- acknowledges each independently

Only missing packets are retransmitted.

---

## Receiver Buffering

Example:

Packets arrive:

```text
1
2
4
5
3
```

Receiver state:

```text
Receive 1
Receive 2
Buffer 4
Buffer 5
Receive 3

↓

Deliver

1
2
3
4
5
```

This preserves ordered delivery to the application.

---

# Architecture

```text
                    +----------------------+
                    |      Sender          |
                    +----------------------+
                               |
                Read file chunk
                               |
                               v
                    Create Packet
                               |
                               v
                    Marshal Packet
                               |
                               v
                    Send over UDP
                               |
                               v
                 Start Retransmission Timer
                               |
         +---------------------+----------------------+
         |                                            |
         | ACK received                               |
         |                                            |
         v                                            v
 Stop Timer & Slide Window                  Timeout Occurs
                                                     |
                                                     v
                                           Retransmit Packet
```

---

## Receiver

```text
Receive UDP Packet
        |
        v
Verify CRC32
        |
        v
Old Packet?
 |        \
Yes        No
 |          |
ACK Again   |
            v
Within Receive Window?
 |          \
No          Yes
 |            |
Ignore        |
              v
Already Buffered?
 |            \
Yes            No
 |              |
ACK Again       |
                v
Buffer Packet
                |
                v
Send ACK
                |
                v
Flush Consecutive Packets
```

---

# Concurrency Model

Sender uses two goroutines.

```text
             Sender

        +----------------+
        | SendFile()     |
        +----------------+
                 |
                 |
                 +--------------------+
                                      |
                                      |
                                      v
                         +-----------------------+
                         | receiveACKs()         |
                         +-----------------------+
```

Shared state is protected using:

- sync.Mutex
- sync.Cond

This prevents race conditions while allowing asynchronous ACK processing.

---

# Retransmission Mechanism

Each transmitted packet starts its own timer.

```text
Packet Sent
     |
     v
Start Timer
     |
     v
ACK Received?
     |
+----+----+
|         |
Yes       No
|          |
Stop      Timeout
Timer      |
            v
     Retransmit Packet
            |
            v
      Restart Timer
```

Unlike Go-Back-N, only the missing packet is retransmitted.

---

# Reliability Scenarios Tested

## Packet Loss

```
Packet 5 Lost

↓

Receiver buffers

6
7
8

↓

Sender timeout

↓

Retransmit 5

↓

Receiver flushes

5
6
7
8
```

---

## ACK Loss

```
Receiver receives packet 5

↓

ACK lost

↓

Sender timeout

↓

Retransmit packet 5

↓

Receiver detects duplicate

↓

ACK sent again

↓

Transfer completes successfully
```

---

# Synchronization Design

The sender shares multiple data structures across goroutines.

Protected resources include:

- Outstanding packets
- ACK table
- Retransmission timers
- Sliding window variables

Synchronization primitives used:

- sync.Mutex
- sync.Cond

The condition variable allows the sender to sleep when the transmission window is full instead of busy waiting.

---

# Project Structure

```
udp-file-transfer/

├── cmd/
│   ├── sender/
│   └── receiver/
│
├── internal/
│   ├── protocol/
│   ├── transport/
│   └── file/
│
├── sample.txt
├── go.mod
└── README.md
```

---

# Testing

Implemented unit tests for:

- Packet serialization
- Packet deserialization
- CRC32 checksum validation
- Corrupted packet detection

Integration tests performed:

- Normal file transfer
- Simulated packet loss
- Simulated ACK loss
- Sliding window behaviour
- Out-of-order packet buffering

---

# Design Decisions

### Why UDP?

To implement reliability manually instead of relying on TCP.

---

### Why Selective Repeat instead of Go-Back-N?

Selective Repeat retransmits only lost packets, reducing unnecessary network traffic.

---

### Why Per-Packet Timers?

Each outstanding packet maintains an independent retransmission timer, enabling selective retransmissions.

---

### Why sync.Cond?

Using a condition variable allows the sender to block efficiently while the window is full instead of repeatedly polling shared state.

---

### Why CRC32?

CRC32 provides fast and reliable error detection for transmitted packets while keeping packet overhead low.

---

# Future Improvements

- Adaptive retransmission timeout based on RTT estimation
- Configurable packet loss simulator
- Throughput and RTT statistics
- Command-line configuration
- Fast retransmit
- Congestion control
- Flow control
- Performance benchmarking

---

# Technologies

- Go
- UDP Sockets
- Goroutines
- Mutexes
- Condition Variables
- CRC32
- Concurrent Programming
- Computer Networks

---

# Key Learnings

Building this project provided hands-on experience with:

- Reliable transport protocols
- Sliding window algorithms
- Selective Repeat ARQ
- Concurrent programming in Go
- Synchronization primitives
- Network protocol design
- State machine implementation
- Error detection using CRC32
- Race condition analysis
- Systems programming
