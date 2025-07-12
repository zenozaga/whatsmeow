package sqlutil

import (
	"fmt"
	"strings"

	"go.mau.fi/util/dbutil"
)

type SqlOptionsSortBy string

type LimitOption struct {
	Limit  int
	Offset int
}

type SqlLimitOptions struct {
	SortBy         *SqlOptionsSortBy
	SortDescending *bool
	Limit          *LimitOption
}

// ToSQL converts the SqlLimitOptions to a SQL string that can be appended to a query.
// It handles sorting and limiting based on the provided options.
// If no options are set, it returns an empty string.
//
// example usage:
//
//	options := &SqlLimitOptions{
//	    SortBy:         ptr.To(SqlOptionsSortBy("created_at")),
//	    SortDescending: ptr.To(true),
//	    Limit:          &LimitOption{Limit: 10, Offset: 0},
//	}
//	sqlFragment := options.ToSQL(dbutil.Postgres)
//	// sqlFragment will be "ORDER BY created_at DESC LIMIT 10 OFFSET 0" if dialect is Postgres
//	// or "ORDER BY created_at DESC LIMIT 10 0" if dialect is SQLite
func (c *SqlLimitOptions) ToSQL(dialect dbutil.Dialect) string {

	// If no limit or sort options are set, return an empty string
	if c.Limit == nil && c.SortBy == nil {
		return ""
	}

	sql := []string{}

	if c.SortBy != nil && *c.SortBy != "" {
		sort := fmt.Sprintf("ORDER BY %s", *c.SortBy)
		if c.SortDescending != nil && *c.SortDescending {
			sort += " DESC"
		} else {
			sort += " ASC"
		}

		sql = append(sql, sort)
	}

	if c.Limit != nil {
		limit := fmt.Sprintf("LIMIT %d %d", c.Limit.Offset, c.Limit.Limit)

		if dialect == dbutil.Postgres {
			limit = fmt.Sprintf("LIMIT %d OFFSET %d", c.Limit.Limit, c.Limit.Offset)
		}

		sql = append(sql, limit)

	}

	if len(sql) == 0 {
		return ""
	}

	return strings.Join(sql, " ")

}
