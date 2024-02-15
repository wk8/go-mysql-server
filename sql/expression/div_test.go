// Copyright 2022 Dolthub, Inc.
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

package expression

import (
	"fmt"
	"testing"

	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/src-d/go-errors.v1"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
)

func TestDiv(t *testing.T) {
	var testCases = []struct {
		name  string
		left  sql.Expression
		right sql.Expression
		exp   interface{}
		err   *errors.Kind
		skip  bool
	}{
		{
			left:  NewLiteral(1, types.Int64),
			right: NewLiteral(0, types.Int64),
			exp:   nil,
		},

		// Unsigned Integers
		{
			left:  NewLiteral(1, types.Uint32),
			right: NewLiteral(1, types.Uint32),
			exp:   "1.0000",
		},
		{
			left:  NewLiteral(1, types.Uint32),
			right: NewLiteral(2, types.Uint32),
			exp:   "0.5000",
		},
		{
			left:  NewLiteral(1, types.Uint64),
			right: NewLiteral(1, types.Uint64),
			exp:   "1.0000",
		},
		{
			left:  NewLiteral(1, types.Uint64),
			right: NewLiteral(2, types.Uint64),
			exp:   "0.5000",
		},

		// Signed Integers
		{
			left:  NewLiteral(1, types.Int32),
			right: NewLiteral(1, types.Int32),
			exp:   "1.0000",
		},
		{
			left:  NewLiteral(1, types.Int32),
			right: NewLiteral(2, types.Int32),
			exp:   "0.5000",
		},
		{
			left:  NewLiteral(-1, types.Int32),
			right: NewLiteral(2, types.Int32),
			exp:   "-0.5000",
		},
		{
			left:  NewLiteral(1, types.Int32),
			right: NewLiteral(-2, types.Int32),
			exp:   "-0.5000",
		},
		{
			left:  NewLiteral(1, types.Int64),
			right: NewLiteral(1, types.Int64),
			exp:   "1.0000",
		},
		{
			left:  NewLiteral(1, types.Int64),
			right: NewLiteral(2, types.Int64),
			exp:   "0.5000",
		},
		{
			left:  NewLiteral(-1, types.Int64),
			right: NewLiteral(2, types.Int64),
			exp:   "-0.5000",
		},
		{
			left:  NewLiteral(1, types.Int64),
			right: NewLiteral(-2, types.Int64),
			exp:   "-0.5000",
		},

		// Unsigned and Signed Integers
		{
			left:  NewLiteral(1, types.Uint32),
			right: NewLiteral(-2, types.Int32),
			exp:   "-0.5000",
		},
		{
			left:  NewLiteral(-1, types.Int64),
			right: NewLiteral(2, types.Uint32),
			exp:   "-0.5000",
		},
		{
			left:  NewLiteral(1, types.Int64),
			right: NewLiteral(123456789, types.Int64),
			exp:   "0.0000",
		},

		// Repeating Decimals
		{
			left:  NewLiteral(1, types.Int64),
			right: NewLiteral(3, types.Int64),
			exp:   "0.3333",
		},
		{
			left:  NewLiteral(1, types.Int64),
			right: NewLiteral(9, types.Int64),
			exp:   "0.1111",
		},
		{
			left:  NewLiteral(1, types.Int64),
			right: NewLiteral(6, types.Int64),
			exp:   "0.1667",
		},

		// Floats
		{
			left:  NewLiteral(1.0, types.Float32),
			right: NewLiteral(3.0, types.Float32),
			exp:   0.3333333333333333,
		},
		{
			left:  NewLiteral(1.0, types.Float32),
			right: NewLiteral(9.0, types.Float32),
			exp:   0.1111111111111111,
		},
		{
			left:  NewLiteral(1.0, types.Float64),
			right: NewLiteral(3.0, types.Float64),
			exp:   0.3333333333333333,
		},
		{
			left:  NewLiteral(1.0, types.Float64),
			right: NewLiteral(9.0, types.Float64),
			exp:   0.1111111111111111,
		},
		{
			// MySQL treats float32 a little differently
			skip:  true,
			left:  NewLiteral(3.14159, types.Float32),
			right: NewLiteral(3.0, types.Float32),
			exp:   1.0471967061360676,
		},
		{
			left:  NewLiteral(3.14159, types.Float64),
			right: NewLiteral(3.0, types.Float64),
			exp:   1.0471966666666666,
		},

		// Decimals
		{
			left:  NewLiteral(decimal.New(1, 0), types.MustCreateDecimalType(10, 0)),
			right: NewLiteral(decimal.New(3, 0), types.MustCreateDecimalType(10, 0)),
			exp:   "0.3333",
		},
		{
			left:  NewLiteral(decimal.New(1000, -3), types.MustCreateDecimalType(10, 3)),
			right: NewLiteral(decimal.New(3, 0), types.MustCreateDecimalType(10, 0)),
			exp:   "0.3333333",
		},
		{
			left:  NewLiteral(decimal.New(1, 0), types.MustCreateDecimalType(10, 0)),
			right: NewLiteral(decimal.New(3000, -3), types.MustCreateDecimalType(10, 3)),
			exp:   "0.3333",
		},
		{
			left:  NewLiteral(decimal.New(314159, -5), types.MustCreateDecimalType(10, 5)),
			right: NewLiteral(decimal.New(3, 0), types.MustCreateDecimalType(10, 0)),
			exp:   "1.047196666",
		},
		{
			left:  NewLiteral(decimal.NewFromFloat(3.14159), types.MustCreateDecimalType(10, 5)),
			right: NewLiteral(3, types.Int64),
			exp:   "1.047196666",
		},

		// Bit
		{
			left:  NewLiteral(0, types.MustCreateBitType(1)),
			right: NewLiteral(1, types.MustCreateBitType(1)),
			exp:   "0.0000",
		},
		{
			left:  NewLiteral(1, types.MustCreateBitType(1)),
			right: NewLiteral(1, types.MustCreateBitType(1)),
			exp:   "1.0000",
		},

		// Year
		{
			left:  NewLiteral(2001, types.YearType_{}),
			right: NewLiteral(2002, types.YearType_{}),
			exp:   "0.9995",
		},

		// Time
		{
			left:  NewLiteral("2001-01-01", types.Date),
			right: NewLiteral("2001-01-01", types.Date),
			exp:   "1.0000",
		},
		{
			left:  NewLiteral("2001-01-01 12:00:00", types.Date),
			right: NewLiteral("2001-01-01 12:00:00", types.Date),
			exp:   "1.0000",
		},
		{
			skip:  true, // need to trim just the date portion
			left:  NewLiteral("2001-01-01 12:00:00.123456", types.Date),
			right: NewLiteral("2001-01-01 12:00:00.123456", types.Date),
			exp:   "1.0000",
		},
		{
			left:  NewLiteral("2001-01-01 12:00:00", types.Datetime),
			right: NewLiteral("2001-01-01 12:00:00", types.Datetime),
			exp:   "1.0000",
		},
		{
			skip:  true, // need to trim just the datetime portion according to precision and use as exponent
			left:  NewLiteral("2001-01-01 12:00:00.123456", types.Datetime),
			right: NewLiteral("2001-01-01 12:00:00.123456", types.Datetime),
			exp:   "1.0000",
		},
		{
			skip:  true, // need to trim just the datetime portion according to precision and use as exponent
			left:  NewLiteral("2001-01-01 12:00:00.123456", types.MustCreateDatetimeType(sqltypes.Datetime, 3)),
			right: NewLiteral("2001-01-01 12:00:00.123456", types.MustCreateDatetimeType(sqltypes.Datetime, 3)),
			exp:   "1.0000000",
		},
		{
			left:  NewLiteral("2001-01-01 12:00:00.123456", types.DatetimeMaxPrecision),
			right: NewLiteral("2001-01-01 12:00:00.123456", types.DatetimeMaxPrecision),
			exp:   "1.0000000000",
		},

		// Text
		{
			left:  NewLiteral("1", types.Text),
			right: NewLiteral("3", types.Text),
			exp:   0.3333333333333333,
		},
		{
			left:  NewLiteral("1.000", types.Text),
			right: NewLiteral("3", types.Text),
			exp:   0.3333333333333333,
		},
		{
			left:  NewLiteral("1", types.Text),
			right: NewLiteral("3.000", types.Text),
			exp:   0.3333333333333333,
		},
		{
			left:  NewLiteral("3.14159", types.Text),
			right: NewLiteral("3", types.Text),
			exp:   1.0471966666666666,
		},
		{
			left:  NewLiteral("1", types.Text),
			right: NewLiteral(decimal.New(3, 0), types.MustCreateDecimalType(10, 0)),
			exp:   0.3333333333333333,
		},
		{
			left:  NewLiteral(decimal.New(1, 0), types.MustCreateDecimalType(10, 0)),
			right: NewLiteral("3", types.Text),
			exp:   0.3333333333333333,
		},
	}

	for _, tt := range testCases {
		name := fmt.Sprintf("%s(%v)/%s(%v)", tt.left.Type(), tt.left, tt.right.Type(), tt.right)
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			if tt.skip {
				t.Skip()
			}
			f := NewDiv(tt.left, tt.right)
			result, err := f.Eval(sql.NewEmptyContext(), nil)
			if tt.err != nil {
				require.Error(err)
				require.True(tt.err.Is(err), err.Error())
				return
			}
			require.NoError(err)
			if dec, ok := result.(decimal.Decimal); ok {
				result = dec.StringFixed(dec.Exponent() * -1)
			}
			assert.Equal(t, tt.exp, result)
		})
	}
}

