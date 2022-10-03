package main

import (
	"context"
	"fmt"
	"golang-manipulate/connection"
	"golang-manipulate/middleware"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

// struct
type SesionData struct{
	IsLogin bool
	Username string
	FlashData string
}
var Data = SesionData{}
type Struktur struct {
	Name string
	Start_date time.Time
	End_date time.Time
	Deskripsi string
	Checkbox []string
	Gambar string
	Duration string
	Id int
	IsLogin bool
	FormatStartDate string
	FormatEndDate string
	Author string
}

type Structuser struct{
	ID int
	Name string
	Email string
	Password string
}

var iniArray = []Struktur{}
func main() {
	route := mux.NewRouter()
	connection.Dbkonek()
	// images
	route.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	// path prefix
	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))
	// routing

	route.HandleFunc("/",home).Methods("GET")
	route.HandleFunc("/add-blog", addProject).Methods("GET")
	route.HandleFunc("/store-blog", middleware.UploadFile(storeProject)).Methods("POST")
	route.HandleFunc("/detail-blog/{id}", detailProject).Methods("GET")
	route.HandleFunc("/edit/{id}", editProject).Methods("GET")
	route.HandleFunc("/update-blog/{id}", updateProject).Methods("POST")
	route.HandleFunc("/delete/{id}", deleteProject).Methods("GET")


	// auth
	route.HandleFunc("/form-register",formRegister).Methods("GET")
	route.HandleFunc("/register",register).Methods("POST")

	
	route.HandleFunc("/form-login",formLogin).Methods("GET")
	route.HandleFunc("/login",login).Methods("POST")

	route.HandleFunc("/logout",logout).Methods("GET")

	// route.HandleFunc("/contact", contact).Methods("GET")

	// server
	fmt.Println("server is runing on 127.0.0.1:5000")
	http.ListenAndServe("127.0.0.1:5000",route)
}

func home(res http.ResponseWriter, req *http.Request)  {
	res.Header().Set("Content-Type","text/html; charset=utf-8")
	theme, err := template.ParseFiles("views/blog/index.html")

	if err != nil {
		res.Write([]byte("massage : HACKER JANGAN MENYERANG !" + err.Error()))
	}

	// sesion
	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(req, "SESSION_KEY")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.Username = session.Values["Name"].(string)
	}

	fm := session.Flashes("message")

	var  flashes []string

	if len(fm) > 0{
		session.Save(req, res)
		for _, f1 := range fm {
			// meamasukan flash message
			flashes = append(flashes, f1.(string))
		}
	}

	Data.FlashData = strings.Join(flashes, "")	

	data,err := connection.Konekdb.Query(context.Background(), "SELECT id, name, description, durations, image, technologies FROM tb_projects")

	var result []Struktur

	for data.Next(){
		var each = Struktur{}

		err := data.Scan(&each.Id, &each.Name, &each.Deskripsi, &each.Duration, &each.Gambar, &each.Checkbox)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		result = append(result, each)
	}

		mapping := map[string]interface{}{
			"DataSesion" : Data,
			"show" :result,
		}

	theme.Execute(res, mapping)
}

func addProject(res http.ResponseWriter, req *http.Request)  {
	res.Header().Set("Content-Type","text/html; charset=utf-8")
	theme, err := template.ParseFiles("views/blog/addproject.html")

	if err != nil {
		res.Write([]byte("massage : HACKER JANGAN MENYERANG !" + err.Error()))
	}

	theme.Execute(res, nil)
}
func storeProject(res http.ResponseWriter, req *http.Request)  {
	err := req.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	title := req.PostForm.Get("title")
	start_date := req.PostForm.Get("start-date")
	end_date := req.PostForm.Get("end-date")
	desc := req.PostForm.Get("desc")
	var checkbox []string
	checkbox = req.Form["tech"]
	dataContext := req.Context().Value("dataFile")
	images := dataContext.(string)

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(req, "SESSION_KEY")

	// Mendapatkan author_id
	author := session.Values["ID"].(int)

	layouts := "2006-01-02"
	convStartDate, _ := time.Parse(layouts, start_date)  
	convEndtDate, _ := time.Parse(layouts, end_date)  

	hourse := convEndtDate.Sub(convStartDate).Hours()
	days := hourse/24
	weeks := days/7
	months := days/30
	years := months/12

	var duration string
	if days >= 1 && days <= 6 {
        duration = strconv.Itoa(int(days)) + " days"
    } else if days >= 7 && days <= 29 {
        duration = strconv.Itoa(int(weeks)) + " weeks"
    } else if days >= 30 && days <= 364 {
        duration = strconv.Itoa(int(months)) + " months"
    } else if days >= 365 {
        duration = strconv.Itoa(int(years)) + " years"
    }
	_, err = connection.Konekdb.Exec(context.Background(), "INSERT INTO tb_projects(name, start_date, end_date, description, durations, author_id, image, technologies) VALUES ($1,$2,$3,$4,$5,$6,$7, $8)",title, convStartDate,convEndtDate, desc, duration, author, images, checkbox)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(res, req, "/", http.StatusMovedPermanently)
}



