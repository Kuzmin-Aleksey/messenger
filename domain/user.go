package domain

type CreateUserInput struct {
	Name     string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	Chats []Chat `json:"chats,omitempty"`
}

type UserInfo struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}
