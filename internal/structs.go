package internal 

type UserApi struct {
	Gender string `json:"gender"`
	Name struct {
		First string `json:"first"`
		Last string `json:"last"`
	} `json:"name"`
	Email string `json:"email"`
	Login struct {
		UUID string `json:"uuid"`
	} `json:"login"`
}

type User struct {
	UUID string `json:"uuid"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Email string `json:"email"`
	Gender string `json:"gender"`
}

type ResponseRandomUsersApi struct {
	Results []UserApi `json:"results"`
	Error string `json:"error"`
}

func (ua UserApi) convertToRawUser() User {
	return User{
		UUID: ua.Login.UUID,
		FirstName: ua.Name.First,
		LastName: ua.Name.Last,
		Email: ua.Email,
		Gender: ua.Gender,
	}
}