// masih eror di ID NYAAA
func detailProject(res http.ResponseWriter, req *http.Request)  {

	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	theme, err := template.ParseFiles("views/blog/detail.html")

	if err != nil {
		res.Write([]byte("Hacker jangan menyerang! :" + err.Error()))
		return
	}
	var blogDetail = Struktur{}

	id, _ := strconv.Atoi(mux.Vars(req)["id"])

	err = connection.Konekdb.QueryRow(context.Background(), " SELECT id, name, start_date, end_date, description, durations, technologies FROM tb_projects WHERE id=$1", id).Scan(&blogDetail.Id, &blogDetail.Name, &blogDetail.Start_date, &blogDetail.End_date, &blogDetail.Deskripsi, &blogDetail.Duration, &blogDetail.Checkbox)

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte ("message ; " + err.Error()))
	}

	blogDetail.FormatStartDate = blogDetail.Start_date.Format("2 January 2006")
	blogDetail.FormatEndDate = blogDetail.End_date.Format("2 January 2006")
	data := map[string]interface{}{
		"detail" : blogDetail,
	}
	theme.Execute(res, data)
}
func editProject(res http.ResponseWriter, req *http.Request){
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/blog/edit-project.html")

	if err != nil {
		res.Write([]byte("message : "+ err.Error()))
		return
	}

	var editProject = Struktur{}

	id, _ := strconv.Atoi(mux.Vars(req)["id"])

	err = connection.Konekdb.QueryRow(context.Background(), " SELECT id, name, start_date, end_date, description, durations FROM tb_projects WHERE id=$1", id).Scan(&editProject.Id, &editProject.Name, &editProject.Start_date, &editProject.End_date, &editProject.Deskripsi, &editProject.Duration)

	editProject.FormatStartDate = editProject.Start_date.Format("2006-01-02")
	editProject.FormatEndDate = editProject.End_date.Format("2006-01-02")

	data := map[string]interface{}{
		"EditProject": editProject,
	}

	tmpl.Execute(res, data)
}

func updateProject(res http.ResponseWriter, req *http.Request){
	id, _ := strconv.Atoi(mux.Vars(req)["id"])
	
	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	title := req.PostForm.Get("nameProject")
	description := req.PostForm.Get("description")
	var checkbox []string
	checkbox = req.Form["tech"]
	start_date := req.PostForm.Get("startDate")
	end_date := req.PostForm.Get("endDate")
	layouts := "2006-01-02"
	convStartDate, _ := time.Parse(layouts, start_date)  
	convEndtDate, _ := time.Parse(layouts, end_date)  

	hourse := convEndtDate.Sub(convStartDate).Hours()
	days := hourse/24
	weeks := days/7
	months := days/30
	years := months/12

	var duration string
	if days >= 1 && days <= 6 {
        duration = strconv.Itoa(int(days)) + " days"
    } else if days >= 7 && days <= 29 {
        duration = strconv.Itoa(int(weeks)) + " weeks"
    } else if days >= 30 && days <= 364 {
        duration = strconv.Itoa(int(months)) + " months"
    } else if days >= 365 {
        duration = strconv.Itoa(int(years)) + " years"
    }
		
	_, err = connection.Konekdb.Exec(context.Background(), "UPDATE tb_projects SET name=$1, start_date=$2, end_date=$3, description=$4, technologies=$5, durations=$6 WHERE id=$7",title,convStartDate,convEndtDate, description, checkbox, duration, id)

	

	http.Redirect(res, req, "/", http.StatusFound)
}


// masih err
func deleteProject(res http.ResponseWriter, req *http.Request){
	id, _ := strconv.Atoi(mux.Vars(req)["id"])

	_,err := connection.Konekdb.Exec(context.Background(), "DELETE FROM tb_projects WHERE id=$1",id)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("message : " + err.Error()))
	}

	http.Redirect(res, req, "/", http.StatusMovedPermanently)
}

func formRegister(res http.ResponseWriter, req *http.Request)  {
	res.Header().Set("Content-Type","text/html; charset=utf-8")
	theme, err := template.ParseFiles("views/auth/register.html")

	if err != nil {
		res.Write([]byte("massage : HACKER JANGAN MENYERANG !" + err.Error()))
	}

	theme.Execute(res, nil)
}

func register(res http.ResponseWriter, req *http.Request)  {
	err := req.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	name := req.PostForm.Get("name")
	email := req.PostForm.Get("email")
	password := req.PostForm.Get("password")

	passwordhash, _ := bcrypt.GenerateFromPassword([]byte(password),10)

	_, err = connection.Konekdb.Exec(context.Background(), "INSERT INTO tb_users(name, email, password) VALUES ($1,$2,$3)",name, email,passwordhash)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(res, req, "/form-login", http.StatusMovedPermanently)	
}
func formLogin(res http.ResponseWriter, req *http.Request)  {
	res.Header().Set("Content-Type","text/html; charset=utf-8")
	theme, err := template.ParseFiles("views/auth/login.html")

	if err != nil {
		res.Write([]byte("massage : HACKER JANGAN MENYERANG !" + err.Error()))
	}

	theme.Execute(res, nil)
}

func login(res http.ResponseWriter, req *http.Request)  {
	err := req.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	email := req.PostForm.Get("email")
	password := req.PostForm.Get("password")


	user := Structuser{}

	// mengambil data email, dan melakukan pengecekan email
	err = connection.Konekdb.QueryRow(context.Background(), "SELECT * FROM tb_users WHERE email=$1",email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("message : " + err.Error()))
		return
	}

	// melakukan pengecekan password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(req, "SESSION_KEY")

	session.Values["Name"] = user.Name
	session.Values["Email"] = user.Email
	session.Values["ID"] = user.ID
	session.Values["IsLogin"] = true
// password jangan d masukin sesion bahaya!
	session.Options.MaxAge = 10800 //3jam(expired cookie)

	session.AddFlash("succesfull login","message")
	session.Save(req, res)


	http.Redirect(res, req, "/", http.StatusMovedPermanently)	
}

func logout(w http.ResponseWriter, r *http.Request) {

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/form-login", http.StatusSeeOther)
}