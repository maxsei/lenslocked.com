package controllers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"lenslocked.com/context"
	"lenslocked.com/models"
	"lenslocked.com/views"
)

const (
	NamedGalleryShowRoute = "galleries_show"
	NamedGalleryEditRoute = "galleries_edit"
	maxMultipartMem       = 1 << 20 //1 megabyte
)

func NewGalleries(gs models.GalleryService, r *mux.Router) *Galleries {
	return &Galleries{
		New:       views.NewView("bootstrap", "galleries/new"),
		ShowView:  views.NewView("bootstrap", "galleries/show"),
		EditView:  views.NewView("bootstrap", "galleries/edit"),
		IndexView: views.NewView("bootstrap", "galleries/index"),
		gs:        gs,
		r:         r,
	}
}

type Galleries struct {
	New       *views.View
	ShowView  *views.View
	EditView  *views.View
	IndexView *views.View
	gs        models.GalleryService
	r         *mux.Router
}

type GalleryForm struct {
	Title string `schema:"title"`
}

// GET /galleries
func (g *Galleries) Index(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	galleries, err := g.gs.ByUserID(user.ID)
	if err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	var vd views.Data
	vd.Yeild = galleries
	g.IndexView.Render(w, r, vd)
}

// GET /galleries/:id
func (g *Galleries) Show(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	var vd views.Data
	vd.Yeild = gallery
	g.ShowView.Render(w, r, vd)
}

// GET /galleries/:id/edit
func (g *Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}
	var vd views.Data
	vd.Yeild = gallery
	g.EditView.Render(w, r, vd)
}

// GET /galleries/:id/edit
func (g *Galleries) Update(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}
	var vd views.Data
	vd.Yeild = gallery
	var form GalleryForm
	if err := parseForm(r, &form); err != nil {
		log.Println(err)
		vd.ErrorAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	gallery.Title = form.Title
	if err := g.gs.Update(gallery); err != nil {
		vd.ErrorAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	vd.SuccessAlert("Gallery successfully updated!")
	g.EditView.Render(w, r, vd)
}

// GET /galleries/:id/images
func (g *Galleries) Upload(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}
	// TODO: parse a mutlipare form
	var vd views.Data
	vd.Yeild = gallery
	err = r.ParseMultipartForm(maxMultipartMem)
	if err != nil {
		vd.ErrorAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	// create directory to put our image in
	galleryPath := fmt.Sprintf("images/galleries/%v/", gallery.ID)
	err = os.MkdirAll(galleryPath, 0755)
	if err != nil {
		vd.ErrorAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	files := r.MultipartForm.File["images"]
	for _, f := range files {
		file, err := f.Open()
		if err != nil {
			vd.ErrorAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}
		defer file.Close()
		dst, err := os.Create(galleryPath + f.Filename)
		if err != nil {
			vd.ErrorAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}
		defer dst.Close()
		_, err = io.Copy(dst, file)
		if err != nil {
			vd.ErrorAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}
	}
	// g.EditView.Render(w, r, vd)
	fmt.Fprintln(w, "Files successfully uploaded.")
}

// POST /galleries
func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form GalleryForm
	if err := parseForm(r, &form); err != nil {
		log.Println(err)
		vd.ErrorAlert(err)
		g.New.Render(w, r, vd)
		return
	}
	user := context.User(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	gallery := models.Gallery{
		Title:  form.Title,
		UserID: user.ID,
	}
	if err := g.gs.Create(&gallery); err != nil {
		vd.ErrorAlert(err)
		g.New.Render(w, r, vd)
		return
	}
	url, err := g.r.Get(NamedGalleryEditRoute).URL("id", fmt.Sprintf("%v", gallery.ID))
	if err != nil {
		//todo make this go to the index page
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
	fmt.Fprint(w, gallery)
}

func (g *Galleries) Delete(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}
	var vd views.Data
	if err := g.gs.Delete(gallery.ID); err != nil {
		vd.ErrorAlert(err)
		vd.Yeild = gallery
		g.EditView.Render(w, r, vd)
		return
	}
	http.Redirect(w, r, "/galleries", http.StatusFound)
}

func (g *Galleries) galleryByID(w http.ResponseWriter, r *http.Request) (*models.Gallery, error) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid gallery ID", http.StatusNotFound)
		return nil, err
	}
	gallery, err := g.gs.ByID(uint(id))
	switch err {
	case models.ErrNotFound:
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return nil, err
	case nil:
		break
	default:
		http.Error(w, "Whoops! Something went wrong.", http.StatusInternalServerError)
		return nil, err
	}
	return gallery, nil
}
