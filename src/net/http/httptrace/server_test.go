// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httptrace

import (
	"context"
	"testing"
)

func TestWithServerTrace(t *testing.T) {
	ctx := context.Background()
	newtrace := &ServerTrace{}

	ctx = WithServerTrace(ctx, newtrace)
	trace := ContextServerTrace(ctx)

	if trace != newtrace {
		t.Errorf("got %q; want %q", trace, newtrace)
	}
}
