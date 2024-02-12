package nodeformatter

import (
	"context"
	"errors"
	"fmt"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

func FormatAConst(ctx context.Context, ac *pg_query.Node_AConst) (string, error) {
	switch val := ac.AConst.Val.(type) {
	case *pg_query.A_Const_Ival:
		return fmt.Sprint(val.Ival.Ival), nil
	case *pg_query.A_Const_Sval:
		return "'" + val.Sval.Sval + "'", nil
	case *pg_query.A_Const_Boolval:
		return fmt.Sprint(val.Boolval.Boolval), nil
	}
	return "", errors.New("FormatAConst not implemented")
}
