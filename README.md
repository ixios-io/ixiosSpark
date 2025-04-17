![Ixios](https://ixios.io/img/ixios-logo.svg)

##  IxiosSpark

A reference implementation of a client for the **Ixios** protocol.

<!--[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=ixiosSpark&metric=alert_status)](https://sonarcloud.io/dashboard?id=ixiosSpark)-->
[![Travis](https://app.travis-ci.com/ixios/ixiosSpark.svg?branch=master)](https://app.travis-ci.com/github/ixios-io/ixiosSpark)
[![Discord](https://img.shields.io/badge/discord-join%20server-7289da.svg)](https://discord.gg/ixios)
![Reference](
https://pkg.go.dev/badge/github.com/ixios-io/ixiosSpark
)

-----

**Ixios** is a highly performant distributed ledger protocol. It supports modular digital signature schemes, high throughput, and fast block times.
The goal of Ixios is to facilitate a highly secure, quantum-resistant, and scalable base layer protocol for which future-proof decentralised applications can be built upon. 
Learn more at https://ixios.io

-----

## Building from source

Building `ixiosSpark` requires Go v1.24.0.

A build script for Ubuntu 24.04 LTS and Ubuntu 24.10 has been provided which will automatically install all dependencies and build ixiosSpark.

```shell
./build.sh
```

After building from source, ixiosSpark can be installed with:

```shell
sudo ./install.sh
```

For other operating systems, we suggest using a pre-compiled binary or docker. See `Quick Install` or `Docker Install`.


-----


## Running `ixiosSpark`

### Hardware Requirements

Minimum:

* Fast CPU with 8 cores
* 16GB RAM
* High performance SSD with 2TB free storage space
* 16 Mbps Internet service

Recommended:

* Fast CPU with 16+ cores
* 32GB+ RAM
* High-performance SSD or NVMe with at least 5TB of free space
* 100+ Mbps Internet service

### Firewall Settings

Ensure the following port is open for inbound/outbound:
* 38383 (TCP, UDP)

### Programmatically interfacing `ixiosSpark` nodes


`ixiosSpark` nodes can be accessed programmatically rather than relying on manual console commands. ixiosSpark supports JSON-RPC APIs through HTTP, WebSockets, and IPC.

The IPC interface is enabled by default with full API access. For security, HTTP and WebSocket interfaces are disabled by default and provide limited API access when enabled. You can customise each of these interfaces according to your requirements.

IPC JSON-RPC API options:

* --ipcapi: Defines the APIs accessible over the IPC-RPC interface (default: admin,debug,eth,sealer,net,personal,txpool,web3).
* --ipcpath: Sets the file path for the IPC socket or pipe within the data directory (explicit paths override this).
* --ipcdisable: Disables the IPC-RPC server.

HTTP JSON-RPC API options:

* --http: Activates the HTTP-RPC server.
* --http.addr: Specifies the interface for the HTTP-RPC server to listen on (default: localhost).
* --http.port: Sets the port for the HTTP-RPC server to listen on (default: 8586).
* --http.api: Defines the APIs accessible over the HTTP-RPC interface (default: ixios,net,web3).
* --http.corsdomain: A comma-separated list of domains allowed for cross-origin requests (enforced by browsers).

Websocket JSON-RPC API options:
* --ws: Activates the WebSocket-RPC server.
* --ws.addr: Specifies the interface for the WebSocket-RPC server to listen on (default: localhost).
* --ws.port: Sets the port for the WebSocket-RPC server to listen on (default: 8587).
* --ws.api: Defines the APIs accessible over the WebSocket-RPC interface (default: ixios,net,web3).
* --ws.origins: Specifies the allowed origins for WebSocket requests.

To interface with an ixiosSpark node that's been set up with these options, you'll need to use programming libraries or tools in your development environment to connect via HTTP, WebSocket, or IPC. All communications must follow the JSON-RPC 2.0 standard (https://www.jsonrpc.org/specification) regardless of connection method. You can efficiently send multiple requests over a single connection.

Security Note: When enabling HTTP or WebSocket connections to your `ixiosSpark` node, exercise extreme caution. Publicly accessible APIs attract malicious actors who may attempt to compromise your node. Additionally, locally running web servers can be accessed by any browser tab on your device, meaning malicious websites could potentially interact with your local APIs without your knowledge.

## License

The `ixiosSpark` library is licensed under the [GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html), see `COPYING.LESSER` file.

The ixiosSpark library contains code from go-ethereum (geth), licensed under the [GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html) where specified, see `COPYING.LESSER` file; licensed under the [GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html) where specified, see `COPYING` file.

Additionally, some components of this repo are licensed with the following additional LICENSE files:
* ./crypto/secp256k1/LICENSE
* ./crypto/ecies/LICENSE
* ./crypto/bn256/cloudflare/LICENSE
* ./crypto/bn256/LICENSE
* ./common/bitutil/LICENSE
* ./common/prque/LICENSE
