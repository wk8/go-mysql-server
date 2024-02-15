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
	"math"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/shopspring/decimal"
	"gopkg.in/src-d/go-errors.v1"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
)

var ErrIntDivDataOutOfRange = errors.NewKind("BIGINT value is out of range (%s DIV %s)")

// '4 scales' are added to scale of the number on the left side of division operator at every division operation.
// The default value is 4, and it can be set using sysvar https://dev.mysql.com/doc/refman/8.0/en/server-system-variables.html#sysvar_div_precision_increment
const divPrecInc = 4

// '9 scales' are added for every non-integer divider(right side).
const divIntPrecInc = 9

const ERDivisionByZero = 1365

var _ ArithmeticOp = (*Div)(nil)
var _ sql.CollationCoercible = (*Div)(nil)

// Div expression represents "/" arithmetic operation
type Div struct {
	BinaryExpressionStub
	ops int32
	// divScale is number of continuous division operations; this value will be available of all layers
	divScale int32 // TODO: calling this divScale is confusing
	// leftmostScale is a length of scale of the leftmost value in continuous division operation
	// It is accessed concurrently read in the .Type() and written in the .Eval() methods.
	leftmostScale               atomic.Int32
	curIntermediatePrecisionInc int
}

// NewDiv creates a new Div / sql.Expression.
func NewDiv(left, right sql.Expression) *Div {
	a := &Div{
		BinaryExpressionStub:        BinaryExpressionStub{LeftChild: left, RightChild: right},
		curIntermediatePrecisionInc: 0,
	}
	a.leftmostScale.Store(0)
	divs := countDivs(a)
	setDivs(a, divs)
	ops := countArithmeticOps(a)
	setArithmeticOps(a, ops)
	return a
}

func (d *Div) Operator() string {
	return sqlparser.DivStr
}

func (d *Div) SetOpCount(i int32) {
	d.ops = i
}

func (d *Div) String() string {
	return fmt.Sprintf("(%s / %s)", d.LeftChild, d.RightChild)
}

func (d *Div) DebugString() string {
	return fmt.Sprintf("(%s / %s)", sql.DebugString(d.LeftChild), sql.DebugString(d.RightChild))
}

// IsNullable implements the sql.Expression interface.
func (d *Div) IsNullable() bool {
	return d.BinaryExpressionStub.IsNullable()
}

// Type returns the result type for this division expression. For nested division expressions, we prefer sending
// the result back as a float when possible, since division with floats is more efficient than division with Decimals.
// However, if this is the outermost division expression in an expression tree, we must return the result as a
// Decimal type in order to match MySQL's results exactly.
func (d *Div) Type() sql.Type {
	return d.determineResultType(isOutermostDiv(d, 0, d.divScale))
}

// internalType returns the internal result type for this division expression. For performance reasons, we prefer
// to use floats internally in division operations wherever possible, since division operations on floats can be
// orders of magnitude faster than division operations on Decimal types.
func (d *Div) internalType() sql.Type {
	return d.determineResultType(false)
}

// CollationCoercibility implements the interface sql.CollationCoercible.
func (*Div) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// WithChildren implements the Expression interface.
func (d *Div) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 2 {
		return nil, sql.ErrInvalidChildrenNumber.New(d, len(children), 2)
	}
	return NewDiv(children[0], children[1]), nil
}

// Eval implements the Expression interface.
func (d *Div) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	lval, rval, err := d.evalLeftRight(ctx, row)
	if err != nil {
		return nil, err
	}

	if lval == nil || rval == nil {
		return nil, nil
	}

	lval, rval = d.convertLeftRight(ctx, lval, rval)

	result, err := d.div(ctx, lval, rval)
	if err != nil {
		return nil, err
	}

	// Decimals must be rounded
	if res, ok := result.(decimal.Decimal); ok {
		if isOutermostArithmeticOp(d, d.ops) {
			finalScale, _ := getFinalScale(ctx, row, d, 0)
			return res.Round(finalScale), nil
		}
	}

	return result, nil
}

