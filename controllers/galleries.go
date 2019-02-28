package controllers

import (
	"fmt"
	"log"
	"net/http"
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

func NewGalleries(gs models.GalleryService, is models.ImageService, r *mux.Router) *Galleries {
	return &Galleries{
		New:       views.NewView("bootstrap", "galleries/new"),
		ShowView:  views.NewView("bootstrap", "galleries/show"),
		EditView:  views.NewView("bootstrap", "galleries/edit"),
		IndexView: views.NewView("bootstrap", "galleries/index"),
		gs:        gs,
		is:        is,
		r:         r,
	}
}

type Galleries struct {
	New       *views.View
	ShowView  *views.View
	EditView  *views.View
	IndexView *views.View
	gs        models.GalleryService
	is        models.ImageService
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
		log.Print(err)
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

// POST /galleries/:id/images
func (g *Galleries) UploadImages(w http.ResponseWriter, r *http.Request) {
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
	err = r.ParseMultipartForm(maxMultipartMem)
	if err != nil {
		vd.ErrorAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	files := r.MultipartForm.File["images"]
	for _, f := range files {
		file, ferr := f.Open()
		if ferr != nil {
			vd.ErrorAlert(ferr)
			g.EditView.Render(w, r, vd)
			return
		}
		img := &models.Image{
			GalleryID: gallery.ID,
			Filename:  f.Filename,
		}
		err = g.is.Create(img, file)
		if err != nil {
			vd.ErrorAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}
	}
	url, err := g.r.Get(NamedGalleryEditRoute).URL("id", fmt.Sprintf("%v", gallery.ID))
	if err != nil {
		http.Redirect(w, r, "/galleries", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

// POST /galleries/:id/images/filename/delete
func (g *Galleries) DeleteImages(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}
	filename := mux.Vars(r)["filename"]
	img := &models.Image{
		GalleryID: gallery.ID,
		Filename:  filename,
	}
	if err := g.is.Delete(img); err != nil {
		var vd views.Data
		vd.Yeild = gallery
		vd.ErrorAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	url, err := g.r.Get(NamedGalleryEditRoute).URL("id", fmt.Sprintf("%v", gallery.ID))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/galleries", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
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
		log.Println(err)
		http.Redirect(w, r, "/galleries", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
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

// galleryByID gets gorilla mux url variables out and returns a gallery with that ID along with
// all of its images and no error.  If one does not exits with that id it will return nil and an error
func (g *Galleries) galleryByID(w http.ResponseWriter, r *http.Request) (*models.Gallery, error) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
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
		log.Println(err)
		http.Error(w, "Whoops! Something went wrong.", http.StatusInternalServerError)
		return nil, err
	}
	images, _ := g.is.ByGalleryID(gallery.ID)
	gallery.Images = images
	return gallery, nil
}
