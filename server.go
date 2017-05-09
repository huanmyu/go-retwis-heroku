package main

import (
	"fmt"
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

type userPosts struct {
	Posts        []userPost
	PrevPosition int
	NextPosition int
	HavePage     bool
	HaveMore     bool
}

type userInfo struct {
	User           *model.User
	FollowingCount int
	FollowersCount int
	UserPosts      *userPosts
	Errstr         string
}

type timeline struct {
	IsLogin         bool
	LatestUsers     []string
	LatestUserPosts []userPost
}

type profile struct {
	User        *model.User
	IsLogin     bool
	IsSelf      bool
	IsFollowing bool
	UserPosts   *userPosts
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	u := &userInfo{}
	err := IsLogin(w, r, 0, 0, u)
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

		if errstr != "" {
			u.Errstr = errstr
			u.UserPosts = &userPosts{}
			err = getUserPosts(0, 10, u.User, u.UserPosts)
			if err != nil {
				t, _ := template.ParseFiles("template/index.html")
				t.Execute(w, nil)
			}

			t, _ := template.ParseFiles("template/home.html")
			t.Execute(w, &u)
		} else {
			http.Redirect(w, r, "/", http.StatusFound)
		}
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	u := model.User{}
	cookie, err := r.Cookie("auth")
	if err != nil {
		t, _ := template.ParseFiles("template/index.html")
		t.Execute(w, nil)
		return
	}

	u.Auth = cookie.Value
	err = u.UpdateUserAuth()
	if err != nil {
		t, _ := template.ParseFiles("template/index.html")
		t.Execute(w, nil)
		return
	}

	cookie.Expires = time.Now().AddDate(-1, 0, 0)
	http.SetCookie(w, cookie)
	t, _ := template.ParseFiles("template/index.html")
	t.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	u := &userInfo{}
	err := IsLogin(w, r, 0, 10, u)
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

	http.Redirect(w, r, "/", http.StatusFound)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	u := &userInfo{}
	err := IsLogin(w, r, 0, 10, u)
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

		user := model.User{
			Username: username,
			Password: password,
		}

		err = user.CreateUser()
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
			http.SetCookie(w, &http.Cookie{Name: "auth", Value: user.Auth, Expires: oneYearAgo})
		} else {
			cookie.Value = user.Auth
			cookie.Expires = oneYearAgo
			http.SetCookie(w, cookie)
		}

		t, _ := template.ParseFiles("template/register.html")
		t.Execute(w, &user)
	} else {
		t, _ := template.ParseFiles("template/register.html")
		t.Execute(w, u.User)
	}
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/favicon.ico")
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	var start int
	err := r.ParseForm()
	if err != nil {
		t, _ := template.ParseFiles("template/index.html")
		t.Execute(w, nil)
	}

	startValue := r.FormValue("start")
	if s, err := strconv.Atoi(startValue); err == nil {
		start = s
	}

	u := &userInfo{}
	err = IsLogin(w, r, start, 10, u)
	if err != nil {
		t, _ := template.ParseFiles("template/index.html")
		t.Execute(w, nil)
		return
	}

	t, _ := template.ParseFiles("template/home.html")
	t.Execute(w, &u)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	var start int
	p := profile{}
	user := model.User{}
	err := r.ParseForm()
	if err != nil {
		log.Println("params miss")
	}

	startValue := r.FormValue("start")
	if s, err := strconv.Atoi(startValue); err == nil {
		start = s
	}

	p.User = &model.User{}
	p.User.Username = r.FormValue("username")

	cookie, err := r.Cookie("auth")
	if err != nil {
		p.IsLogin = false
	} else {
		user.Auth = cookie.Value
		err = user.GetUserByAuth()
		if err != nil {
			p.IsLogin = false
		} else {
			p.IsLogin = true
		}
	}

	err = p.User.GetUserByName()
	if err != nil {
		log.Println("profile user not found")
	}

	if user.UserID == p.User.UserID {
		p.IsSelf = true
		p.IsFollowing = false
	} else {
		p.IsSelf = false
		p.IsFollowing, err = user.IsFollowing(p.User)
		if err != nil {
			log.Println("No Following")
		}
	}

	p.UserPosts = &userPosts{}
	err = getUserPosts(start, 10, p.User, p.UserPosts)
	if err != nil {
		log.Println("No Posts")
	}

	t, _ := template.ParseFiles("template/profile.html")
	t.Execute(w, &p)
}

