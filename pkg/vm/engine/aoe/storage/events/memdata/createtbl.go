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
	imem "matrixone/pkg/vm/engine/aoe/storage/memtable/base"
	"matrixone/pkg/vm/engine/aoe/storage/sched"
	// log "github.com/sirupsen/logrus"
)

type createTableEvent struct {
	BaseEvent
	Collection imem.ICollection
}

func NewCreateTableEvent(ctx *Context) *createTableEvent {
	e := &createTableEvent{}
	e.BaseEvent = BaseEvent{
		BaseEvent: *sched.NewBaseEvent(e, sched.MemdataUpdateEvent, ctx.DoneCB, ctx.Waitable),
		Ctx:       ctx,
	}
	return e
}

func (e *createTableEvent) Execute() error {
	collection := e.Ctx.MTMgr.StrongRefCollection(e.Ctx.TableMeta.ID)
	if collection != nil {
		e.Collection = collection
		return nil
	}
	meta := e.Ctx.TableMeta

	tableData, err := e.Ctx.Tables.StrongRefTable(meta.ID)
	if err != nil {
		tableData, err = e.Ctx.Tables.RegisterTable(meta)
		if err != nil {
			return err
		}
		tableData.Ref()
	}
	collection, err = e.Ctx.MTMgr.RegisterCollection(tableData)
	if err != nil {
		tableData.Unref()
		return err
	}

	e.Collection = collection

	return nil
}
