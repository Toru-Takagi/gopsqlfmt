package enumconv

import (
	"errors"

	pg_query "github.com/pganalyze/pg_query_go/v4"
)

func BoolExprTypeToString(bet pg_query.BoolExprType) (string, error) {
	switch bet {
	case pg_query.BoolExprType_AND_EXPR:
		return "AND", nil
	case pg_query.BoolExprType_OR_EXPR:
		return "OR", nil
	case pg_query.BoolExprType_NOT_EXPR:
		return "NOT", nil
	}
	return "", errors.New("BoolExprTypeToString: unknown BoolExprType")
}
