# Dero Swap Service

A cryptocurrency swap service supporting multiple blockchain networks including Dero, Bitcoin, Litecoin, Monero, and Pirate Chain (ARRR).

## Overview

The Dero Swap Service enables secure peer-to-peer cryptocurrency swaps between different blockchain networks. It can operate in two modes:
- **Server Mode**: Hosts a swap service that clients can connect to
- **Client Mode**: Connects to existing swap servers to perform trades

## Supported Cryptocurrencies

- **DERO** (Primary blockchain)
- **Bitcoin (BTC)**
- **Litecoin (LTC)**
- **Monero (XMR)**
- **Pirate Chain (ARRR)**

## Configuration

### Configuration Files

The service requires two main configuration files:

1. **`config.json`** - Main configuration settings
2. **`fees.json`** - Fee structure for swaps and withdrawals

### Configuration Structure

#### Main Configuration Example (`config.json`)

```json
{
  "listen": "0.0.0.0:8080",

  "server": "swap.example.com:8080",
  "nickname": "YourNickname",

  "btc_daemon": "127.0.0.1:8334",
  "btc_dir": "/path/to/btc/data_dir",
  "btc_login": "username:password",

  "ltc_daemon": "127.0.0.1:9332",
  "ltc_dir": "/path/to/ltc/data_dir",
  "ltc_login": "username:password",

  "arrr_daemon": "127.0.0.1:45453",
  "arrr_dir": "/path/to/arrr/data_dir",

  "dero_daemon": "127.0.0.1:10102",
  "dero_wallet": "127.0.0.1:10103",
  "dero_login": "username:password",

  "Monero_wallet": "127.0.0.1:18090",
  "monero_login": "username:password",

  "pairs": ["btc", "ltc", "xmr", "arrr"]
}
```

Use a dummy path if the deamon runs on a remote machine. The path is only required for local cookie authentication.

#### Fee Configuration (`fees.json`)

```json
{
  "fees": 12,
  "withdrawal": {
    "ltc": 0.0002,
    "xmr": 0.00006,
    "arrr": 0.001
  }
}
```

### Configuration Parameters

| Parameter | Description | Required |
|-----------|-------------|----------|
| `listen`        | Server listening address                   | Server only          |
| `server`        | Server address to connect to               | Client only          |
| `nickname`      | Client nickname for identification         | Client only          |
| `btc_daemon`    | Bitcoin daemon RPC endpoint                | If BTC enabled       |
| `btc_dir`       | Bitcoin wallet directory path              | If BTC enabled       |
| `btc_login`     | Bitcoin RPC authentication                 | If BTC enabled       |
| `ltc_daemon`    | Litecoin daemon RPC endpoint               | If LTC enabled       |
| `ltc_dir`       | Litecoin wallet directory path             | If LTC enabled       |
| `ltc_login`     | Litecoin RPC authentication                | If LTC enabled       |
| `arrr_daemon`   | Pirate Chain daemon RPC endpoint           | If ARRR enabled      |
| `arrr_dir`      | Pirate Chain wallet directory path         | If ARRR enabled      |
| `dero_daemon`   | Dero daemon RPC endpoint                   | Always required      |
| `dero_wallet`   | Dero wallet RPC endpoint                   | Always required      |
| `dero_login`    | Dero wallet RPC authentication             | Always required      |
| `monero_wallet` | Monero wallet RPC endpoint                 | If XMR pair enabled  |
| `pairs`         | Array of enabled trading pairs             | Required             |

## Installation & Setup

### Prerequisites

Ensure you have the following blockchain daemons and wallets running:

1. **Dero Daemon** and **Dero Wallet** (required)
2. **Bitcoin Core** (if trading BTC)
3. **Litecoin Core** (if trading LTC)
4. **Monero Wallet** (if trading XMR)
5. **Pirate Chain Daemon** (if trading ARRR)

### Building

```bash
go mod tidy
go build -o dero-swap
```

## Usage

### Server Mode

To run as a swap server:

```bash
./dero-swap --server
```

This will:
- Load configuration from `config.json`
- Initialize all configured wallets
- Start the swap service on the configured listen address
- Begin accepting client connections

### Client Mode

To run as a client:

```bash
./dero-swap
```

This will:
- Connect to the configured server
- Register with the provided nickname
- Enable trading for all available pairs

## Directory Structure

```
swaps/
├── active/     # Currently processing swaps
├── expired/    # Expired/failed swaps  
└── done/       # Completed swaps
```

## Supported Operations

- Cross-chain cryptocurrency swaps
- Price discovery
- Automatic swap execution
- Wallet balance monitoring
- Transaction history tracking

## Version

Current version: **v0.8.5**

## Repository

GitHub: https://github.com/8lecramm/dero-swap-service
