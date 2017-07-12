// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package httptrace provides mechanisms to trace the events within
// HTTP client requests and server responses.
package httptrace

import (
	"context"
	"reflect"
)

// unique type to prevent assignment.
type serverEventContextKey struct{}

// ContextServerTrace returns the ServerTrace associated with the
// provided context. If none, it returns nil.
func ContextServerTrace(ctx context.Context) *ServerTrace {
	trace, _ := ctx.Value(serverEventContextKey{}).(*ServerTrace)
	return trace
}

// WithServerTrace returns a new context based on the provided parent
// ctx. HTTP server responses made with the returned context will use
// the provided trace hooks, in addition to any previous hooks
// registered with ctx. Any hooks defined in the provided trace will
// be called first.
func WithServerTrace(ctx context.Context, trace *ServerTrace) context.Context {
	if trace == nil {
		panic("nil trace")
	}
	old := ContextServerTrace(ctx)
	trace.compose(old)

	return context.WithValue(ctx, serverEventContextKey{}, trace)
}

// ServerTrace is a set of hooks to run at various stages of an ongoing
// HTTP response. Any particular hook may be nil. Functions may be
// called concurrently from different goroutines and some may be called
// after the request has completed or failed.
type ServerTrace struct {
}

// compose modifies t such that it respects the previously-registered hooks in old,
// subject to the composition policy requested in t.Compose.
func (t *ServerTrace) compose(old *ServerTrace) {
	if old == nil {
		return
	}
	tv := reflect.ValueOf(t).Elem()
	ov := reflect.ValueOf(old).Elem()
	structType := tv.Type()
	for i := 0; i < structType.NumField(); i++ {
		tf := tv.Field(i)
		hookType := tf.Type()
		if hookType.Kind() != reflect.Func {
			continue
		}
		of := ov.Field(i)
		if of.IsNil() {
			continue
		}
		if tf.IsNil() {
			tf.Set(of)
			continue
		}

		// Make a copy of tf for tf to call. (Otherwise it
		// creates a recursive call cycle and stack overflows)
		tfCopy := reflect.ValueOf(tf.Interface())

		// We need to call both tf and of in some order.
		newFunc := reflect.MakeFunc(hookType, func(args []reflect.Value) []reflect.Value {
			tfCopy.Call(args)
			return of.Call(args)
		})
		tv.Field(i).Set(newFunc)
	}
}
