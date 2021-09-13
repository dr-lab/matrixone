// Copyright 2021 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dio

import (
	"context"
	"fmt"
	e "matrixone/pkg/vm/engine/aoe/storage"
	base "matrixone/pkg/vm/engine/aoe/storage/dataio/iface"
	// log "github.com/sirupsen/logrus"
)

var (
	READER_FACTORY = &ReaderFactory{
		Builders: make(map[string]base.ReaderBuilder),
	}
)

type ReaderFactory struct {
	Opts     *e.Options
	Dirname  string
	Builders map[string]base.ReaderBuilder
}

func (wf *ReaderFactory) GetOpts() *e.Options {
	return wf.Opts
}

func (wf *ReaderFactory) GetDir() string {
	return wf.Dirname
}

func (wf *ReaderFactory) Init(opts *e.Options, dirname string) {
	wf.Opts = opts
	wf.Dirname = dirname
}

func (wf *ReaderFactory) RegisterBuilder(name string, rb base.ReaderBuilder) {
	_, ok := wf.Builders[name]
	if ok {
		panic(fmt.Sprintf("Duplicate reader %s found", name))
	}
	wf.Builders[name] = rb
}

func (wf *ReaderFactory) MakeReader(name string, ctx context.Context) base.Reader {
	rb, ok := wf.Builders[name]
	if !ok {
		return nil
	}
	return rb.Build(wf, ctx)
}
