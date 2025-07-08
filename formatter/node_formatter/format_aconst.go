package nodeformatter

import (
	"context"
	"fmt"

	pg_query "github.com/pganalyze/pg_query_go/v6"
)

func FormatAConst(ctx context.Context, ac *pg_query.Node_AConst) (string, error) {
	switch val := ac.AConst.Val.(type) {
	case *pg_query.A_Const_Ival:
		return fmt.Sprint(val.Ival.Ival), nil
	case *pg_query.A_Const_Sval:
		return "'" + val.Sval.Sval + "'", nil
	case *pg_query.A_Const_Boolval:
		return fmt.Sprint(val.Boolval.Boolval), nil
	case *pg_query.A_Const_Fval:
		return val.Fval.Fval, nil
	case *pg_query.A_Const_Bsval:
		return "'" + string(val.Bsval.Bsval) + "'", nil
	case nil:
		return "NULL", nil
	}
	return "", fmt.Errorf("FormatAConst not implemented for type %T", ac.AConst.Val)
}