func (d *Div) evalLeftRight(ctx *sql.Context, row sql.Row) (interface{}, interface{}, error) {
	var lval, rval interface{}
	var err error

	// division used with Interval error is caught at parsing the query
	lval, err = d.LeftChild.Eval(ctx, row)
	if err != nil {
		return nil, nil, err
	}

	// this operation is only done on the left value as the scale/fraction part of the leftmost value
	// is used to calculate the scale of the final result. If the value is GetField of decimal type column
	// the decimal value evaluated does not always match the scale of column type definition
	if dt, ok := d.LeftChild.Type().(sql.DecimalType); ok {
		if dVal, ok := lval.(decimal.Decimal); ok {
			ts := int32(dt.Scale())
			if ts > dVal.Exponent()*-1 {
				lval, err = decimal.NewFromString(dVal.StringFixed(ts))
				if err != nil {
					return nil, nil, err
				}
			}
		}
	}

	rval, err = d.RightChild.Eval(ctx, row)
	if err != nil {
		return nil, nil, err
	}

	return lval, rval, nil
}

// convertLeftRight returns the most appropriate type for left and right evaluated values,
// which may or may not be converted from its original type.
// It checks for float type column reference, then the both values converted to the same float type.
// Integer column references are treated as floats internally for performance reason, but the final result
// from the expression tree is converted to a Decimal in order to match MySQL's behavior.
// The decimal types of left and right value does NOT need to be the same. Both the types
// should be preserved.
func (d *Div) convertLeftRight(ctx *sql.Context, left interface{}, right interface{}) (interface{}, interface{}) {
	typ := d.internalType()
	lIsTimeType := types.IsTime(d.LeftChild.Type())
	rIsTimeType := types.IsTime(d.RightChild.Type())

	if types.IsFloat(typ) {
		left = convertValueToType(ctx, typ, left, lIsTimeType)
	} else {
		left = convertToDecimalValue(left, lIsTimeType)
	}

	if types.IsFloat(typ) {
		right = convertValueToType(ctx, typ, right, rIsTimeType)
	} else {
		right = convertToDecimalValue(right, rIsTimeType)
	}

	return left, right
}

func (d *Div) div(ctx *sql.Context, lval, rval interface{}) (interface{}, error) {
	switch l := lval.(type) {
	case float32:
		switch r := rval.(type) {
		case float32:
			if r == 0 {
				arithmeticWarning(ctx, ERDivisionByZero, "Division by 0")
				return nil, nil
			}
			return l / r, nil
		}
	case float64:
		switch r := rval.(type) {
		case float64:
			if r == 0 {
				arithmeticWarning(ctx, ERDivisionByZero, "Division by 0")
				return nil, nil
			}
			return l / r, nil
		}
	case decimal.Decimal:
		switch r := rval.(type) {
		case decimal.Decimal:
			if r.Equal(decimal.NewFromInt(0)) {
				arithmeticWarning(ctx, ERDivisionByZero, "Division by 0")
				return nil, nil
			}

			lScale, rScale := -1*l.Exponent(), -1*r.Exponent()
			inc := int32(math.Ceil(float64(lScale+rScale+divPrecInc) / divIntPrecInc))
			if lScale != 0 && rScale != 0 {
				lInc := int32(math.Ceil(float64(lScale) / divIntPrecInc))
				rInc := int32(math.Ceil(float64(rScale) / divIntPrecInc))
				inc2 := lInc + rInc
				if inc2 > inc {
					inc = inc2
				}
			}
			scale := inc * divIntPrecInc
			l = l.Truncate(scale)
			r = r.Truncate(scale)

			// give it buffer of 2 additional scale to avoid the result to be rounded
			res := l.DivRound(r, scale+2)
			res = res.Truncate(scale)
			return res, nil
		}
	}

	return nil, errUnableToCast.New(lval, rval)
}

