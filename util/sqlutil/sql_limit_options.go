package sqlutil

import (
	"fmt"
	"strings"

	"go.mau.fi/util/dbutil"
)

type SqlOptionsSortBy string

type PagingOption struct {
	Offset *int
	Limit  int
}

type SqlLimitOptions struct {
	SortBy         *SqlOptionsSortBy
	SortDescending *bool
	Paging         *PagingOption
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
//	    Limit:          &PagingOption{Limit: 10, Offset: 0},
//	}
//	sqlFragment := options.ToSQL(dbutil.Postgres)
//	// sqlFragment will be "ORDER BY created_at DESC LIMIT 10 OFFSET 0" if dialect is Postgres
//	// or "ORDER BY created_at DESC LIMIT 10 0" if dialect is SQLite
func (c *SqlLimitOptions) ToSQL(dialect dbutil.Dialect) string {

	// If no limit or sort options are set, return an empty string
	if c.Paging == nil && c.SortBy == nil {
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

	if c.Paging != nil {

		var sqlPaging string
		var parts []string

		parts = append(parts, fmt.Sprintf("LIMIT %d", c.Paging.Limit))

		if c.Paging.Offset != nil && *c.Paging.Offset > 0 {
			parts = append(parts, fmt.Sprintf("OFFSET %d", *c.Paging.Offset))
		}

		sqlPaging = strings.Join(parts, " ")
		sql = append(sql, sqlPaging)
	}

	if len(sql) == 0 {
		return ""
	}

	return strings.Join(sql, " ")

}
