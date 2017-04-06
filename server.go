package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/bowenchen6/go-retwis-heroku/model"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

type errs struct {
	RegisterErr string
	LoginErr    string
}

type userPost struct {
	Username string
	Content  string
	Elapsed  string
}

type userInfo struct {
	User           *model.User
	FollowingCount int
	FollowersCount int
	UserPosts      []userPost
	PrevPosition   int
	NextPosition   int
	HavePage       bool
	HaveMore       bool
	Errstr         string
}

type timeline struct {
	IsLogin         bool
	LatestUsers     []string
	LatestUserPosts []userPost
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	u, err := IsLogin(w, r, 0, 0)
	var errstr string
	if err != nil {
		t, _ := template.ParseFiles("template/index.html")
		t.Execute(w, nil)
	} else {
		err = r.ParseForm()
		if err != nil {
			errstr = err.Error()
		}

		content := r.FormValue("status")
		p := model.Post{
			UserID:  u.User.UserID,
			Content: content,
		}

		err = p.CreatePost()
		if err != nil {
			errstr = err.Error()
		}

		u.Errstr = errstr

		postCount, _ := u.User.GetUserPostCount()
		u.User.GetUserPosts(0, 10)

		for i := range u.User.PostIDs {
			postID := u.User.PostIDs[i]
			p.GetPost(postID)
			up := userPost{
				Username: u.User.Username,
				Content:  p.Content,
				Elapsed:  strconv.FormatInt(p.CreatedAt, 10),
			}
			u.UserPosts = append(u.UserPosts, up)
		}

		if postCount > 10 {
			u.HaveMore = true
			u.NextPosition = 10
		} else {
			u.NextPosition = 0
			u.HaveMore = false
		}

		u.PrevPosition = 0
		u.HavePage = false
		t, _ := template.ParseFiles("template/home.html")
		t.Execute(w, &u)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	u := model.User{}
	if cookie, err := r.Cookie("auth"); err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		u.Auth = cookie.Value
		err = u.UpdateUserAuth()
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	u, err := IsLogin(w, r, 0, 10)
	if err != nil {
		err := r.ParseForm()
		if err != nil {
			errInfo := errs{
				LoginErr: err.Error(),
			}
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, &errInfo)
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "" {
			errInfo := errs{
				LoginErr: "username must be set",
			}
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, &errInfo)
			return
		}

		if password == "" {
			errInfo := errs{
				LoginErr: "password must be set",
			}
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, &errInfo)
			return
		}

		user := model.User{
			Username: username,
		}

		err = user.GetUserByName()
		if err != nil {
			errInfo := errs{
				LoginErr: "user not found",
			}
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, &errInfo)
			return
		}

		if model.Md5PasswordWithSalt(user.Salt, password) != user.Password {
			errInfo := errs{
				LoginErr: "password not correct",
			}
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, &errInfo)
			return
		}

		oneYearAgo := time.Now().Add(time.Hour * 24 * 365)
		if cookie, err := r.Cookie("auth"); err != nil {
			http.SetCookie(w, &http.Cookie{Name: "auth", Value: user.Auth, Expires: oneYearAgo})
		} else {
			cookie.Value = user.Auth
			cookie.Expires = oneYearAgo
			http.SetCookie(w, cookie)
		}
	}

	t, _ := template.ParseFiles("template/home.html")
	t.Execute(w, &u)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	u, err := IsLogin(w, r, 0, 10)
	if err != nil {
		err := r.ParseForm()
		if err != nil {
			errInfo := errs{
				RegisterErr: err.Error(),
			}
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, &errInfo)
		}

		username := r.FormValue("username")
		password := r.FormValue("password")
		password2 := r.FormValue("password2")

		if username == "" {
			errInfo := errs{
				RegisterErr: "username must be set",
			}
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, &errInfo)
			return
		}

		if password == "" || password2 == "" {
			errInfo := errs{
				RegisterErr: "password or repeat password must be set",
			}
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, &errInfo)
			return
		}

		if password != password2 {
			errInfo := errs{
				RegisterErr: "password not same",
			}
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, &errInfo)
			return
		}

		u := model.User{
			Username: username,
			Password: password,
		}

		err = u.CreateUser()
		if err != nil {
			errInfo := errs{
				RegisterErr: err.Error(),
			}
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, &errInfo)
			return
		}

		oneYearAgo := time.Now().Add(time.Hour * 24 * 365)
		if cookie, err := r.Cookie("auth"); err != nil {
			http.SetCookie(w, &http.Cookie{Name: "auth", Value: u.Auth, Expires: oneYearAgo})
		} else {
			cookie.Value = u.Auth
			cookie.Expires = oneYearAgo
			http.SetCookie(w, cookie)
		}

		t, _ := template.ParseFiles("template/register.html")
		t.Execute(w, &u)
		return
	}

	t, _ := template.ParseFiles("template/home.html")
	t.Execute(w, &u)
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/favicon.ico")
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	startValue := r.FormValue("start")
	var start int
	if s, err := strconv.Atoi(startValue); err == nil {
		start = s
	}

	u, err := IsLogin(w, r, start, 10)
	if err != nil {
		t, _ := template.ParseFiles("template/index.html")
		t.Execute(w, nil)
		return
	}

	t, _ := template.ParseFiles("template/home.html")
	t.Execute(w, &u)
}