// determineResultType looks at the expressions in the expression tree with this division operation and determines
// the result type of this division expression. This involves looking at the types of the expressions in the tree,
// and looking for float types or Decimal types. If |outermostResult| is false, then we prefer to treat ints as floats
// (instead of Decimals) for performance reasons, but when |outermostResult| is true, we must treat ints as Decimals
// in order to match MySQL's behavior.
func (d *Div) determineResultType(outermostResult bool) sql.Type {
	//TODO: what if both BindVars? should be constant folded
	rTyp := d.RightChild.Type()
	if types.IsDeferredType(rTyp) {
		return rTyp
	}
	lTyp := d.LeftChild.Type()
	if types.IsDeferredType(lTyp) {
		return lTyp
	}

	if types.IsText(lTyp) || types.IsText(rTyp) {
		return types.Float64
	}

	if types.IsJSON(lTyp) || types.IsJSON(rTyp) {
		return types.Float64
	}

	if types.IsFloat(lTyp) || types.IsFloat(rTyp) {
		return types.Float64
	}

	// TODO: see if we can actually do this
	//if !outermostResult {
	//	return types.Float64
	//}

	// Decimal only results from here on

	if types.IsDatetimeType(lTyp) {
		if dtType, ok := lTyp.(sql.DatetimeType); ok {
			scale := uint8(dtType.Precision() + divPrecInc)
			if scale > types.DecimalTypeMaxScale {
				scale = types.DecimalTypeMaxScale
			}
			// TODO: determine actual precision
			return types.MustCreateDecimalType(types.DecimalTypeMaxPrecision, scale)
		}
	}

	if types.IsDecimal(lTyp) {
		prec, scale := lTyp.(types.DecimalType_).Precision(), lTyp.(types.DecimalType_).Scale()
		scale = scale + divPrecInc
		if d.ops == -1 {
			scale = (scale/9 + 1) * 9
			prec = prec + scale
		} else {
			prec = prec + divPrecInc
		}

		if prec > types.DecimalTypeMaxPrecision {
			prec = types.DecimalTypeMaxPrecision
		}
		if scale > types.DecimalTypeMaxScale {
			scale = types.DecimalTypeMaxScale
		}
		return types.MustCreateDecimalType(prec, scale)
	}

	// All other types are treated as if they were integers
	if d.ops == -1 {
		return types.MustCreateDecimalType(types.DecimalTypeMaxPrecision, 9)
	}
	return types.MustCreateDecimalType(types.DecimalTypeMaxPrecision, divPrecInc)
}

// getFloatOrMaxDecimalType returns either Float64 or Decimal type with max precision and scale
// depending on column reference, expression types and evaluated value types. Otherwise, the return
// type is always max decimal type. |treatIntsAsFloats| is used for division operation optimization.
func getFloatOrMaxDecimalType(e sql.Expression, treatIntsAsFloats bool) sql.Type {
	var resType sql.Type
	var maxWhole, maxFrac uint8
	sql.Inspect(e, func(expr sql.Expression) bool {
		switch c := expr.(type) {
		case *GetField:
			ct := c.Type()
			if treatIntsAsFloats && types.IsInteger(ct) {
				resType = types.Float64
				return false
			}
			// If there is float type column reference, the result type is always float.
			if types.IsFloat(ct) {
				resType = types.Float64
				return false
			}
			if types.IsDecimal(ct) {
				dt := ct.(sql.DecimalType)
				p, s := dt.Precision(), dt.Scale()
				if whole := p - s; whole > maxWhole {
					maxWhole = whole
				}
				if s > maxFrac {
					maxFrac = s
				}
			}
		case *Convert:
			if c.cachedDecimalType != nil {
				p, s := GetPrecisionAndScale(c.cachedDecimalType)
				if whole := p - s; whole > maxWhole {
					maxWhole = whole
				}
				if s > maxFrac {
					maxFrac = s
				}
			}
		case *Literal:
			if types.IsNumber(c.Type()) {
				l, err := c.Eval(nil, nil)
				if err == nil {
					p, s := GetPrecisionAndScale(l)
					if whole := p - s; whole > maxWhole {
						maxWhole = whole
					}
					if s > maxFrac {
						maxFrac = s
					}
				}
			}
		case sql.FunctionExpression:
			// Mod.Type() calls this, so ignore it for infinite loop
			if c.FunctionName() != "mod" {
				resType = c.Type()
			}
		}
		return true
	})
	if resType == types.Float64 {
		return resType
	}

	// defType is defined by evaluating all number literals available and defined column type.
	defType, derr := types.CreateDecimalType(maxWhole+maxFrac, maxFrac)
	if derr != nil {
		return types.MustCreateDecimalType(65, 10)
	}

	return defType
}

// convertToDecimalValue returns value converted to decimaltype.
// If the value is invalid, it returns decimal 0. This function
// is used for 'div' or 'mod' arithmetic operation, which requires
// the result value to have precise precision and scale.
func convertToDecimalValue(val interface{}, isTimeType bool) interface{} {
	if isTimeType {
		val = convertTimeTypeToString(val)
	}
	switch v := val.(type) {
	case bool:
		val = 0
		if v {
			val = 1
		}
	default:
	}

	if _, ok := val.(decimal.Decimal); !ok {
		p, s := GetPrecisionAndScale(val)
		if p > types.DecimalTypeMaxPrecision {
			p = types.DecimalTypeMaxPrecision
		}
		if s > types.DecimalTypeMaxScale {
			s = types.DecimalTypeMaxScale
		}
		dtyp, err := types.CreateDecimalType(p, s)
		if err != nil {
			val = decimal.Zero
		}
		val, _, err = dtyp.Convert(val)
		if err != nil {
			val = decimal.Zero
		}
	}

	return val
}

