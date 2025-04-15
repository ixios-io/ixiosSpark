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
	"enode://7ef7cdb36148b99e8a8f09d9ac42a332b6747adf86280caf00b12ca9fb603f34051cc110e27712ea9d430817877802f41fd5caa40097074eacec16fdc67f526e@152.42.210.116:38383", // sgp-ixios-bootnode
	"enode://7fb0fec81c863e0a762912833d4d0150b7e05fdd4c9410ef899dd3c99a58bc2c7aad5d54b8fc3793bb02e8f3c98ccfbcd31eb07c44e49a1ff82e33af22cb0589@18.163.48.248:38383",  // hkg-ixios-bootnode
	"enode://06bb3252f81d4fdbdd5a5544ffeb9a9160f6c9d1c787bfc9f9b71756de7eb49a0f143fd427de883431f5ea266bc3275ed7c880a79c554362c8a37faab91ea7bf@3.28.197.2:38383",     // uae-ixios-bootnode
	"enode://0b27cf61c40133a246cd3f42d93b86db6427eaf7cec83ea5d113ab9fbb52eafc54c7d605b910339924f95844bb14d9a074df18e61d6b8b61f2f788d86de49b58@128.199.4.68:38383",   // sfo3-ixios-bootnode
	"enode://1aa3b4e06463f46c6a44e83bd21369bbbd8c1481e35fa5f30f383171208aa93e5b36cdf61d21e33bb513c2b076b582751d4aeb90081c4f15602dc81f43c0c978@174.138.14.204:38383", // ams3-ixios-bootnode
	"enode://371e0994f61b1aa67432e6099eea74d28b508ab413106df31d5b03c3f10d0ace190ff602bceaa9b79aac8f504bab05d4d7e85f0587b6a6c67b7e4675e8abefb5@16.63.31.191:38383",   // zrh-ixios-bootnode
}

// AetherBloomBootnodes are the enode URLs of the P2P bootstrap nodes running on the AetherBloom test network.
var AetherBloomBootnodes = []string{
	"enode://0354498b2ba17e29b9d442baa29087aea34a1655d7628c5760a115752ee81105442bf4ed638cd464bd6ee8b41bdeeed469c863c5e1e25cb2f4ceff7b603ebe59@178.128.51.186:38383", // aetherbloom-bootnode (SGP1)
}

// AetherForgeBootnodes are the enode URLs of the P2P bootstrap nodes running on the AetherForge test network.
var AetherForgeBootnodes = []string{
	"enode://6ee6a56c752ea0f5f4290a8ab52c8a486767855a8e379652ac22e6fa53c53e05f1116d73a1efbd3f7721486d7e1bbc1d91b96026887f821675723bd52d5fd1ca@209.97.173.155:38383", // aetherforge-bootnode (SGP1)
}

// NeoDawnBootnodes are the enode URLs of the P2P bootstrap nodes running on the NeoDawn test network.
var NeoDawnBootnodes = []string{
	// todo
}

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
	case NeoDawnGenesisHash:
		net = "neodawn"
	default:
		return ""
	}
	return dnsPrefix + protocol + "." + net + ".ixios.org"
}
