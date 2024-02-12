package nodeformatter

import (
	"context"
	"errors"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

func FormatSortByDir(ctx context.Context, sortBy *pg_query.Node_SortBy) (string, error) {
	switch sortBy.SortBy.SortbyDir {
	case pg_query.SortByDir_SORTBY_ASC:
		return " ASC", nil
	case pg_query.SortByDir_SORTBY_DESC:
		return " DESC", nil
	case pg_query.SortByDir_SORTBY_DEFAULT:
		return "", nil
	}
	return "", errors.New("FormatSortByDir not implemented")
}
