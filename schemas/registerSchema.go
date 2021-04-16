package schemas

type RegisterUser struct {
	Username        string `json:"username" binding:"required,min=3,max=100"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=6,max=100"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}
