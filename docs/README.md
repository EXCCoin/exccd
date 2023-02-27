### Table of Contents
1. [About](#About)
2. [Getting Started](#GettingStarted)
    1. [Installation](#Installation)
    2. [Configuration](#Configuration)
    3. [Controlling and Querying exccd via exccctl](#exccctlConfig)
    4. [Mining](#Mining)
3. [Help](#Help)
    1. [Network Configuration](#NetworkConfig)
    2. [Wallet](#Wallet)
4. [Contact](#Contact)
    1. [Community](#ContactCommunity)
5. [Developer Resources](#DeveloperResources)
    1. [Code Contribution Guidelines](#ContributionGuidelines)
    2. [JSON-RPC Reference](#JSONRPCReference)
    3. [Go Modules](#GoModules)
    4. [Module Hierarchy](#ModuleHierarchy)
6. [Simulation Network (--simnet) Reference](#SimnetReference)

<a name="About" />

### 1. About

exccd is a full node  implementation written in [Go](https://golang.org),
and is licensed under the [copyfree](http://www.copyfree.org) ISC License.

This software is currently under active development.  It is extremely stable and
has been in production use since February 2016.

It also properly relays newly mined blocks, maintains a transaction pool, and
relays individual transactions that have not yet made it into a block.  It
ensures all individual transactions admitted to the pool follow the rules
required into the block chain and also includes the vast majority of the more
strict checks which filter transactions based on miner requirements ("standard"
transactions).

<a name="GettingStarted" />

### 2. Getting Started

<a name="Installation" />

**2.1 Installation**<br />

The first step is to install exccd.  The installation instructions can be found
[here](https://github.com/EXCCoin/exccd/tree/master/README.md#Installation).

<a name="Configuration" />

**2.2 Configuration**<br />

exccd has a number of [configuration](https://pkg.go.dev/github.com/EXCCoin/exccd)
options, which can be viewed by running: `$ exccd --help`.

<a name="exccctlConfig" />

**2.3 Controlling and Querying exccd via exccctl**<br />

[exccctl](https://github.com/EXCCoin/exccctl) is a command line utility that can be
used to both control and query exccd via
[RPC](https://www.wikipedia.org/wiki/Remote_procedure_call).  exccd does **not**
enable its RPC server by default; You must configure at minimum both an RPC
username and password or both an RPC limited username and password:

* exccd.conf configuration file
```
[Application Options]
rpcuser=myuser
rpcpass=SomeDecentp4ssw0rd
rpclimituser=mylimituser
rpclimitpass=Limitedp4ssw0rd
```
* exccctl.conf configuration file
```
[Application Options]
rpcuser=myuser
rpcpass=SomeDecentp4ssw0rd
```
OR
```
[Application Options]
rpclimituser=mylimituser
rpclimitpass=Limitedp4ssw0rd
```
For a list of available options, run: `$ exccctl --help`

<a name="Mining" />

**2.4 Mining**<br />
exccd supports the [getwork](https://github.com/EXCCoin/exccd/tree/master/docs/json_rpc_api.mediawiki#getwork)
RPC.  The limited user cannot access this RPC.<br />

**1. Add the payment addresses with the `miningaddr` option.**<br />

```
[Application Options]
rpcuser=myuser
rpcpass=SomeDecentp4ssw0rd
miningaddr=DsExampleAddress1
miningaddr=DsExampleAddress2
```

**2. Add exccd's RPC TLS certificate to system Certificate Authority list.**<br />

`cgminer` uses [curl](https://curl.haxx.se/) to fetch data from the RPC server.
Since curl validates the certificate by default, we must install the `exccd` RPC
certificate into the default system Certificate Authority list.

**Ubuntu**<br />

1. Copy rpc.cert to /usr/share/ca-certificates: `# cp /home/user/.exccd/rpc.cert /usr/share/ca-certificates/exccd.crt`<br />
2. Add exccd.crt to /etc/ca-certificates.conf: `# echo exccd.crt >> /etc/ca-certificates.conf`<br />
3. Update the CA certificate list: `# update-ca-certificates`<br />

**3. Set your mining software url to use https.**<br />

`$ cgminer -o https://127.0.0.1:9109 -u rpcuser -p rpcpassword`

<a name="Help" />

### 3. Help

<a name="NetworkConfig" />

**3.1 Network Configuration**<br />
* [What Ports Are Used by Default?](https://github.com/EXCCoin/exccd/tree/master/docs/default_ports.md)
* [How To Listen on Specific Interfaces](https://github.com/EXCCoin/exccd/tree/master/docs/configure_peer_server_listen_interfaces.md)
* [How To Configure RPC Server to Listen on Specific Interfaces](https://github.com/EXCCoin/exccd/tree/master/docs/configure_rpc_server_listen_interfaces.md)
* [Configuring exccd with Tor](https://github.com/EXCCoin/exccd/tree/master/docs/configuring_tor.md)

<a name="Wallet" />

**3.2 Wallet**<br />

exccd was intentionally developed without an integrated wallet for security
reasons.  Please see [exccwallet](https://github.com/EXCCoin/exccwallet) for more
information.

<a name="Contact" />

### 4. Contact

<a name="ContactCommunity" />

**4.1 Community**<br />

If you have any further questions you can find us at:

https://excc.co/

<a name="DeveloperResources" />

### 5. Developer Resources

<a name="ContributionGuidelines" />

**5.1 Code Contribution Guidelines**

* [Code Contribution Guidelines](https://github.com/EXCCoin/exccd/tree/master/docs/code_contribution_guidelines.md)

<a name="JSONRPCReference" />

**5.2 JSON-RPC Reference**

* [JSON-RPC Reference](https://github.com/EXCCoin/exccd/tree/master/docs/json_rpc_api.mediawiki)
* [RPC Examples](https://github.com/EXCCoin/exccd/tree/master/docs/json_rpc_api.mediawiki#8-example-code)

<a name="GoModules" />

**5.3 Go Modules**

The following versioned modules are provided by exccd repository:

* [rpcclient/v7](https://github.com/EXCCoin/exccd/tree/master/rpcclient) - Implements
  a robust and easy to use Websocket-enabled Exchangecoin JSON-RPC client
* [dcrjson/v4](https://github.com/EXCCoin/exccd/tree/master/dcrjson) - Provides
  infrastructure for working with Exchangecoin JSON-RPC APIs
* [rpc/jsonrpc/types/v3](https://github.com/EXCCoin/exccd/tree/master/rpc/jsonrpc/types) -
  Provides concrete types via dcrjson for the chain server JSON-RPC commands,
  return values, and notifications
* [wire](https://github.com/EXCCoin/exccd/tree/master/wire) - Implements the
  Exchangecoin wire protocol
* [peer/v3](https://github.com/EXCCoin/exccd/tree/master/peer) - Provides a common
  base for creating and managing Exchangecoin network peers
* [blockchain/v4](https://github.com/EXCCoin/exccd/tree/master/blockchain) -
  Implements Exchangecoin block handling and chain selection rules
  * [stake/v4](https://github.com/EXCCoin/exccd/tree/master/blockchain/stake) -
    Provides an API for working with stake transactions and other portions
    related to the Proof-of-Stake (PoS) system
  * [standalone/v2](https://github.com/EXCCoin/exccd/tree/master/blockchain/standalone) -
    Provides standalone functions useful for working with the Exchangecoin blockchain
    consensus rules
* [txscript/v4](https://github.com/EXCCoin/exccd/tree/master/txscript) -
  Implements the Exchangecoin transaction scripting language
* [dcrec](https://github.com/EXCCoin/exccd/tree/master/dcrec) - Provides constants
  for the supported cryptographic signatures supported by Exchangecoin scripts
  * [secp256k1/v4](https://github.com/EXCCoin/exccd/tree/master/dcrec/secp256k1) -
    Implements the secp256k1 elliptic curve
  * [edwards/v2](https://github.com/EXCCoin/exccd/tree/master/dcrec/edwards) -
    Implements the edwards25519 twisted Edwards curve
* [database/v3](https://github.com/EXCCoin/exccd/tree/master/database) -
  Provides a database interface for the Exchangecoin block chain
* [dcrutil/v4](https://github.com/EXCCoin/exccd/tree/master/dcrutil) - Provides
  Exchangecoin-specific convenience functions and types
* [chaincfg/v3](https://github.com/EXCCoin/exccd/tree/master/chaincfg) - Defines
  chain configuration parameters for the standard Exchangecoin networks and allows
  callers to define their own custom Exchangecoin networks for testing purposes
  * [chainhash](https://github.com/EXCCoin/exccd/tree/master/chaincfg/chainhash) -
    Provides a generic hash type and associated functions that allows the
    specific hash algorithm to be abstracted
* [certgen](https://github.com/EXCCoin/exccd/tree/master/certgen) - Provides a
  function for creating a new TLS certificate key pair, typically used for
  encrypting RPC and websocket communications
* [addrmgr/v2](https://github.com/EXCCoin/exccd/tree/master/addrmgr) - Provides a
  concurrency safe Exchangecoin network address manager
* [connmgr/v3](https://github.com/EXCCoin/exccd/tree/master/connmgr) - Implements
  a generic Exchangecoin network connection manager
* [hdkeychain/v3](https://github.com/EXCCoin/exccd/tree/master/hdkeychain) -
  Provides an API for working with  Exchangecoin hierarchical deterministic extended
  keys
* [gcs/v3](https://github.com/EXCCoin/exccd/tree/master/gcs) - Provides an API for
  building and using Golomb-coded set filters useful for light clients such as
  SPV wallets
* [lru](https://github.com/EXCCoin/exccd/tree/master/lru) - Implements a generic
  concurrent safe least-recently-used cache with near O(1) perf
* [container/apbf](https://github.com/EXCCoin/exccd/tree/master/container/apbf) -
  Implements an optimized Age-Partitioned Bloom Filter
* [crypto/blake256](https://github.com/EXCCoin/exccd/tree/master/crypto/blake256) -
  Implements 14-round BLAKE-256 and BLAKE-224 hash functions (SHA-3 candidate)
* [crypto/ripemd160](https://github.com/EXCCoin/exccd/tree/master/crypto/ripemd160) -
   Implements the RIPEMD-160 hash algorithm
* [math/uint256](https://github.com/EXCCoin/exccd/tree/master/math/uint256) -
  Implements highly optimized fixed precision unsigned 256-bit integer
  arithmetic

<a name="ModuleHierarchy" />

**5.4 Module Hierarchy**

The following diagram shows an overview of the hierarchy for the modules
provided by the exccd repository.

![Module Hierarchy](./assets/module_hierarchy.svg)

<a name="SimnetReference" />

**6. Simulation Network (--simnet)**

When developing Exchangecoin applications or testing potential changes, it is often
extremely useful to have a test network where transactions are actually mined
into blocks, difficulty levels are low enough to generate blocks on demand, it
is possible to easily cause chain reorganizations for testing purposes, and
otherwise have full control over the network.

In order to facilitate these scenarios, `exccd` provides a simulation network
(`--simnet`), where the difficulty starts extremely low to enable fast CPU
mining of blocks.  Simnet also has some modified functionality that helps
developers avoid common issues early in development.

See the full reference for more details:

* [Simulation Network Reference](https://github.com/EXCCoin/exccd/tree/master/docs/simnet_environment.mediawiki)
