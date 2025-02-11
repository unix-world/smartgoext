// Copyright ©2021 The go-pdf Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// v.20241215.1258
// (c) unix-world.org
// license: BSD

package fpdf

type wbuffer struct {
	p []byte
	c int
}

func (w *wbuffer) u8(v uint8) {
	w.p[w.c] = v
	w.c++
}

func (w *wbuffer) bytes() []byte { return w.p }
