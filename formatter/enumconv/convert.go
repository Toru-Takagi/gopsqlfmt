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

func JoinTypeToString(jt pg_query.JoinType) (string, error) {
	switch jt {
	case pg_query.JoinType_JOIN_INNER:
		return "INNER JOIN", nil
	case pg_query.JoinType_JOIN_LEFT:
		return "LEFT JOIN", nil
	case pg_query.JoinType_JOIN_FULL:
		return "FULL JOIN", nil
	case pg_query.JoinType_JOIN_RIGHT:
		return "RIGHT JOIN", nil
	case pg_query.JoinType_JOIN_SEMI:
		return "SEMI JOIN", nil
	case pg_query.JoinType_JOIN_ANTI:
		return "ANTI JOIN", nil
	}
	return "", errors.New("JoinTypeToString: unknown JoinType")
}