func timelineHandler(w http.ResponseWriter, r *http.Request) {
	u := model.User{}
	p := model.Post{}
	t := timeline{}

	cookie, err := r.Cookie("auth")
	if err != nil {
		t.IsLogin = false
	} else {
		u.Auth = cookie.Value
		err = u.GetUserByAuth()
		if err != nil {
			t.IsLogin = false
		} else {
			t.IsLogin = true
		}
	}

	t.LatestUsers, err = u.GetLastUsers()
	if err != nil {
		log.Println("No User Register!")
	}
	LatestPostIDs, err := p.GetTimelinePosts()
	if err != nil {
		log.Println("No Posts!")
	}
	for i := range LatestPostIDs {
		postID := LatestPostIDs[i]
		err = p.GetPost(postID)
		if err != nil {
			log.Println("No Post with id: " + postID)
		}

		user := model.User{
			UserID: p.UserID,
		}

		var username string
		err = user.GetUserByUserID()
		if err != nil {
			username = ""
		} else {
			username = user.Username
		}

		up := userPost{
			Username: username,
			Content:  p.Content,
			Elapsed:  fmt.Sprintf("%v", time.Since(time.Unix(p.CreatedAt, 0))),
		}
		t.LatestUserPosts = append(t.LatestUserPosts, up)
	}

	tem, _ := template.ParseFiles("template/timeline.html")
	tem.Execute(w, &t)
}

func followHandler(w http.ResponseWriter, r *http.Request) {
	u := model.User{}
	follow := model.User{}
	err := r.ParseForm()
	if err != nil {
		log.Println("params miss")
	}

	cookie, err := r.Cookie("auth")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	u.Auth = cookie.Value
	err = u.GetUserByAuth()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	userIDStr := r.FormValue("userid")
	if s, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
		follow.UserID = s
	} else {
		log.Println("params error")
	}
	follow.GetUserByUserID()

	following := r.FormValue("following")

	u.AddOrRemFollowingUser(follow, following)
	http.Redirect(w, r, "/profile?username="+follow.Username, http.StatusFound)
}

// IsLogin check user is or not login
func IsLogin(w http.ResponseWriter, r *http.Request, start, count int, data *userInfo) error {
	data.User = &model.User{}
	cookie, err := r.Cookie("auth")
	if err != nil {
		return err
	}

	data.User.Auth = cookie.Value
	err = data.User.GetUserByAuth()
	if err != nil {
		return err
	}

	data.UserPosts = &userPosts{}
	err = getUserPosts(start, count, data.User, data.UserPosts)
	if err != nil {
		return err
	}

	data.FollowingCount = len(data.User.Following)
	data.FollowersCount = len(data.User.Followers)

	return nil
}

func getUserPosts(start, count int, u *model.User, ups *userPosts) error {
	if count > 0 {
		p := model.Post{}
		postCount, err := u.GetUserPostCount()
		err = u.GetUserPosts(start, count)
		if err != nil {
			return err
		}

		for i := range u.PostIDs {
			postID := u.PostIDs[i]
			err = p.GetPost(postID)
			if err != nil {
				return err
			}

			up := userPost{
				Username: u.Username,
				Content:  p.Content,
				Elapsed:  fmt.Sprintf("%v", time.Since(time.Unix(p.CreatedAt, 0))),
			}
			ups.Posts = append(ups.Posts, up)
		}

		nextPosition := start + count
		if postCount > int64(nextPosition) {
			ups.HaveMore = true
			ups.NextPosition = nextPosition
		} else {
			ups.NextPosition = 0
			ups.HaveMore = false
		}

		if start > 0 {
			ups.PrevPosition = start - count
			ups.HavePage = true
		} else {
			ups.PrevPosition = 0
			ups.HavePage = false
		}
	}
	return nil
}

// danger
func flushHandler(w http.ResponseWriter, r *http.Request) {
	err := model.FlushDB()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithIndentJSON(w, http.StatusOK, map[string]string{"code": "200", "result": "flush success"})
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	u := model.User{}

	var IsLogin bool
	cookie, err := r.Cookie("auth")
	if err != nil {
		IsLogin = false
	} else {
		u.Auth = cookie.Value
		err = u.GetUserByAuth()
		if err != nil {
			IsLogin = false
		} else {
			IsLogin = true
		}
	}

	_, err = u.GetLastUsers()
	if err != nil {
		log.Println("No User Register!")
	}
	c := struct {
		IsLogin bool
	}{
		IsLogin: IsLogin,
	}
	tem, _ := template.ParseFiles("template/chat.html")
	tem.Execute(w, c)
}

func wordHandler(w http.ResponseWriter, r *http.Request) {
	word := model.Word{Name: "go"}
	err := word.CreateWord()
	if err != nil {
		log.Println(err)
	}
	respondWithIndentJSON(w, http.StatusOK, word)
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
	mux.HandleFunc("/profile", profileHandler)
	mux.HandleFunc("/follow", followHandler)
	mux.HandleFunc("/flush", flushHandler)

	mux.HandleFunc("/chat", chatHandler)

	mux.HandleFunc("/word", wordHandler)

	// Includes some default middlewares
	n := negroni.Classic()
	n.UseHandler(mux)
	n.Run(":" + port)
}
