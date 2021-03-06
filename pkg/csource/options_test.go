// Copyright 2017 syzkaller project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package csource

import (
	"fmt"
	"reflect"
	"testing"
)

//{Threaded:true Collide:true Repeat:true Procs:1 Sandbox:none Fault:false FaultCall:-1 FaultNth:0 EnableTun:true UseTmpDir:true HandleSegv:true WaitRepeat:true Debug:false Repro:false}

func TestParseOptions(t *testing.T) {
	for _, opts := range allOptionsSingle() {
		data := opts.Serialize()
		got, err := DeserializeOptions(data)
		if err != nil {
			t.Fatalf("failed to deserialize %q: %v", data, err)
		}
		if !reflect.DeepEqual(got, opts) {
			t.Fatalf("opts changed, got:\n%+v\nwant:\n%+v", got, opts)
		}
	}
}

func TestParseOptionsCanned(t *testing.T) {
	// Dashboard stores csource options with syzkaller reproducers,
	// so we need to be able to parse old formats.
	canned := map[string]Options{
		"{Threaded:true Collide:true Repeat:true Procs:1 Sandbox:none Fault:false FaultCall:-1 FaultNth:0 EnableTun:true UseTmpDir:true HandleSegv:true WaitRepeat:true Debug:false Repro:false}": Options{
			Threaded:   true,
			Collide:    true,
			Repeat:     true,
			Procs:      1,
			Sandbox:    "none",
			Fault:      false,
			FaultCall:  -1,
			FaultNth:   0,
			EnableTun:  true,
			UseTmpDir:  true,
			HandleSegv: true,
			WaitRepeat: true,
			Debug:      false,
			Repro:      false,
		},
		"{Threaded:true Collide:true Repeat:true Procs:1 Sandbox: Fault:false FaultCall:-1 FaultNth:0 EnableTun:true UseTmpDir:true HandleSegv:true WaitRepeat:true Debug:false Repro:false}": Options{
			Threaded:   true,
			Collide:    true,
			Repeat:     true,
			Procs:      1,
			Sandbox:    "",
			Fault:      false,
			FaultCall:  -1,
			FaultNth:   0,
			EnableTun:  true,
			UseTmpDir:  true,
			HandleSegv: true,
			WaitRepeat: true,
			Debug:      false,
			Repro:      false,
		},
	}
	for data, want := range canned {
		got, err := DeserializeOptions([]byte(data))
		if err != nil {
			t.Fatalf("failed to deserialize %q: %v", data, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("deserialize %q\ngot:\n%+v\nwant:\n%+v", data, got, want)
		}
	}
}

func allOptionsSingle() []Options {
	var opts []Options
	fields := reflect.TypeOf(Options{}).NumField()
	for i := 0; i < fields; i++ {
		opts = append(opts, enumerateField(Options{}, i)...)
	}
	return opts
}

func allOptionsPermutations() []Options {
	opts := []Options{Options{}}
	fields := reflect.TypeOf(Options{}).NumField()
	for i := 0; i < fields; i++ {
		var newOpts []Options
		for _, opt := range opts {
			newOpts = append(newOpts, enumerateField(opt, i)...)
		}
		opts = newOpts
	}
	return opts
}

func enumerateField(opt Options, field int) []Options {
	var opts []Options
	s := reflect.ValueOf(&opt).Elem()
	fldName := s.Type().Field(field).Name
	fld := s.Field(field)
	if fldName == "Sandbox" {
		for _, sandbox := range []string{"", "none", "setuid", "namespace"} {
			fld.SetString(sandbox)
			opts = append(opts, opt)
		}
	} else if fldName == "Procs" {
		for _, procs := range []int64{1, 4} {
			fld.SetInt(procs)
			opts = append(opts, opt)
		}
	} else if fldName == "FaultCall" {
		opts = append(opts, opt)
	} else if fldName == "FaultNth" {
		opts = append(opts, opt)
	} else if fld.Kind() == reflect.Bool {
		for _, v := range []bool{false, true} {
			fld.SetBool(v)
			opts = append(opts, opt)
		}
	} else {
		panic(fmt.Sprintf("field '%v' is not boolean", fldName))
	}
	var checked []Options
	for _, opt := range opts {
		if err := opt.Check(); err == nil {
			checked = append(checked, opt)
		}
	}
	return checked
}
