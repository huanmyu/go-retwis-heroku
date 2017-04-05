package main

//type userForm struct {
//	username  string
//	password  string
//	password2 string
//}

//func Register(w http.ResponseWriter, r *http.Request) {
//	http.ServeFile(w, r, "public/register.html")
//}

//// Login when user login call
//func Login(w http.ResponseWriter, r *http.Request) {
//	var user userForm

//	decoder := json.NewDecoder(r.Body)
//	if err := decoder.Decode(&user); err != nil {
//		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
//		return
//	}
//	defer r.Body.Close()

//	u := User{
//		Username: user.username,
//	}

//	salt, password, auth, err := u.GetUserByName()
//	if err != nil {
//		respondWithError(w, http.StatusBadRequest, "user not found")
//		return
//	}

//	if md5PasswordWithSalt(salt, user.password) != password {
//		respondWithError(w, http.StatusBadRequest, "Password not correct")
//		return
//	}

//	oneYearAgo := time.Now().Add(time.Hour * 24 * 365)
//	if cookie, err := r.Cookie("auth"); err != nil {
//		http.SetCookie(w, &http.Cookie{Name: "auth", Value: auth, Expires: oneYearAgo})
//	} else {
//		cookie.Value = auth
//		cookie.Expires = oneYearAgo
//		http.SetCookie(w, cookie)
//	}
//}

//// IsLogin check if user is login
//func IsLogin(w http.ResponseWriter, r *http.Request) {
//	if cookie, err := r.Cookie("auth"); err != nil {
//		u := User{
//			Auth: cookie.Value,
//		}
//		err = u.GetUserByAuth()
//		if err != nil {
//			respondWithError(w, http.StatusBadRequest, "auth fail")
//			return
//		}
//	} else {
//		respondWithError(w, http.StatusBadRequest, "auth fail")
//	}
//}

//// Logout when user logout call
//func Logout(w http.ResponseWriter, r *http.Request) {
//	err := r.ParseForm()
//	if err != nil {
//		respondWithError(w, http.StatusBadRequest, "parameter miss")
//		return
//	}
//	userID := r.FormValue("userID")
//	u := User{
//		UserID: userID,
//	}
//	err = u.UpdateUserAuth()
//	if err != nil {
//		respondWithError(w, http.StatusBadRequest, "parameter miss")
//	}
//}
