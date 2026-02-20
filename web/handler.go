package web

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"maps"
	"net/http"
	"runtime/debug"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/nasermirzaei89/scribble/auth"
	authcontext "github.com/nasermirzaei89/scribble/auth/context"
	"github.com/nasermirzaei89/scribble/contents"
)

var (
	//go:embed templates/*
	templatesFS embed.FS

	//go:embed static/*
	staticFS embed.FS
)

const defaultSiteTitle = "Scribble"

type Handler struct {
	mux         *http.ServeMux
	handler     http.Handler
	tpl         *template.Template
	static      fs.FS
	authSvc     *auth.Service
	contentsSvc *contents.Service
	cookieStore *sessions.CookieStore
	sessionName string
	assetHashes map[string]string
}

var _ http.Handler = (*Handler)(nil)

func NewHandler(
	authSvc *auth.Service,
	contentsSvc *contents.Service,
	cookieStore *sessions.CookieStore,
	sessionName string,
	csrfAuthKeys []byte,
	csrfTrustedOrigins []string,
) (*Handler, error) {
	h := &Handler{
		mux:         nil,
		handler:     nil,
		tpl:         nil,
		authSvc:     authSvc,
		contentsSvc: contentsSvc,
		cookieStore: cookieStore,
		sessionName: sessionName,
		assetHashes: make(map[string]string),
	}

	{
		tpl, err := template.New("").Funcs(h.funcs()).ParseFS(templatesFS, "templates/*.gohtml")
		if err != nil {
			return nil, fmt.Errorf("failed to parse templates: %w", err)
		}

		h.tpl = tpl
	}

	{
		static, err := fs.Sub(staticFS, "static")
		if err != nil {
			return nil, fmt.Errorf("failed to sub static fs: %w", err)
		}

		h.static = static
	}

	{
		h.mux = &http.ServeMux{}
		h.handler = h.mux

		h.registerRoutes()
	}

	{
		h.handler = h.authMiddleware(h.handler)

		{
			csrfMiddleware := csrf.Protect(
				csrfAuthKeys,
				csrf.TrustedOrigins(csrfTrustedOrigins),
			)

			h.handler = csrfMiddleware(h.handler)
		}

		h.handler = recoverMiddleware(h.handler)
	}

	return h, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}

func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("/", h.HandleIndex)

	h.mux.Handle("GET /register", h.HandleRegisterPage())
	h.mux.Handle("POST /register", h.HandleRegister())
	h.mux.Handle("GET /login", h.HandleLoginPage())
	h.mux.Handle("POST /login", h.HandleLogin())
	h.mux.Handle("GET /logout", h.HandleLogoutPage())
	h.mux.Handle("POST /logout", h.HandleLogout())

	h.mux.Handle("GET /create-post", h.HandleCreatePostPage())
	h.mux.Handle("POST /create-post", h.HandleCreatePost())
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(ctx context.Context) {
			if err := recover(); err != nil {
				slog.ErrorContext(
					ctx,
					"recovered from panic",
					"error",
					err,
					"stack",
					string(debug.Stack()),
				)

				http.Error(w, "internal error occurred", http.StatusInternalServerError)
			}
		}(r.Context())

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) renderTemplate(w http.ResponseWriter, r *http.Request, name string, extraData map[string]any,
) {
	data := map[string]any{
		"CurrentPath":     r.URL.Path,
		"Lang":            "en",
		"Dir":             "ltr",
		"IsAuthenticated": isAuthenticated(r),
	}

	maps.Copy(data, extraData)

	data["SiteTitle"] = defaultSiteTitle

	if extraData["SiteTitle"] != nil {
		data["SiteTitle"] = fmt.Sprintf("%s | %s", extraData["SiteTitle"], data["SiteTitle"])
	}

	err := h.tpl.ExecuteTemplate(w, name, data)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to render template", "name", name, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		return
	}
}

func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		h.HandleHomePage(w, r)

		return
	}

	h.HandleStatic(w, r)
}

// HandleStatic serves static files.
func (h *Handler) HandleStatic(w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Cache-Control", "public, max-age=3600")
	http.FileServer(http.FS(h.static)).ServeHTTP(w, r)
}

