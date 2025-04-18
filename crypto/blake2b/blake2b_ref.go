// The IxiosSpark library is distributed with the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// Copyright 2016 The Golang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !amd64 || appengine || gccgo
// +build !amd64 appengine gccgo

package blake2b

func f(h *[8]uint64, m *[16]uint64, c0, c1 uint64, flag uint64, rounds uint64) {
	fGeneric(h, m, c0, c1, flag, rounds)
}
