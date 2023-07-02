package nodeformatter

import (
	"context"

	pg_query "github.com/pganalyze/pg_query_go/v4"
)

func FormatRelation(ctx context.Context, relation *pg_query.RangeVar) (string, error) {
	if relation == nil {
		return "", nil
	}

	return " " + relation.Relname, nil
}
