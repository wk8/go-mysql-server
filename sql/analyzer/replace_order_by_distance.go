package analyzer

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
)

// replaceIdxSort applies an IndexAccess when there is an `OrderBy` over a prefix of any columns with Indexes
func replaceIdxOrderByDistance(ctx *sql.Context, a *Analyzer, n sql.Node, scope *plan.Scope, sel RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return replaceIdxOrderByDistanceHelper(ctx, scope, n, nil)
}

func replaceIdxOrderByDistanceHelper(ctx *sql.Context, scope *plan.Scope, node sql.Node, sortNode plan.Sortable) (sql.Node, transform.TreeIdentity, error) {
	switch n := node.(type) {
	case plan.Sortable:
		sortNode = n // lowest parent sort node
	case *plan.ResolvedTable:
		if sortNode == nil {
			return n, transform.SameTree, nil
		}

		table := n.UnderlyingTable()
		idxTbl, ok := table.(sql.IndexAddressableTable)
		if !ok {
			return n, transform.SameTree, nil
		}
		if indexSearchable, ok := table.(sql.IndexSearchableTable); ok && indexSearchable.SkipIndexCosting() {
			return n, transform.SameTree, nil
		}

		tableAliases, err := getTableAliases(sortNode, scope)
		if err != nil {
			return n, transform.SameTree, nil
		}

		var idx sql.Index
		idxs, err := idxTbl.GetIndexes(ctx)
		if err != nil {
			return nil, transform.SameTree, err
		}
		sfExprs := normalizeExpressions(tableAliases, sortNode.GetSortFields().ToExpressions()...)
		sfAliases := aliasedExpressionsInNode(sortNode)

		// TODO: Instead of checking both sides of the expression,
		// use a previous pass to normalize distance functions so
		// that the literal is always on the same side.
		if len(sfExprs) != 1 {
			return n, transform.SameTree, nil
		}
		distance, isDistance := sfExprs[0].(*expression.Distance)
		if !isDistance {
			return n, transform.SameTree, nil
		}
		var column sql.Expression
		_, leftIsLiteral := distance.LeftChild.(*expression.Literal)
		if leftIsLiteral {
			column = distance.RightChild
		} else {
			_, rightIsLiteral := distance.RightChild.(*expression.Literal)
			if rightIsLiteral {
				column = distance.LeftChild
			} else {
				return n, transform.SameTree, nil
			}
		}

		for _, idxCandidate := range idxs {
			if idxCandidate.IsSpatial() {
				continue
			}
			if !idxCandidate.CanSupportOrderBy(distance) {
				continue
			}
			if isSortFieldsValidPrefix([]sql.Expression{column}, sfAliases, idxCandidate.Expressions()) {
				idx = idxCandidate
				break
			}
		}
		if idx == nil {
			return n, transform.SameTree, nil
		}

		var limit sql.Expression
		if topn, ok := sortNode.(*plan.TopN); ok {
			limit = topn.Limit
		}

		lookup := sql.IndexLookup{
			Index:  idx,
			Ranges: sql.MySQLRangeCollection{},
			VectorOrderAndLimit: sql.OrderAndLimit{
				OrderBy: distance,
				Limit:   limit,
			},
		}
		nn, err := plan.NewStaticIndexedAccessForTableNode(n, lookup)
		if err != nil {
			return nil, transform.SameTree, err
		}
		return nn, transform.NewTree, err
	}

	allSame := transform.SameTree
	newChildren := make([]sql.Node, len(node.Children()))
	for i, child := range node.Children() {
		var err error
		same := transform.SameTree
		switch c := child.(type) {
		case *plan.Project, *plan.TableAlias, *plan.ResolvedTable, *plan.Filter, *plan.Limit, *plan.Offset, *plan.Sort, *plan.IndexedTableAccess:
			newChildren[i], same, err = replaceIdxOrderByDistanceHelper(ctx, scope, child, sortNode)
		default:
			newChildren[i] = c
		}
		if err != nil {
			return nil, transform.SameTree, err
		}
		allSame = allSame && same
	}

	if allSame {
		return node, transform.SameTree, nil
	}

	// if sort node was replaced with indexed access, drop sort node
	if node == sortNode {
		return newChildren[0], transform.NewTree, nil
	}

	newNode, err := node.WithChildren(newChildren...)
	if err != nil {
		return nil, transform.SameTree, err
	}
	return newNode, transform.NewTree, nil
}
