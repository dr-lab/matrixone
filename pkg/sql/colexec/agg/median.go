// Copyright 2022 Matrix Origin
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

package agg

import (
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"sort"
)

type decimal64Slice []types.Decimal64
type decimal128Slice []types.Decimal128

func (s decimal64Slice) Len() int {
	return len(s)
}

func (s decimal64Slice) Less(i, j int) bool {
	return s[i].Lt(s[j])
}

func (s decimal64Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s decimal128Slice) Len() int {
	return len(s)
}

func (s decimal128Slice) Less(i, j int) bool {
	return s[i].Lt(s[j])
}

func (s decimal128Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type numericSlice[T Numeric] []T

func (s numericSlice[T]) Len() int {
	return len(s)
}

func (s numericSlice[T]) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s numericSlice[T]) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type Median[T Numeric] struct {
	Vals []numericSlice[T]
}

type Decimal64Median struct {
	Vals []decimal64Slice
}
type Decimal128Median struct {
	Vals []decimal128Slice
}

func MedianReturnType(typs []types.Type) types.Type {
	switch typs[0].Oid {
	case types.T_decimal64:
		return types.New(types.T_decimal128, 0, typs[0].Scale, typs[0].Precision)
	case types.T_decimal128:
		return types.New(types.T_decimal128, 0, typs[0].Scale, typs[0].Precision)
	case types.T_float32, types.T_float64:
		return types.New(types.T_float64, 0, 0, 0)
	case types.T_int8, types.T_int16, types.T_int32, types.T_int64:
		return types.New(types.T_float64, 0, 0, 0)
	case types.T_uint8, types.T_uint16, types.T_uint32, types.T_uint64:
		return types.New(types.T_float64, 0, 0, 0)
	default:
		return types.Type{}
	}
}

func NewMedian[T Numeric]() *Median[T] {
	return &Median[T]{}
}

func (m *Median[T]) Grows(cnt int) {
	if len(m.Vals) == 0 {
		m.Vals = make([]numericSlice[T], 0, cnt)
	}
	for i := 0; i < cnt; i++ {
		m.Vals = append(m.Vals, make(numericSlice[T], 0))
	}
}

func (m *Median[T]) Eval(vs []float64) []float64 {
	for i := range vs {
		cnt := len(m.Vals[i])
		if cnt == 0 {
			continue
		}
		if !sort.IsSorted(m.Vals[i]) {
			sort.Sort(m.Vals[i])
		}
		if cnt&1 == 1 {
			vs[i] = float64(m.Vals[i][cnt>>1])
		} else {
			vs[i] = float64(m.Vals[i][cnt>>1]+m.Vals[i][(cnt>>1)-1]) / 2
		}
	}
	return vs
}

func (m *Median[T]) Fill(i int64, value T, _ float64, z int64, isEmpty bool, isNull bool) (float64, bool) {
	if !isNull {
		for j := int64(0); j < z; j++ {
			m.Vals[i] = append(m.Vals[i], value)
		}
		return 0, false
	}
	return 0, isEmpty
}

func (m *Median[T]) Merge(xIndex int64, yIndex int64, _ float64, _ float64, xEmpty bool, yEmpty bool, yMedian any) (float64, bool) {
	if !yEmpty {
		yM := yMedian.(*Median[T])
		if !sort.IsSorted(yM.Vals[yIndex]) {
			sort.Sort(yM.Vals[yIndex])
		}
		if xEmpty {
			m.Vals[xIndex] = append(m.Vals[xIndex], yM.Vals[yIndex]...)
			return 0, false
		}
		newCnt := len(m.Vals[xIndex]) + len(yM.Vals[yIndex])
		newData := make(numericSlice[T], newCnt)
		if !sort.IsSorted(m.Vals[xIndex]) {
			sort.Sort(m.Vals[xIndex])
		}
		merge[T](m.Vals[xIndex], yM.Vals[yIndex], newData, func(a, b T) bool { return a < b })
		m.Vals[xIndex] = newData
		return 0, false
	}

	return 0, xEmpty
}

func (m *Median[T]) MarshalBinary() ([]byte, error) {
	return types.Encode(&m.Vals)
}

func (m *Median[T]) UnmarshalBinary(data []byte) error {
	return types.Decode(data, &m.Vals)
}

func NewD64Median() *Decimal64Median {
	return &Decimal64Median{}
}

func (m *Decimal64Median) Grows(cnt int) {
	if len(m.Vals) == 0 {
		m.Vals = make([]decimal64Slice, 0, cnt)
	}
	for i := 0; i < cnt; i++ {
		m.Vals = append(m.Vals, make(decimal64Slice, 0))
	}
}

func (m *Decimal64Median) Eval(vs []types.Decimal128) []types.Decimal128 {
	for i := range vs {
		cnt := len(m.Vals[i])
		if cnt == 0 {
			continue
		}
		if !sort.IsSorted(m.Vals[i]) {
			sort.Sort(m.Vals[i])
		}
		if cnt&1 == 1 {
			vs[i] = types.Decimal128_FromDecimal64(m.Vals[i][cnt>>1])
		} else {
			a := types.Decimal128_FromDecimal64(m.Vals[i][cnt>>1])
			b := types.Decimal128_FromDecimal64(m.Vals[i][(cnt>>1)-1])
			vs[i] = a.Add(b).DivInt64(2)
		}
	}
	return vs
}

