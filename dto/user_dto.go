package dto

import "fetchTest/model"

type UserDto struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

func ToUserDto(user model.User) UserDto {
	return UserDto{
		Username: user.Username,
		Email:    user.Email,
	}
}

type QueryPageDto struct {
	Contestantname string  `json:"contestantname"`
	Rank           int     `json:"rank"`
	Oldrating      float64 `json:"oldrating"`
	Newrating      float64 `json:"newrating"`
	Deltarating    float64 `json:"deltarating"`
	Dataregion     string  `json:"dataregion"`
}

//func ToQueryPageDto() {
//
//}
