package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"lenslocked.com/controllers"
	"lenslocked.com/middleware"
	"lenslocked.com/models"
	"lenslocked.com/rand"
)

func main() {
	cfgReq := flag.Bool("prod", false, "set to true in production to ensure that a .config is used when provided")
	flag.Parse()

	cfg, err := LoadConfig(*cfgReq)
	must(err)
	dbCnfg := DefaultPostgresConfig()
	services, err := models.NewServices(
		models.WithGorm(dbCnfg.Dialect(), dbCnfg.ConnectionInfo()),
		models.WithUser(cfg.Pepper, cfg.HMACKey),
		models.WithGallery(),
		models.WithImage(),
		models.WithLogMode(!cfg.InProd()),
	)
	must(err)
	defer services.Close()
	services.AutoMigrate()
	// services.DestructiveReset()

	r := mux.NewRouter()
	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers(services.User)
	galleriesC := controllers.NewGalleries(services.Gallery, services.Image, r)

	b, err := rand.Bytes(32)
	must(err)
	csrfMw := csrf.Protect(b, csrf.Secure(cfg.InProd()))

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
	// FileServer for static assets
	assetHandler := http.FileServer(http.Dir("./assets/"))
	assetHandler = http.StripPrefix("/assets/", assetHandler)
	r.PathPrefix("/assets/").Handler(assetHandler)
	// Image Routes
	imageHandler := http.FileServer(http.Dir("./images/"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))
	// Gallery Routes
	r.HandleFunc("/galleries", OwnerMw.ApplyFn(galleriesC.Index)).Methods("GET")
	r.Handle("/galleries/new", OwnerMw.Apply(galleriesC.New)).Methods("GET")
	r.HandleFunc("/galleries", OwnerMw.ApplyFn(galleriesC.Create)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images", OwnerMw.ApplyFn(galleriesC.UploadImages)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images/{filename}/delete", OwnerMw.ApplyFn(galleriesC.DeleteImages)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/update", OwnerMw.ApplyFn(galleriesC.Update)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/delete", OwnerMw.ApplyFn(galleriesC.Delete)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesC.Show).
		Methods("GET").Name(controllers.NamedGalleryShowRoute)
	r.HandleFunc("/galleries/{id:[0-9]+}/edit", OwnerMw.ApplyFn(galleriesC.Edit)).
		Methods("GET").Name(controllers.NamedGalleryEditRoute)
	// TODO: config this

	fmt.Printf("listening and serving on port %d\n", cfg.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), csrfMw(userMw.Apply(r)))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
