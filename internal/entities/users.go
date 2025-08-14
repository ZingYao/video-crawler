package entities

import "time"

type UserEntity struct {
	Id          string    `json:"id"`
	Nickname    string    `json:"nickname"`
	Username    string    `json:"username"`
	Salt        string    `json:"salt"`
	Password    string    `json:"password"`
	IsAdmin     bool      `json:"is_admin"`
	AllowLogin  bool      `json:"allow_login"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastLoginAt time.Time `json:"last_login_at"`
	LoginCount  int       `json:"login_count"`
}

type (
	LoginRequest struct {
		Username string `json:"username" binding:"required"` // 用户名 必填
		Password string `json:"password" binding:"required"` // 密码 必填
	}
	LoginResponse struct {
		Id       string `json:"id"`
		Nickname string `json:"nickname"`
		Token    string `json:"token"`
		IsAdmin  *bool  `json:"is_admin,omitempty"`
	}
)

type RegisterRequest struct {
	LoginRequest
	Nickname string `json:"nickname"` // 昵称
}

type UserDetailResponse struct {
	Id           string         `json:"id"`
	Nickname     string         `json:"nickname"`
	Username     string         `json:"username"`
	IsAdmin      bool           `json:"is_admin"`
	AllowLogin   bool           `json:"allow_login"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	LastLoginAt  time.Time      `json:"last_login_at"`
	LoginCount   int            `json:"login_count"`
	LoginHistory []LoginHistory `json:"login_history"`
}

type UserSaveRequest struct {
	Nickname   string `json:"nickname" binding:"required"`
	Username   string `json:"username" binding:"required"`
	UserId     string `json:"user_id" binding:"required"`
	IsAdmin    bool   `json:"is_admin"`
	AllowLogin bool   `json:"allow_login"`
	Password   string `json:"password"`
}

type UserList struct {
	Id          string    `json:"id"`
	Nickname    string    `json:"nickname"`
	Username    string    `json:"username"`
	IsAdmin     bool      `json:"is_admin"`
	AllowLogin  bool      `json:"allow_login"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastLoginAt time.Time `json:"last_login_at"`
	LoginCount  int       `json:"login_count"`
}

type AllowLoginStatusChangeRequest struct {
	UserId     string `json:"user_id" binding:"required"`
	AllowLogin bool   `json:"allow_login"`
}

type UserDeleteRequest struct {
	UserId string `json:"user_id" binding:"required"`
}
