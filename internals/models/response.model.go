package models

type SuccessResponse struct {
	Success bool `json:"success" example:"true"`
	Status  int  `json:"status" example:"200"`
	Data    any  `json:"-"`
}

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Status  int    `json:"status" example:"500"`
	Error   string `json:"error" example:"error message"`
}

type InternalServerError struct {
	ErrorResponse
	Status int `example:"500"`
}

type BadRequestError struct {
	ErrorResponse
	Status int `example:"400"`
}
