package paginators

import (
	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
)

func CreateProfilePaginator(
	cursor paginator.Cursor,
	order *paginator.Order,
	limit *int,
) *paginator.Paginator {
	opts := []paginator.Option{
		&paginator.Config{
			Keys:  []string{"ID"},
			Limit: 10,
			Order: paginator.DESC,
		},
	}
	if limit != nil {
		opts = append(opts, paginator.WithLimit(*limit))
	}
	if order != nil {
		opts = append(opts, paginator.WithOrder(*order))
	}
	if cursor.After != nil {
		opts = append(opts, paginator.WithAfter(*cursor.After))
	}
	if cursor.Before != nil {
		opts = append(opts, paginator.WithBefore(*cursor.Before))
	}
	return paginator.New(opts...)
}
