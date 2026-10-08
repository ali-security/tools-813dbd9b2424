// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"
)

const ExampleOffset = exampleOffset

// TestServeNoPortFlag checks that the serve command no longer offers the
// -port flag, which listened for remote connections on all network
// interfaces.
func TestServeNoPortFlag(t *testing.T) {
	typ := reflect.TypeOf(Serve{})
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Tag.Get("flag") == "port" {
			t.Errorf("Serve.%s defines the -port flag, which listens on all network interfaces", field.Name)
		}
	}
}

// runServe runs s and returns its result, failing the test if it does not
// return promptly (for example because it is serving connections).
func runServe(t *testing.T, s *Serve) error {
	t.Helper()
	errc := make(chan error, 1)
	go func() {
		errc <- s.Run(context.Background())
	}()
	select {
	case err := <-errc:
		return err
	case <-time.After(10 * time.Second):
		t.Fatalf("serve -listen=%s did not return: the server is accepting connections", s.Address)
		return nil
	}
}

// TestServeRejectsImplicitAllInterfaces checks that -listen refuses an
// address with an empty host, which would implicitly bind all network
// interfaces.
func TestServeRejectsImplicitAllInterfaces(t *testing.T) {
	for _, addr := range []string{":0", ":"} {
		err := runServe(t, &Serve{Address: addr, app: &Application{}})
		if err == nil || !strings.Contains(err.Error(), "implicitly binds all network interfaces") {
			t.Errorf("serve -listen=%s: got error %v, want implicit all-interfaces error", addr, err)
		}
	}
}

// TestServeAllowsExplicitHost checks that -listen still accepts an address
// with an explicit host. An out-of-range port makes the listen itself fail,
// so the server never starts accepting connections.
func TestServeAllowsExplicitHost(t *testing.T) {
	addr := "127.0.0.1:99999"
	err := runServe(t, &Serve{Address: addr, app: &Application{}})
	if err == nil {
		t.Fatalf("serve -listen=%s: got nil error, want listen error", addr)
	}
	if strings.Contains(err.Error(), "implicitly binds all network interfaces") {
		t.Errorf("serve -listen=%s: explicit host rejected: %v", addr, err)
	}
}
