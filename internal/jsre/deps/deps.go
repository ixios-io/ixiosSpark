// Copyright 2017 The go-ethereum Authors
// This file is part of the IxiosSpark library, which builds upon the source code of the go-ethereum library.
//
// The IxiosSpark library, including the go-ethereum library source code it is based on, is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The IxiosSpark and go-ethereum library source code are distributed with the hope that they will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package deps contains the console JavaScript dependencies Go embedded.
package deps

import (
	_ "embed"
)

//go:embed web3.js
var Web3JS string

//go:embed bignumber.js
var BigNumberJS string
