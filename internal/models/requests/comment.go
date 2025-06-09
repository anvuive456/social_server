package requests

type GetPostCommentsRequest struct {
	PostID    uint32 `json:"post_id"`
	Limit     int    `json:"limit" binding:"required"`
	Sort      string `json:"sort,omitempty"`
	SortOrder string `json:"sort_order,omitempty"`
	After     string `json:"after,omitempty"`
	Before    string `json:"before,omitempty"`
}
