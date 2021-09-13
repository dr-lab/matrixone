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

package engine

import (
	log "github.com/sirupsen/logrus"
	"matrixone/pkg/catalog"
	"matrixone/pkg/logutil"
	"matrixone/pkg/vm/engine"
	"matrixone/pkg/vm/engine/aoe/common/codec"
	"matrixone/pkg/vm/engine/aoe/common/helper"
	adb "matrixone/pkg/vm/engine/aoe/storage/db"
	"matrixone/pkg/vm/metadata"
	"time"
)

func (db *database) Type() int {
	return db.typ
}

func (db *database) Delete(epoch uint64, name string) error {
	t0 := time.Now()
	defer func() {
		logutil.Debugf("time cost %d ms", time.Since(t0).Milliseconds())
	}()
	_, err := db.catalog.DropTable(epoch, db.id, name)
	return err
}

func (db *database) Create(epoch uint64, name string, defs []engine.TableDef, pdef *engine.PartitionBy, _ *engine.DistributionBy, comment string) error {
	t0 := time.Now()
	defer func() {
		logutil.Debugf("time cost %d ms", time.Since(t0).Milliseconds())
	}()
	tbl, err := helper.Transfer(db.id, 0, 0, name, comment, defs, pdef)
	if err != nil {
		return err
	}
	_, err = db.catalog.CreateTable(epoch, db.id, tbl)
	return err
}

func (db *database) Relations() []string {
	t0 := time.Now()
	defer func() {
		logutil.Debugf("time cost %d ms", time.Since(t0).Milliseconds())
	}()
	var rs []string
	tbs, err := db.catalog.ListTables(db.id)
	if err != nil {
		return rs
	}
	for _, tb := range tbs {
		rs = append(rs, tb.Name)
	}
	return rs
}

func (db *database) Relation(name string) (engine.Relation, error) {
	t0 := time.Now()
	defer func() {
		logutil.Debugf("time cost %d ms", time.Since(t0).Milliseconds())
	}()
	tablets, err := db.catalog.GetTablets(db.id, name)
	if err != nil {
		return nil, err
	}
	if tablets == nil || len(tablets) == 0 {
		return nil, catalog.ErrTableNotExists
	}

	r := &relation{
		pid:     db.id,
		tbl:     &tablets[0].Table,
		catalog: db.catalog,
		mp:      make(map[string]*adb.Relation),
	}
	r.tablets = tablets
	ldb := db.catalog.Driver.AOEStore()
	for _, tbl := range tablets {
		if ids, err := db.catalog.Driver.GetSegmentIds(tbl.Name, tbl.ShardId); err != nil {
			log.Errorf("get segmentInfos for tablet %s failed, %s", tbl.Name, err.Error())
		} else {
			if len(ids.Ids) == 0 {
				continue
			}
			addr := db.catalog.Driver.RaftStore().GetRouter().LeaderPeerStore(tbl.ShardId).ClientAddr
			if lRelation, err := ldb.Relation(tbl.Name); err == nil {
				r.mp[tbl.Name] = lRelation
			}
			for _, id := range ids.Ids {
				r.segments = append(r.segments, engine.SegmentInfo{
					Version:  ids.Version,
					Id:       string(codec.Uint642Bytes(id)),
					GroupId:  string(codec.Uint642Bytes(tbl.ShardId)),
					TabletId: tbl.Name,
					Node: metadata.Node{
						Id:   addr,
						Addr: addr,
					},
				})
			}

		}
	}
	return r, nil
}
