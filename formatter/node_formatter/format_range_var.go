package nodeformatter

import (
	"context"
	"fmt"

	pg_query "github.com/pganalyze/pg_query_go/v4"
)

func FormatFromAndTable(ctx context.Context, n *pg_query.Node_RangeVar) (string, error) {
	table := fmt.Sprintf("\nFROM %s", n.RangeVar.Relname)
	if n.RangeVar.Alias != nil {
		return fmt.Sprintf("%s %s", table, n.RangeVar.Alias.Aliasname), nil
	}
	return table, nil
}
