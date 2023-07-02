package nodeformatter

import (
	"context"

	pg_query "github.com/pganalyze/pg_query_go/v4"
)

func FormatRelation(ctx context.Context, relation interface{}) (string, error) {
	if relation == nil {
		return "", nil
	}

	if rangeVar, ok := relation.(*pg_query.RangeVar); ok {
		return " " + rangeVar.Relname, nil
	}

	return "", nil
}
