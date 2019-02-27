package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"lenslocked.com/controllers"
	"lenslocked.com/middleware"
	"lenslocked.com/models"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "$m0kycat"
	dbname   = "lenslocked_dev"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)
	services, err := models.NewServices(psqlInfo)
	must(err)
	defer services.Close()
	services.AutoMigrate()
	// services.DestructiveReset()

	r := mux.NewRouter()
	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers(services.User)
	galleriesC := controllers.NewGalleries(services.Gallery, services.Image, r)

	userMw := middleware.User{UserService: services.User}
	OwnerMw := middleware.Owner{User: userMw}

	/*
		Remember routes are prioritized on a first come first serve basis
		That is the routes that are declared first are also handled first
	*/
	// Standard Routes
	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	// User Routes
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	// Image Routes
	imageHandler := http.FileServer(http.Dir("./images/"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))
	// Gallery Routes
	r.HandleFunc("/galleries", OwnerMw.ApplyFn(galleriesC.Index)).Methods("GET")
	r.Handle("/galleries/new", OwnerMw.Apply(galleriesC.New)).Methods("GET")
	r.HandleFunc("/galleries", OwnerMw.ApplyFn(galleriesC.Create)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images", OwnerMw.ApplyFn(galleriesC.Upload)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/update", OwnerMw.ApplyFn(galleriesC.Update)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/delete", OwnerMw.ApplyFn(galleriesC.Delete)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesC.Show).
		Methods("GET").Name(controllers.NamedGalleryShowRoute)
	r.HandleFunc("/galleries/{id:[0-9]+}/edit", OwnerMw.ApplyFn(galleriesC.Edit)).
		Methods("GET").Name(controllers.NamedGalleryEditRoute)

	http.ListenAndServe(":8080", userMw.Apply(r))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
