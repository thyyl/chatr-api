package user

// ============================================================
// Request
// ============================================================
type CreateLocalUserRequest struct {
	Name string `json:"name" binding:"required"`
}

type GetUserRequest struct {
	Id string `form:"id" binding:"required"`
}

// ============================================================
// Response
// ============================================================
type UserDto struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Photo string `json:"photo"`
}

type GoogleUserDto struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Photo string `json:"photo"`
}
