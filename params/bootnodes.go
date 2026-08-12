// IxiosSpark is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// This file is part of the IxiosSpark library, which builds upon the source code of the geth library.
// The IxiosSpark source code is distributed with the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
// Copyright 2025 The ixiosSpark Authors, Copyright 2015-2024 The go-ethereum Authors (geth)
// You should have received a copy of the GNU Lesser General Public License
// with IxiosSpark. If not, see <http://www.gnu.org/licenses/>.

package params

import "github.com/ixios-io/ixiosSpark/common"

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the mainnet Ixios network.
var MainnetBootnodes = []string{
	"enode://7ef7cdb36148b99e8a8f09d9ac42a332b6747adf86280caf00b12ca9fb603f34051cc110e27712ea9d430817877802f41fd5caa40097074eacec16fdc67f526e@104.248.97.10:38383",
	"enode://4e7d9ffe77bb1ffd98cf8fda252865003b7c9fce4847928f522f526a7dde6277c37f4c63fd44013c863ee7b046b99e21e92ba5a5a4b0ff2ecf879dcb1c1e371d@143.198.242.92:38383",
	"enode://fb6a1a7858093c3d8cf891a19822a8eea40d45406225e227e1cc0a9e39e1d3f0e36cd7395904bb18c325925a7439f981ab3bf0f0d04cc5f90a401eaba9acb4a5@164.90.247.229:38383",
}

// AetherBloomBootnodes are the enode URLs of the P2P bootstrap nodes running on the AetherBloom test network.
var AetherBloomBootnodes = []string{}

// AetherForgeBootnodes are the enode URLs of the P2P bootstrap nodes running on the AetherForge test network.
var AetherForgeBootnodes = []string{
	"enode://4bf11e0099e77db3394c5d037a59f11e68b7d5f30a322d77315600e289c007c9164e6323d4d40ece50c8393ff2bad9a78af7c5a76cb81a3a5343fd70253a4300@178.128.62.251:38383",
	"enode://8971b6514b5f6980abc24efc58726d5d24f10e50a830cdb536add5add7f734af544461b4419e446aa69339c0519ab424a28a59e0333c74401263cc2c05899da2@137.184.4.22:38383",
}

// AetherNexusBootnodes are the enode URLs of the P2P bootstrap nodes running on the AetherNexus test network.
var AetherNexusBootnodes = []string{}

var V5Bootnodes = []string{}

// todo
const dnsPrefix = "enrtree://@"

// KnownDNSNetwork returns the address of a public DNS-based node list for the given
// genesis hash and protocol.
func KnownDNSNetwork(genesis common.Hash, protocol string) string {
	return "" // todo

	var net string
	switch genesis {
	case MainnetGenesisHash:
		net = "mainnet"
	case AetherForgeGenesisHash:
		net = "aetherForge"
	case AetherBloomGenesisHash:
		net = "aetherBloom"
	case AetherNexusGenesisHash:
		net = "aetherNexus"
	default:
		return ""
	}
	return dnsPrefix + protocol + "." + net + ".ixios.org"
}