func (m *Decimal64Median) Fill(i int64, value types.Decimal64, ov types.Decimal128, z int64, isEmpty bool, isNull bool) (types.Decimal128, bool) {
	if !isNull {
		for j := int64(0); j < z; j++ {
			m.Vals[i] = append(m.Vals[i], value)
		}
		return types.Decimal128_Zero, false
	}
	return types.Decimal128_Zero, isEmpty
}

func (m *Decimal64Median) Merge(xIndex int64, yIndex int64, _ types.Decimal128, _ types.Decimal128, xEmpty bool, yEmpty bool, yMedian any) (types.Decimal128, bool) {
	if !yEmpty {
		yM := yMedian.(*Decimal64Median)
		if !sort.IsSorted(yM.Vals[yIndex]) {
			sort.Sort(yM.Vals[yIndex])
		}
		if xEmpty {
			m.Vals[xIndex] = append(m.Vals[xIndex], yM.Vals[yIndex]...)
			return types.Decimal128_Zero, false
		}
		newCnt := len(m.Vals[xIndex]) + len(yM.Vals[yIndex])
		newData := make(decimal64Slice, newCnt)
		if !sort.IsSorted(m.Vals[xIndex]) {
			sort.Sort(m.Vals[xIndex])
		}
		merge(m.Vals[xIndex], yM.Vals[yIndex], newData, func(a, b types.Decimal64) bool { return a.Lt(b) })
		m.Vals[xIndex] = newData
		return types.Decimal128_Zero, false
	}

	return types.Decimal128_Zero, xEmpty
}

func (m *Decimal64Median) MarshalBinary() ([]byte, error) {
	return types.Encode(&m.Vals)
}
func (m *Decimal64Median) UnmarshalBinary(dt []byte) error {
	return types.Decode(dt, &m.Vals)
}

func NewD128Median() *Decimal128Median {
	return &Decimal128Median{}
}

func (m *Decimal128Median) Grows(cnt int) {
	if len(m.Vals) == 0 {
		m.Vals = make([]decimal128Slice, 0, cnt)
	}
	for i := 0; i < cnt; i++ {
		m.Vals = append(m.Vals, make(decimal128Slice, 0))
	}
}

func (m *Decimal128Median) Eval(vs []types.Decimal128) []types.Decimal128 {
	for i := range vs {
		cnt := len(m.Vals[i])
		if cnt == 0 {
			continue
		}
		if !sort.IsSorted(m.Vals[i]) {
			sort.Sort(m.Vals[i])
		}
		if cnt&1 == 1 {
			vs[i] = m.Vals[i][cnt>>1]
		} else {
			vs[i] = m.Vals[i][cnt>>1].Add(m.Vals[i][(cnt>>1)-1]).DivInt64(2)
		}
	}
	return vs
}

func (m *Decimal128Median) Fill(i int64, value types.Decimal128, _ types.Decimal128, z int64, isEmpty bool, isNull bool) (types.Decimal128, bool) {
	if !isNull {
		for j := int64(0); j < z; j++ {
			m.Vals[i] = append(m.Vals[i], value)
		}
		return types.Decimal128_Zero, false
	}
	return types.Decimal128_Zero, isEmpty
}

func (m *Decimal128Median) Merge(xIndex int64, yIndex int64, _ types.Decimal128, _ types.Decimal128, xEmpty bool, yEmpty bool, yMedian any) (types.Decimal128, bool) {
	if !yEmpty {
		yM := yMedian.(*Decimal128Median)
		if !sort.IsSorted(yM.Vals[yIndex]) {
			sort.Sort(yM.Vals[yIndex])
		}
		if xEmpty {
			m.Vals[xIndex] = append(m.Vals[xIndex], yM.Vals[yIndex]...)
			return types.Decimal128_Zero, false
		}
		if !sort.IsSorted(m.Vals[xIndex]) {
			sort.Sort(m.Vals[xIndex])
		}
		newCnt := len(m.Vals[xIndex]) + len(yM.Vals[yIndex])
		newData := make(decimal128Slice, newCnt)
		merge(m.Vals[xIndex], yM.Vals[yIndex], newData, func(a, b types.Decimal128) bool { return a.Lt(b) })
		m.Vals[xIndex] = newData
		return types.Decimal128_Zero, false
	}
	return types.Decimal128_Zero, xEmpty
}

func (m *Decimal128Median) MarshalBinary() ([]byte, error) {
	return types.Encode(&m.Vals)
}
func (m *Decimal128Median) UnmarshalBinary(dt []byte) error {
	return types.Decode(dt, &m.Vals)
}

func merge[T Numeric | types.Decimal64 | types.Decimal128](s1, s2, rs []T, lt func(a, b T) bool) []T {
	i, j, cnt := 0, 0, 0
	for i < len(s1) && j < len(s2) {
		if lt(s1[i], s2[j]) {
			rs[cnt] = s1[i]
			i++
		} else {
			rs[cnt] = s2[j]
			j++
		}
		cnt++
	}
	for ; i < len(s1); i++ {
		rs[cnt] = s1[i]
		cnt++
	}
	for ; j < len(s2); j++ {
		rs[cnt] = s2[j]
		cnt++
	}
	return rs
}
