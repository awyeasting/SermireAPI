package models 

type User struct {
	Email		string `json:"email"`
	Username 	string `json:"username"`
	FirstName 	string `json:"firstname"`
	LastName 	string `json:"lastname"`
	Password 	string `json:"password"`
	Token 		string `json:"token"`
}

type ResponseResult struct {
	Error		string `json:"error"`
	Result 		string `json:"result"`
}