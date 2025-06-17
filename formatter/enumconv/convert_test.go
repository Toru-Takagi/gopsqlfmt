package enumconv_test

import (
	"errors"
	"testing"

	"github.com/Toru-Takagi/gopsqlfmt/formatter/enumconv"
	pg_query "github.com/pganalyze/pg_query_go/v6"
	"github.com/stretchr/testify/assert"
)

func TestBoolExprTypeToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		exprType pg_query.BoolExprType
		want     string
		wantErr  error
	}{
		{
			name:     "AND",
			exprType: pg_query.BoolExprType_AND_EXPR,
			want:     "AND",
		},
		{
			name:     "OR",
			exprType: pg_query.BoolExprType_OR_EXPR,
			want:     "OR",
		},
		{
			name:     "NOT",
			exprType: pg_query.BoolExprType_NOT_EXPR,
			want:     "NOT",
		},
		{
			name:     "unknown",
			exprType: pg_query.BoolExprType(999),
			wantErr:  errors.New("BoolExprTypeToString: unknown BoolExprType"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := enumconv.BoolExprTypeToString(tt.exprType)
			assert.Equal(t, tt.want, actual)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