// countDivs returns the number of division operators in order on the left child node of the current node.
// This lets us count how many division operator used one after the other. E.g. 24/3/2/1 will have this structure:
//
//		     'div'
//		     /   \
//		   'div'  1
//		   /   \
//		 'div'  2
//		 /   \
//	    24    3
func countDivs(e sql.Expression) int32 {
	if e == nil {
		return 0
	}

	if a, ok := e.(*Div); ok {
		return countDivs(a.LeftChild) + 1
	}

	if a, ok := e.(ArithmeticOp); ok {
		return countDivs(a.Left())
	}

	return 0
}

// setDivs will set each node's DivScale to the number counted by countDivs. This allows us to
// keep track of whether the current Div expression is the last Div operation, so the result is
// rounded appropriately.
func setDivs(e sql.Expression, dScale int32) {
	if e == nil {
		return
	}

	if a, isArithmeticOp := e.(ArithmeticOp); isArithmeticOp {
		if d, ok := a.(*Div); ok {
			d.divScale = dScale
		}
		setDivs(a.Left(), dScale)
		setDivs(a.Right(), dScale)
	}

	if tup, ok := e.(Tuple); ok {
		for _, expr := range tup {
			setDivs(expr, dScale)
		}
	}

	return
}

// getScaleOfLeftmostValue find the leftmost/first value of all continuous divisions.
// E.g. 24/50/3.2/2/1 will return 2 for len('50') of number '24.50'.
func getScaleOfLeftmostValue(ctx *sql.Context, row sql.Row, e sql.Expression, d, dScale int32) int32 {
	if e == nil {
		return 0
	}

	if a, ok := e.(*Div); ok {
		d = d + 1
		if d == dScale {
			lval, err := a.LeftChild.Eval(ctx, row)
			if err != nil {
				return 0
			}
			_, s := GetPrecisionAndScale(lval)
			// the leftmost value can be row value of decimal type column
			// the evaluated value does not always match the scale of column type definition
			typ := a.LeftChild.Type()
			if dt, dok := typ.(sql.DecimalType); dok {
				ts := dt.Scale()
				if ts > s {
					s = ts
				}
			}
			return int32(s)
		} else {
			return getScaleOfLeftmostValue(ctx, row, a.LeftChild, d, dScale)
		}
	}

	return 0
}

// isOutermostDiv returns whether the expression we're currently evaluating is
// the last division operation of all continuous divisions.
// E.g. the top 'div' (divided by 1) is the outermost/last division that is calculated:
//
//		     'div'
//		     /   \
//		   'div'  1
//		   /   \
//		 'div'  2
//		 /   \
//	    24    3
func isOutermostDiv(e sql.Expression, d, dScale int32) bool {
	if e == nil {
		return false
	}

	if a, ok := e.(*Div); ok {
		d = d + 1
		if d == dScale {
			return true
		}
		return isOutermostDiv(a.LeftChild, d, dScale)
	}

	if a, ok := e.(ArithmeticOp); ok {
		return isOutermostDiv(a.Left(), d, dScale)
	}

	return false
}

