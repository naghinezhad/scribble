package web

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/nasermirzaei89/scribble/authentication"
	authcontext "github.com/nasermirzaei89/scribble/authentication/context"
)

func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := h.getSessionValue(r, sessionIDKey)
		if err != nil {
			if _, ok := errors.AsType[*SessionValueNotFoundError](err); !ok {
				slog.ErrorContext(
					r.Context(),
					"error on getting session value",
					"key",
					sessionIDKey,
					"error",
					err,
				)
				http.Error(w, "error on getting session value", http.StatusInternalServerError)

				return
			}
		}

		if sessionID != nil && sessionID.(string) != "" {
			session, err := h.authSvc.GetSession(r.Context(), sessionID.(string))
			if err != nil {
				if _, ok := errors.AsType[*authentication.SessionNotFoundError](err); ok {
					err = h.deleteSessionValue(w, r, sessionIDKey)
					if err != nil {
						slog.ErrorContext(
							r.Context(),
							"error on deleting session value",
							"key",
							sessionIDKey,
							"error",
							err,
						)
						http.Error(
							w,
							"error on deleting session value",
							http.StatusInternalServerError,
						)

						return
					}

					next.ServeHTTP(w, r)

					return
				}

				slog.ErrorContext(
					r.Context(),
					"error on getting session",
					"sessionId",
					sessionID,
					"error",
					err,
				)
				http.Error(w, "error on getting session", http.StatusInternalServerError)

				return
			}

			r = r.WithContext(authcontext.WithSessionID(r.Context(), session.ID))

			user, err := h.authSvc.GetUser(r.Context(), session.UserID)
			if err != nil {
				if _, ok := errors.AsType[*authentication.UserNotFoundError](err); ok {
					err = h.authSvc.Logout(r.Context(), session.ID)
					if err != nil {
						slog.ErrorContext(
							r.Context(),
							"error on logging out session",
							"sessionId",
							session.ID,
							"error",
							err,
						)
						http.Error(w, "error on logging out session", http.StatusInternalServerError)

						return
					}

					err = h.deleteSessionValue(w, r, sessionIDKey)
					if err != nil {
						slog.ErrorContext(
							r.Context(),
							"error on deleting session value",
							"key",
							sessionIDKey,
							"error",
							err,
						)
						http.Error(
							w,
							"error on deleting session value",
							http.StatusInternalServerError,
						)

						return
					}

					next.ServeHTTP(w, r)

					return
				}

				slog.ErrorContext(r.Context(), "error retrieving user", "error", err)
				http.Error(w, "error on retrieving user", http.StatusInternalServerError)

				return
			}

			r = r.WithContext(authcontext.WithSubject(r.Context(), user.ID))
		}

		next.ServeHTTP(w, r)
	})
}

func isAuthenticated(ctx context.Context) bool {
	return authcontext.GetSubject(ctx) != authcontext.Anonymous
}

func isAuthenticatedRequest(r *http.Request) bool {
	return isAuthenticated(r.Context())
}

func (h *Handler) AuthenticatedOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticatedRequest(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)

			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) GuestOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isAuthenticatedRequest(r) {
			http.Redirect(w, r, "/", http.StatusSeeOther)

			return
		}

		next.ServeHTTP(w, r)
	})
}
