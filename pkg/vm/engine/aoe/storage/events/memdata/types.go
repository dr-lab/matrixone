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

package memdata

import (
	e "matrixone/pkg/vm/engine/aoe/storage"
	bmgrif "matrixone/pkg/vm/engine/aoe/storage/buffer/manager/iface"
	"matrixone/pkg/vm/engine/aoe/storage/layout/base"
	"matrixone/pkg/vm/engine/aoe/storage/layout/table/v1"
	mtif "matrixone/pkg/vm/engine/aoe/storage/memtable/base"
	md "matrixone/pkg/vm/engine/aoe/storage/metadata/v1"
	"matrixone/pkg/vm/engine/aoe/storage/ops"
	"matrixone/pkg/vm/engine/aoe/storage/sched"
)

type Context struct {
	Opts                             *e.Options
	DoneCB                           ops.OpDoneCB
	Waitable                         bool
	Tables                           *table.Tables
	MTMgr                            mtif.IManager
	IndexBufMgr, MTBufMgr, SSTBufMgr bmgrif.IBufferManager
	FsMgr                            base.IManager
	TableMeta                        *md.Table
}

type BaseEvent struct {
	sched.BaseEvent
	Ctx *Context
}