// getFinalScale returns the final scale of the result value.
// it traverses both the left and right nodes looking for Div, Arithmetic, and Literal nodes
func getFinalScale(ctx *sql.Context, row sql.Row, e sql.Expression, d int32) (int32, bool) {
	if e == nil {
		return 0, false
	}

	if div, isDiv := e.(*Div); isDiv {
		// TODO: there's gotta be a better way of determining if this is the leftmost div...
		finalScale := int32(divPrecInc)
		d = d + 1
		if d == div.divScale {
			// TODO: redundant call to Eval for LeftChild
			lval, err := div.LeftChild.Eval(ctx, row)
			if err != nil {
				return 0, false
			}
			_, s := GetPrecisionAndScale(lval)
			typ := div.LeftChild.Type()
			if dt, dok := typ.(sql.DecimalType); dok {
				ts := dt.Scale()
				if ts > s {
					s = ts
				}
			}
			finalScale += int32(s)
		} else {
			// We only care about left scale for divs
			leftScale, _ := getFinalScale(ctx, row, div.LeftChild, d)
			finalScale += leftScale
		}

		if finalScale > types.DecimalTypeMaxScale {
			finalScale = types.DecimalTypeMaxScale
		}
		return finalScale, true
	}

	if a, isArith := e.(*Arithmetic); isArith {
		leftScale, leftHasDiv := getFinalScale(ctx, row, a.Left(), d)
		rightScale, rightHasDiv := getFinalScale(ctx, row, a.Right(), d)
		var finalScale int32
		switch a.Operator() {
		case sqlparser.PlusStr, sqlparser.MinusStr:
			if leftScale > rightScale {
				finalScale = leftScale
			} else {
				finalScale = rightScale
			}
		case sqlparser.MultStr:
			finalScale = leftScale + rightScale
		}
		if finalScale > types.DecimalTypeMaxScale {
			finalScale = types.DecimalTypeMaxScale
		}
		return finalScale, leftHasDiv || rightHasDiv
	}

	// TODO: this is just a guess of what mod should do with scale; test this
	if m, isMod := e.(*Mod); isMod {
		leftScale, leftHasDiv := getFinalScale(ctx, row, m.LeftChild, d)
		rightScale, rightHasDiv := getFinalScale(ctx, row, m.RightChild, d)
		finalScale := leftScale
		if rightScale > finalScale {
			finalScale = rightScale
		}
		if finalScale > types.DecimalTypeMaxScale {
			finalScale = types.DecimalTypeMaxScale
		}
		return finalScale, leftHasDiv || rightHasDiv
	}

	// TODO: likely need a case for IntDiv

	s := uint8(0)
	if lit, isLit := e.(*Literal); isLit {
		_, s = GetPrecisionAndScale(lit.value)
	}
	typ := e.Type()
	if dt, dok := typ.(sql.DecimalType); dok {
		ts := dt.Scale()
		if ts > s {
			s = ts
		}
	}

	return int32(s), false
}

// GetDecimalPrecisionAndScale returns precision and scale for given string formatted float/double number.
func GetDecimalPrecisionAndScale(val string) (uint8, uint8) {
	scale := 0
	precScale := strings.Split(strings.TrimPrefix(val, "-"), ".")
	if len(precScale) != 1 {
		scale = len(precScale[1])
	}
	precision := len((precScale)[0]) + scale
	return uint8(precision), uint8(scale)
}

// GetPrecisionAndScale converts the value to string format and parses it to get the precision and scale.
func GetPrecisionAndScale(val interface{}) (uint8, uint8) {
	var str string
	switch v := val.(type) {
	case time.Time:
		str = fmt.Sprintf("%v", v.In(time.UTC).Format("2006-01-02 15:04:05"))
	case decimal.Decimal:
		str = v.StringFixed(v.Exponent() * -1)
	case float32:
		d := decimal.NewFromFloat32(v)
		str = d.StringFixed(d.Exponent() * -1)
	case float64:
		d := decimal.NewFromFloat(v)
		str = d.StringFixed(d.Exponent() * -1)
	default:
		str = fmt.Sprintf("%v", v)
	}
	return GetDecimalPrecisionAndScale(str)
}

var _ ArithmeticOp = (*IntDiv)(nil)
var _ sql.CollationCoercible = (*IntDiv)(nil)

// IntDiv expression represents integer "div" arithmetic operation
type IntDiv struct {
	BinaryExpressionStub
	ops int32
}

// NewIntDiv creates a new IntDiv 'div' sql.Expression.
func NewIntDiv(left, right sql.Expression) *IntDiv {
	a := &IntDiv{BinaryExpressionStub{LeftChild: left, RightChild: right}, 0}
	ops := countArithmeticOps(a)
	setArithmeticOps(a, ops)
	return a
}

func (i *IntDiv) Operator() string {
	return sqlparser.IntDivStr
}

func (i *IntDiv) SetOpCount(i2 int32) {
	i.ops = i2
}

func (i *IntDiv) String() string {
	return fmt.Sprintf("(%s div %s)", i.LeftChild, i.RightChild)
}

func (i *IntDiv) DebugString() string {
	return fmt.Sprintf("(%s div %s)", sql.DebugString(i.LeftChild), sql.DebugString(i.RightChild))
}

// IsNullable implements the sql.Expression interface.
func (i *IntDiv) IsNullable() bool {
	return i.BinaryExpressionStub.IsNullable()
}

// Type returns the greatest type for given operation.
func (i *IntDiv) Type() sql.Type {
	lTyp := i.LeftChild.Type()
	rTyp := i.RightChild.Type()

	if types.IsUnsigned(lTyp) || types.IsUnsigned(rTyp) {
		return types.Uint64
	}

	return types.Int64
}

