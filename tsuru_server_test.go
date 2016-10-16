// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

var tsuruServer interface {
	url() string
	reset()
} = newFakeTsuruServer()
