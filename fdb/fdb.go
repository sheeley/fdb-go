// FoundationDB Go API
// Copyright (c) 2013 FoundationDB, LLC

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package fdb

/*
 #define FDB_API_VERSION 100
 #include <foundationdb/fdb_c.h>
*/
import "C"

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

/* Would put this in futures.go but for the documented issue with
/* exports and functions in preamble
/* (https://code.google.com/p/go-wiki/wiki/cgo#Global_functions) */
//export notifyChannel
func notifyChannel(ch *chan struct{}) {
	*ch <- struct{}{}
}

type Error struct {
	Code C.fdb_error_t
}

func (e Error) Error() string {
	return fmt.Sprintf("%s (%d)", C.GoString(C.fdb_get_error(e.Code)), e.Code)
}

var apiVersion int

func setOpt(setter func(*C.uint8_t, C.int) C.fdb_error_t, param []byte) error {
	if err := setter(byteSliceToPtr(param), C.int(len(param))); err != 0 {
		return Error{Code: err}
	}

	return nil
}

type networkOptions struct{}

func (opt networkOptions) setOpt(code int, param []byte) error {
	return setOpt(func(p *C.uint8_t, pl C.int) C.fdb_error_t {
		return C.fdb_network_set_option(C.FDBNetworkOption(code), p, pl)
	}, param)
}

type api struct {
	Options networkOptions
}

func APIVersion(ver int) (*api, error) {
	if apiVersion != 0 {
		return nil, fmt.Errorf("FoundationDB API already loaded at version %d", apiVersion)
	}
	if e := C.fdb_select_api_version_impl(C.int(ver), 100); e != 0 {
		return nil, fmt.Errorf("FoundationDB API error (requested API 100)")
	}
	apiVersion = ver
	return &api{}, nil
}

var networkStarted bool
var networkMutex sync.Mutex

func (api *api) startNetwork() error {
	if e := C.fdb_setup_network(); e != 0 {
		return Error{Code: e}
	}
	go C.fdb_run_network()

	networkStarted = true

	return nil
}

func (api *api) StartNetwork() error {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	return api.startNetwork()
}

type DBConfig struct {
	ClusterFile string
	DBName []byte
}

func (api *api) Open(conf *DBConfig) (db *Database, e error) {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	if !networkStarted {
		e = api.startNetwork()
		if e != nil {
			return
		}
	}

	var cf *C.char
	if conf != nil {
		cf = C.CString(conf.ClusterFile)
	}
	f := C.fdb_create_cluster(cf)
	fdb_future_block_until_ready(f)
	outc := &C.FDBCluster{}
	if err := C.fdb_future_get_cluster(f, &outc); err != 0 {
		return nil, Error{Code: err}
	}
	C.fdb_future_destroy(f)
	c := &Cluster{c: outc}
	runtime.SetFinalizer(c, (*Cluster).destroy)

	var dbname []byte
	if conf == nil {
		dbname = []byte("DB")
	} else {
		dbname = conf.DBName
	}

	db, e = c.OpenDatabase(dbname)

	return
}

func (api *api) CreateCluster(cluster string) (*Cluster, error) {
	var cf *C.char

	if len(cluster) != 0 {
		cf = C.CString(cluster)
	}

	f := C.fdb_create_cluster(cf)
	fdb_future_block_until_ready(f)
	outc := &C.FDBCluster{}
	if err := C.fdb_future_get_cluster(f, &outc); err != 0 {
		return nil, Error{Code: err}
	}
	C.fdb_future_destroy(f)
	c := &Cluster{c: outc}
	runtime.SetFinalizer(c, (*Cluster).destroy)
	return c, nil
}

func byteSliceToPtr(b []byte) *C.uint8_t {
	if len(b) > 0 {
		return (*C.uint8_t)(unsafe.Pointer(&b[0]))
	} else {
		return nil
	}
}