func timelineHandler(w http.ResponseWriter, r *http.Request) {
	//	var u model.User{}
	//	var p model.Post{}
	//	var t timeline{}

	//	cookie, err := r.Cookie("auth")
	//	if err != nil {

	//	}

	//	u.Auth = cookie.Value
	//	err = u.GetUserByAuth()
	//	if err != nil {

	//	}
	//	t.LatestUsers, err := u.GetLastUsers()
	//	t.LatestPosts, err = p.GetTimelinePosts()
	//	t, _ := template.ParseFiles("template/home.html")
	//	t.Execute(w, &u)

}

// IsLogin check user is or not login
func IsLogin(w http.ResponseWriter, r *http.Request, start, count int) (userInfo, error) {
	u := model.User{}
	p := model.Post{}
	data := userInfo{}
	cookie, err := r.Cookie("auth")
	if err != nil {
		return data, err
	}

	u.Auth = cookie.Value
	err = u.GetUserByAuth()
	if err != nil {
		return data, err
	}

	if count > 0 {
		postCount, _ := u.GetUserPostCount()
		err = u.GetUserPosts(start, count)
		if err != nil {
			return data, err
		}

		for i := range u.PostIDs {
			postID := u.PostIDs[i]
			err = p.GetPost(postID)
			if err != nil {
				return data, err
			}

			up := userPost{
				Username: u.Username,
				Content:  p.Content,
				Elapsed:  strconv.FormatInt(p.CreatedAt, 10),
			}
			data.UserPosts = append(data.UserPosts, up)
		}

		nextPosition := start + count
		if postCount > int64(nextPosition) {
			data.HaveMore = true
			data.NextPosition = nextPosition
		} else {
			data.NextPosition = 0
			data.HaveMore = false
		}

		if start > 0 {
			data.PrevPosition = start - count
			data.HavePage = true
		} else {
			data.PrevPosition = 0
			data.HavePage = false
		}
	}

	data.User = &u
	data.FollowingCount = len(u.Following)
	data.FollowersCount = len(u.Followers)

	return data, nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	mux := mux.NewRouter()
	mux.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/favicon.ico", faviconHandler)
	mux.HandleFunc("/register", registerHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)

	mux.HandleFunc("/post", postHandler)
	mux.HandleFunc("/timeline", timelineHandler)
	// Includes some default middlewares
	n := negroni.Classic()
	n.UseHandler(mux)
	n.Run(":" + port)
}
