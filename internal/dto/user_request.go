package dto

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=4"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email" binding:"required,email"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=4"`
	Password string `json:"password" binding:"required,min=8"`
}

type AddUserRequest struct {
	Username string `json:"username" binding:"required,min=4"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email" binding:"required,email"`
	Role     string `json:"role" binding:"required,oneof=admin user banned"`
}

type UpdateUserRequest struct {
	ID       uint    `json:"id" binding:"required"`
	Username *string `json:"username" binding:"omitempty,min=4"`
	Password *string `json:"password" binding:"omitempty,min=8"`
	Email    *string `json:"email" binding:"omitempty,email"`
	Role     *string `json:"role" binding:"omitempty,oneof=admin user banned"`
}
