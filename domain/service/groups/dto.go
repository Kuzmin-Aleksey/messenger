package groups

type UpdateGroupDTO struct {
	Name string `json:"name"`
}

type GetUsersDTO struct {
	UserId int    `json:"user_id"`
	Role   string `json:"role"`
}
