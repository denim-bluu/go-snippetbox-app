package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"snippetbox.joonkang.net/internal/models"
	"snippetbox.joonkang.net/internal/validator"
)

type snippetCreateForm struct {
	Title               string `schema:"title,required"`
	Content             string `schema:"content,required"`
	Expires             int    `schema:"expires,required"`
	validator.Validator `schema:"-"`
}

type userSignupForm struct {
	Name                string `schema:"name,required"`
	Email               string `schema:"email,required"`
	Password            string `schema:"password,required"`
	validator.Validator `schema:"-"`
}
type userLoginForm struct {
	Email               string `schema:"email,required"`
	Password            string `schema:"password,required"`
	validator.Validator `schema:"-"`
}

func (app *application) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	snippets, err := app.snippetModel.Latest()
	if err != nil {
		app.serverError(w, r, err)
	}
	data := app.newTemplateData(w, r)
	data.Snippets = snippets

	session, err := app.cookieStore.Get(r, "session")
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	fm := session.Flashes("create-message")
	if fm != nil {
		data.Flash = fm[0].(string)
		session.Save(r, w)
	}

	app.render(w, r, http.StatusOK, "home.html", data)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	snippet, err := app.snippetModel.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := app.newTemplateData(w, r)
	data.Snippet = snippet

	session, err := app.cookieStore.Get(r, "session")
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	fm := session.Flashes("create-message")
	if fm != nil {
		data.Flash = fm[0].(string)
		session.Save(r, w)
	}

	app.render(w, r, http.StatusOK, "view.html", data)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(w, r)
	data.Form = snippetCreateForm{
		Expires: 365,
	}
	app.render(w, r, http.StatusOK, "create.html", data)
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	var form snippetCreateForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.Check(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.Check(validator.MaxStringLength(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.Check(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.Check(validator.PermitteValues(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !form.Valid() {
		data := app.newTemplateData(w, r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "create.html", data)
		return
	}

	id, err := app.snippetModel.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	session, err := app.cookieStore.Get(r, "session")
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	session.AddFlash("Snippet successfully created!", "create-message")
	session.Save(r, w)
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

func (app *application) snippetDelete(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(w, r)
	ids, err := app.snippetModel.GetIDs()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data.IDs = ids
	app.render(w, r, http.StatusOK, "delete.html", data)
}

func (app *application) snippetDeletePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(r.PostForm.Get("id"))
	if err != nil {
		app.notFound(w)
		return
	}

	err = app.snippetModel.Delete(id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(w, r)
	data.Form = userSignupForm{}
	app.render(w, r, http.StatusOK, "signup.html", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.Check(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.Check(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.Check(validator.StrPattenMatch(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.Check(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.Check(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(w, r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	err = app.userModel.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Address is already in use")
			data := app.newTemplateData(w, r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
			return
		} else {
			app.serverError(w, r, err)
		}
	}

	session, err := app.cookieStore.Get(r, "session")
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	session.AddFlash("User signup complete!", "create-message")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(w, r)
	data.Form = userLoginForm{}
	app.render(w, r, http.StatusOK, "login.html", data)
}
func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.Check(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.Check(validator.StrPattenMatch(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.Check(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(w, r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	id, err := app.userModel.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or Password is incorrect")
			data := app.newTemplateData(w, r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
			return
		} else {
			app.serverError(w, r, err)
			return
		}
	}

	session, err := app.cookieStore.Get(r, "session")
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	session.AddFlash("User login complete!", "create-message")
	session.Values["authenticatedUserID"] = id
	if err = sessions.Save(r, w); err != nil {
		app.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	session, err := app.cookieStore.Get(r, "session")
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	delete(session.Values, "authenticatedUserID")
	session.AddFlash("User logout complete!", "create-message")
	if err = sessions.Save(r, w); err != nil {
		app.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
