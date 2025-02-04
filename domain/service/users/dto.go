package users

type CreateUserDTO struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Name     string `json:"username"`
	RealName string `json:"real_name"`
}

type UpdateUserDTO struct {
	Phone       string `json:"phone"`
	OldPassword string `json:"last_pass"`
	Password    string `json:"password"`
	Name        string `json:"username"`
	RealName    string `json:"real_name"`
}

type FindUserDTO struct {
	UserId   int    `json:"user_id"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
}

type UserResponseDTO struct {
	Id          int    `json:"id,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Name        string `json:"username,omitempty"`
	RealName    string `json:"real_name"`
	ContactName string `json:"contact_name,omitempty"`
	ShowPhone   bool   `json:"show_phone,omitempty"`
}

type UpdatePasswordDTO struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type CreateContactDTO struct {
	Phone string `json:"phone"`
	Name  string `json:"name"`
}
