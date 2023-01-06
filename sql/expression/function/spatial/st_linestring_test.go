// Copyright 2023 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spatial

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
)

func TestStartPoint(t *testing.T) {
	t.Run("simple case", func(t *testing.T) {
		require := require.New(t)
		s := sql.Point{X: 123, Y: 456}
		e := sql.Point{X: 456, Y: 789}
		l := sql.LineString{Points: []sql.Point{s, e}}
		f := NewStartPoint(expression.NewLiteral(l, sql.LineStringType{}))
		v, err := f.Eval(sql.NewEmptyContext(), nil)
		require.NoError(err)
		require.Equal(s, v)
	})

	t.Run("simple case with SRID", func(t *testing.T) {
		require := require.New(t)
		s := sql.Point{SRID: sql.GeoSpatialSRID, X: 123, Y: 456}
		e := sql.Point{SRID: sql.GeoSpatialSRID, X: 456, Y: 789}
		l := sql.LineString{Points: []sql.Point{s, e}}
		f := NewStartPoint(expression.NewLiteral(l, sql.LineStringType{}))
		v, err := f.Eval(sql.NewEmptyContext(), nil)
		require.NoError(err)
		require.Equal(s, v)
	})
}

func TestEndPoint(t *testing.T) {
	t.Run("simple case", func(t *testing.T) {
		require := require.New(t)
		s := sql.Point{X: 123, Y: 456}
		e := sql.Point{X: 456, Y: 789}
		l := sql.LineString{Points: []sql.Point{s, e}}
		f := NewEndPoint(expression.NewLiteral(l, sql.LineStringType{}))
		v, err := f.Eval(sql.NewEmptyContext(), nil)
		require.NoError(err)
		require.Equal(e, v)
	})

	t.Run("simple case with SRID", func(t *testing.T) {
		require := require.New(t)
		s := sql.Point{SRID: sql.GeoSpatialSRID, X: 123, Y: 456}
		e := sql.Point{SRID: sql.GeoSpatialSRID, X: 456, Y: 789}
		l := sql.LineString{Points: []sql.Point{s, e}}
		f := NewEndPoint(expression.NewLiteral(l, sql.LineStringType{}))
		v, err := f.Eval(sql.NewEmptyContext(), nil)
		require.NoError(err)
		require.Equal(e, v)
	})
}

func TestIsClosed(t *testing.T) {
	t.Run("simple case is closed", func(t *testing.T) {
		require := require.New(t)
		s := sql.Point{X: 0, Y: 0}
		p1 := sql.Point{X: 1, Y: 1}
		p2 := sql.Point{X: 2, Y: 2}
		p3 := sql.Point{X: 3, Y: 3}
		l := sql.LineString{Points: []sql.Point{s, p1, p2, p3, s}}
		f := NewIsClosed(expression.NewLiteral(l, sql.LineStringType{}))
		v, err := f.Eval(sql.NewEmptyContext(), nil)
		require.NoError(err)
		require.Equal(true, v)
	})

	t.Run("simple case with SRID is closed", func(t *testing.T) {
		require := require.New(t)
		s := sql.Point{SRID: sql.GeoSpatialSRID, X: 0, Y: 0}
		p1 := sql.Point{SRID: sql.GeoSpatialSRID, X: 1, Y: 1}
		p2 := sql.Point{SRID: sql.GeoSpatialSRID, X: 2, Y: 2}
		p3 := sql.Point{SRID: sql.GeoSpatialSRID, X: 3, Y: 3}
		l := sql.LineString{Points: []sql.Point{s, p1, p2, p3, s}}
		f := NewIsClosed(expression.NewLiteral(l, sql.LineStringType{}))
		v, err := f.Eval(sql.NewEmptyContext(), nil)
		require.NoError(err)
		require.Equal(true, v)
	})

	t.Run("simple case is not closed", func(t *testing.T) {
		require := require.New(t)
		s := sql.Point{X: 0, Y: 0}
		p1 := sql.Point{X: 1, Y: 1}
		p2 := sql.Point{X: 2, Y: 2}
		p3 := sql.Point{X: 3, Y: 3}
		l := sql.LineString{Points: []sql.Point{s, p1, p2, p3}}
		f := NewIsClosed(expression.NewLiteral(l, sql.LineStringType{}))
		v, err := f.Eval(sql.NewEmptyContext(), nil)
		require.NoError(err)
		require.Equal(false, v)
	})

	t.Run("zero length linestring is closed", func(t *testing.T) {
		require := require.New(t)
		s := sql.Point{X: 0, Y: 0}
		l := sql.LineString{Points: []sql.Point{s, s, s}}
		f := NewIsClosed(expression.NewLiteral(l, sql.LineStringType{}))
		v, err := f.Eval(sql.NewEmptyContext(), nil)
		require.NoError(err)
		require.Equal(true, v)
	})
}
