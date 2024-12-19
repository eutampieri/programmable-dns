# Programmable DNS

## Overview

This project is a DNS resolver that supports various types of DNS resolution strategies, including basic, static, suffix, and merge resolvers. It allows for flexible configuration of DNS settings through a JSON configuration file, enabling users to define networks, domains, and resolver behaviors.

## Features

- **Basic Resolver**: Resolves DNS queries using a specified DNS server.
- **DoT Resolver**: Resolves DNS queries using a specified DoT server.
- **Static Resolver**: Maps domain names to IP addresses based on a static configuration.
- **Suffix Resolver**: Modifies DNS queries and responses based on specified suffixes.
- **Merge Resolver**: Combines multiple resolvers and returns the first successful response.

## Configuration

The configuration for the DNS resolver is defined in a JSON format. Below is a sample configuration file:

```json
[
  {
    "network": "10.0.0.0/24",
    "domain": "iot.example.com",
    "resolver": {
      "type": "suffix",
      "server": "10.20.0.1:53",
      "oldSuffix": "example.com.",
      "newSuffix": "iot.example.com."
    }
  },
  {
    "domain": "vpn.example.com",
    "network": "10.1.0.0/24",
    "resolver": {
      "type": "static",
      "base": "vpn.example.com",
      "domainsToIPs": {
        "oneiric": "10.1.0.1",
        "focal": "10.1.0.2"
      }
    }
  },
  {
    "domain": "lan.mydomain.org",
    "network": "10.2.0.0/24",
    "resolver": {
      "type": "suffix",
      "server": "10.2.0.254:53",
      "oldSuffix": "lan.",
      "newSuffix": "lan.mydomain.org."
    }
  },
  {
    "network": "192.168.1.0/24",
    "resolver": {
      "type": "merge",
      "resolvers": [
        {
          "type": "basic",
          "server": "10.20.0.1:53"
        },
        {
          "type": "basic",
          "server": "10.21.0.1:53"
        }
      ]
    }
  }
]
```

### Configuration Fields

- **network**: The CIDR notation of the network.
- **domain**: The domain name associated with the resolver.
- **resolver**: An object defining the type of resolver and its specific settings.
  - **type**: The type of resolver (e.g., `basic`, `static`, `suffix`, `merge`, `dot`).
  - **server**: The DNS server to query (for `basic`, `dot` and `suffix` resolvers).
  - **oldSuffix**: The suffix to be replaced in DNS queries (for `suffix` resolvers).
  - **newSuffix**: The new suffix to be used in DNS queries (for `suffix` resolvers).
  - **base**: The base domain for static resolution (for `static` resolvers).
  - **domainsToIPs**: A mapping of domain names to IP addresses (for `static` resolvers).
  - **resolvers**: An array of resolvers to be merged (for `merge` resolvers).

## Usage

To use this DNS resolver, you need to:

1. Create a JSON configuration file based on the provided sample.