// CollationCoercibility implements the interface sql.CollationCoercible.
func (*IntDiv) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// WithChildren implements the Expression interface.
func (i *IntDiv) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 2 {
		return nil, sql.ErrInvalidChildrenNumber.New(i, len(children), 2)
	}
	return NewIntDiv(children[0], children[1]), nil
}

// Eval implements the Expression interface.
func (i *IntDiv) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	lval, rval, err := i.evalLeftRight(ctx, row)
	if err != nil {
		return nil, err
	}

	if lval == nil || rval == nil {
		return nil, nil
	}

	lval, rval = i.convertLeftRight(ctx, lval, rval)

	return intDiv(ctx, lval, rval)
}

func (i *IntDiv) evalLeftRight(ctx *sql.Context, row sql.Row) (interface{}, interface{}, error) {
	var lval, rval interface{}
	var err error

	// int division used with Interval error is caught at parsing the query
	lval, err = i.LeftChild.Eval(ctx, row)
	if err != nil {
		return nil, nil, err
	}

	rval, err = i.RightChild.Eval(ctx, row)
	if err != nil {
		return nil, nil, err
	}

	return lval, rval, nil
}

// convertLeftRight return most appropriate value for left and right from evaluated value,
// which can might or might not be converted from its original value.
// It checks for float type column reference, then the both values converted to the same float types.
// If there is no float type column reference, both values should be handled as decimal type
// The decimal types of left and right value does NOT need to be the same. Both the types
// should be preserved.
func (i *IntDiv) convertLeftRight(ctx *sql.Context, left interface{}, right interface{}) (interface{}, interface{}) {
	var typ sql.Type
	lTyp, rTyp := i.LeftChild.Type(), i.RightChild.Type()
	lIsTimeType := types.IsTime(lTyp)
	rIsTimeType := types.IsTime(rTyp)

	if types.IsText(lTyp) || types.IsText(rTyp) {
		typ = types.Float64
	} else if types.IsUnsigned(lTyp) && types.IsUnsigned(rTyp) {
		typ = types.Uint64
	} else if (lIsTimeType && rIsTimeType) || (types.IsSigned(lTyp) && types.IsSigned(rTyp)) {
		typ = types.Int64
	} else {
		typ = types.MustCreateDecimalType(types.DecimalTypeMaxPrecision, 0)
	}

	if types.IsInteger(typ) || types.IsFloat(typ) {
		left = convertValueToType(ctx, typ, left, lIsTimeType)
		right = convertValueToType(ctx, typ, right, rIsTimeType)
	} else {
		left = convertToDecimalValue(left, lIsTimeType)
		right = convertToDecimalValue(right, rIsTimeType)
	}

	return left, right
}

func intDiv(ctx *sql.Context, lval, rval interface{}) (interface{}, error) {
	switch l := lval.(type) {
	case uint64:
		switch r := rval.(type) {
		case uint64:
			if r == 0 {
				arithmeticWarning(ctx, ERDivisionByZero, "Division by 0")
				return nil, nil
			}
			return l / r, nil
		}
	case int64:
		switch r := rval.(type) {
		case int64:
			if r == 0 {
				arithmeticWarning(ctx, ERDivisionByZero, "Division by 0")
				return nil, nil
			}
			return l / r, nil
		}
	case float64:
		switch r := rval.(type) {
		case float64:
			if r == 0 {
				arithmeticWarning(ctx, ERDivisionByZero, "Division by 0")
				return nil, nil
			}
			res := l / r
			return int64(math.Floor(res)), nil
		}
	case decimal.Decimal:
		switch r := rval.(type) {
		case decimal.Decimal:
			if r.Equal(decimal.NewFromInt(0)) {
				arithmeticWarning(ctx, ERDivisionByZero, "Division by 0")
				return nil, nil
			}

			// intDiv operation gets the integer part of the divided value without rounding the result with 0 precision
			// We get division result with non-zero precision and then truncate it to get integer part without it being rounded
			divRes := l.DivRound(r, 2).Truncate(0)

			// cannot use IntPart() function of decimal.Decimal package as it returns 0 as undefined value for out of range value
			// it causes valid result value of 0 to be the same as invalid out of range value of 0. The fraction part
			// should not be rounded, so truncate the result wih 0 precision.
			intPart, err := strconv.ParseInt(divRes.String(), 10, 64)
			if err != nil {
				return nil, ErrIntDivDataOutOfRange.New(l.StringFixed(l.Exponent()), r.StringFixed(r.Exponent()))
			}

			return intPart, nil
		}
	}

	return nil, errUnableToCast.New(lval, rval)
}
