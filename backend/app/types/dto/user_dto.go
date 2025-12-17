package dto

type UserDTO struct {
	UserID   uint
	Username string
	NickName string
	Avatar   string
}

type TokenDTO struct {
	AccessToken  string
	RefreshToken string
}