func (h *Handler) HandleHomePage(w http.ResponseWriter, r *http.Request) {
	posts, err := h.contentsSvc.ListPosts(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list posts", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		return
	}

	postsWithAuthors, err := h.preloadPostAuthor(r.Context(), posts)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to preload post authors", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		return
	}

	data := map[string]any{
		"Posts": postsWithAuthors,
	}

	h.renderTemplate(w, r, "home-page.gohtml", data)
}

type PostWithAuthor struct {
	contents.Post

	Author *auth.User
}

func (h *Handler) preloadPostAuthor(ctx context.Context, posts []*contents.Post) ([]*PostWithAuthor, error) {
	var result []*PostWithAuthor

	// TODO: optimize this by batching user retrieval instead of doing it one by one
	for _, post := range posts {
		author, err := h.authSvc.GetUser(ctx, post.AuthorID)
		if err != nil {
			return nil, fmt.Errorf("failed to get author: %w", err)
		}

		result = append(result, &PostWithAuthor{
			Post:   *post,
			Author: author,
		})
	}

	return result, nil
}

func (h *Handler) HandleRegisterPage() http.Handler {
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			csrf.TemplateTag: csrf.TemplateField(r),
			"SiteTitle":      "Register",
		}

		h.renderTemplate(w, r, "register-page.gohtml", data)
	})

	return h.GuestOnly(hf)
}

func (h *Handler) HandleRegister() http.Handler {
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to parse form", "error", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)

			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		err = h.authSvc.Register(r.Context(), username, password)
		if err != nil {
			var userAlreadyExistsErr *auth.UserAlreadyExistsError
			switch {
			case errors.As(err, &userAlreadyExistsErr):
				http.Error(w, "Username already exists", http.StatusConflict)
			default:
				slog.ErrorContext(r.Context(), "failed to register user", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}

			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	return h.GuestOnly(hf)
}

func (h *Handler) HandleLoginPage() http.Handler {
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			csrf.TemplateTag: csrf.TemplateField(r),
			"SiteTitle":      "Login",
		}

		h.renderTemplate(w, r, "login-page.gohtml", data)
	})

	return h.GuestOnly(hf)
}

func (h *Handler) HandleLogin() http.Handler {
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to parse form", "error", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)

			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		session, err := h.authSvc.Login(r.Context(), username, password)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrInvalidCredentials):
				http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			default:
				slog.ErrorContext(r.Context(), "failed to login user", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}

			return
		}

		err = h.setSessionValue(w, r, sessionIDKey, session.ID)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to set session ID", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)

			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	return h.GuestOnly(hf)
}

func (h *Handler) HandleLogoutPage() http.Handler {
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			csrf.TemplateTag: csrf.TemplateField(r),
			"SiteTitle":      "Logout",
		}

		h.renderTemplate(w, r, "logout-page.gohtml", data)
	})

	return h.AuthenticatedOnly(hf)
}

func (h *Handler) HandleLogout() http.Handler {
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID, ok := authcontext.SessionIDFromContext(r.Context())
		if ok {
			err := h.authSvc.Logout(r.Context(), sessionID)
			if err != nil {
				slog.ErrorContext(r.Context(), "error on logout", "sessionId", sessionID, "error", err)
				http.Error(w, "error on logout", http.StatusInternalServerError)

				return
			}
		}

		err := h.deleteSessionValue(w, r, sessionIDKey)
		if err != nil {
			slog.ErrorContext(
				r.Context(),
				"error on deleting session value",
				"key",
				sessionIDKey,
				"error",
				err,
			)
			http.Error(w, "error on deleting session value", http.StatusInternalServerError)

			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	return h.AuthenticatedOnly(hf)
}

func (h *Handler) HandleCreatePostPage() http.Handler {
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			csrf.TemplateTag: csrf.TemplateField(r),
			"SiteTitle":      "Create Post",
		}

		h.renderTemplate(w, r, "create-post-page.gohtml", data)
	})

	return h.AuthenticatedOnly(hf)
}

func (h *Handler) HandleCreatePost() http.Handler {
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to parse form", "error", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)

			return
		}

		content := r.FormValue("content")

		currentUser, err := h.authSvc.GetCurrentUser(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to get current user", "error", err)
			http.Error(w, "Failed to get current user", http.StatusInternalServerError)

			return
		}

		_, err = h.contentsSvc.CreatePost(r.Context(), contents.CreatePostRequest{
			AuthorID: currentUser.ID,
			Content:  content,
		})
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to create post", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)

			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	return h.AuthenticatedOnly(hf)
}
