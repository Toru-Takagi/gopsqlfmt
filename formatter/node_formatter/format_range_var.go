package nodeformatter

import (
	"context"
	"fmt"

	pg_query "github.com/pganalyze/pg_query_go/v4"
)

func FormatFromAndTable(ctx context.Context, n *pg_query.Node_RangeVar) (string, error) {
	tableName, err := FormatTableName(ctx, n)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("\nFROM %s", tableName), nil
}

func FormatTableName(ctx context.Context, n *pg_query.Node_RangeVar) (string, error) {
	table := n.RangeVar.Relname
	if n.RangeVar.Alias != nil {
		return fmt.Sprintf("%s %s", table, n.RangeVar.Alias.Aliasname), nil
	}
	return table, nil
}