// TestDivUsesFloatsInternally tests that division expression trees internally use floating point types when operating
// on integers, but when returning the final result from the expression tree, it is returned as a Decimal.
func TestDivUsesFloatsInternally(t *testing.T) {
	t.Skip("TODO: see if we can actually enable this")
	bottomDiv := NewDiv(NewGetField(0, types.Int32, "", false), NewGetField(1, types.Int64, "", false))
	middleDiv := NewDiv(bottomDiv, NewGetField(2, types.Int64, "", false))
	topDiv := NewDiv(middleDiv, NewGetField(3, types.Int64, "", false))

	result, err := topDiv.Eval(sql.NewEmptyContext(), sql.NewRow(250, 2, 5, 2))
	require.NoError(t, err)
	dec, isDecimal := result.(decimal.Decimal)
	require.True(t, isDecimal)
	require.Equal(t, "12.5", dec.String())

	// Internal nodes should use floats for division with integers (for performance reasons), but the top node
	// should return a Decimal (to match MySQL's behavior).
	require.Equal(t, types.Float64, bottomDiv.Type())
	require.Equal(t, types.Float64, middleDiv.Type())
	require.True(t, types.IsDecimal(topDiv.Type()))
}

func TestIntDiv(t *testing.T) {
	var testCases = []struct {
		name                string
		left, right         interface{}
		leftType, rightType sql.Type
		expected            int64
		null                bool
	}{
		{"1 div 1", 1, 1, types.Int64, types.Int64, 1, false},
		{"8 div 3", 8, 3, types.Int64, types.Int64, 2, false},
		{"1 div 3", 1, 3, types.Int64, types.Int64, 0, false},
		{"0 div -1024", 0, -1024, types.Int64, types.Int64, 0, false},
		{"1 div 0", 1, 0, types.Int64, types.Int64, 0, true},
		{"0 div 0", 1, 0, types.Int64, types.Int64, 0, true},
		{"10.24 div 0.6", 10.24, 0.6, types.Float64, types.Float64, 17, false},
		{"-10.24 div 0.6", -10.24, 0.6, types.Float64, types.Float64, -17, false},
		{"-10.24 div -0.6", -10.24, -0.6, types.Float64, types.Float64, 17, false},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			result, err := NewIntDiv(
				NewLiteral(tt.left, tt.leftType),
				NewLiteral(tt.right, tt.rightType),
			).Eval(sql.NewEmptyContext(), sql.NewRow())
			require.NoError(err)
			if tt.null {
				assert.Equal(t, nil, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}


func BenchmarkDiv(b *testing.B) {
	require := require.New(b)
	ctx := sql.NewEmptyContext()
	l := NewLiteral(1, types.Int64)
	r := NewLiteral(3, types.Int64)
	d := NewDiv(l, r)
	for i := 0; i < b.N; i++ {
		res, err := d.Eval(ctx, nil)
		require.NoError(err)
		if dec, ok := res.(decimal.Decimal); ok {
			res = dec.StringFixed(dec.Exponent() * -1)
		}
		require.Equal(res, "0.3333")
	}
}

func BenchmarkDivFloat(b *testing.B) {
	require := require.New(b)
	ctx := sql.NewEmptyContext()
	l := NewLiteral(1.0, types.Float64)
	r := NewLiteral(3.0, types.Float64)
	d := NewDiv(l, r)
	for i := 0; i < b.N; i++ {
		res, err := d.Eval(ctx, nil)
		require.NoError(err)
		if dec, ok := res.(decimal.Decimal); ok {
			res = dec.StringFixed(dec.Exponent() * -1)
		}
		require.Equal(res, 1.0/3.0)
	}
}