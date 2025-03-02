package routes

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Z3DRP/lessor-service/config"
	"github.com/Z3DRP/lessor-service/internal/auth"
	"github.com/Z3DRP/lessor-service/internal/crane"
	"github.com/Z3DRP/lessor-service/internal/middlewares"
	"github.com/Z3DRP/lessor-service/internal/services/alssr"
	"github.com/Z3DRP/lessor-service/internal/services/property"
	"github.com/Z3DRP/lessor-service/internal/services/usr"
	"github.com/Z3DRP/lessor-service/pkg/utils"
	"github.com/golang-jwt/jwt/v4"
)

func NewServer(sconfig *config.ZServerConfig, alsrHndlr alssr.AlessorHandler, usrHndlr usr.UserHandler, prptyHandler property.PropertyHandler) (*http.Server, error) {

	mux := http.NewServeMux()
	registerRoutes(mux, alsrHndlr, usrHndlr, prptyHandler)

	mwChain := middlewares.MiddlewareChain(handlePanic, loggerMiddleware, headerMiddleware, contextMiddleware)
	server := &http.Server{
		Addr:         sconfig.Address,
		ReadTimeout:  time.Second * time.Duration(sconfig.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(sconfig.WriteTimeout),
		Handler:      mwChain(mux),
	}

	return server, nil
}

func registerRoutes(mux *http.ServeMux, aHndlr alssr.AlessorHandler, uHndlr usr.UserHandler, pHndlr property.PropertyHandler) {
	mux.HandleFunc("POST /sign-in", uHndlr.HandleLogin)
	mux.HandleFunc("POST /sign-up", uHndlr.HandleSignUp)
	mux.HandleFunc("GET /user-details", uHndlr.HandleGetDetails)
	mux.HandleFunc("GET /alessor", aHndlr.HandleGetAlessors)
	mux.HandleFunc("GET /alessor/{id}", aHndlr.HandleGetAlessor)
	mux.HandleFunc("POST /alessor/{id}", aHndlr.HandleCreateAlessor)
	mux.HandleFunc("PUT /alessor/{id}", aHndlr.HandleUpdateAlessor)
	mux.HandleFunc("DELETE /alessor/{id}", aHndlr.HandleDeleteAlessor)
	mux.HandleFunc("GET /user", uHndlr.HandleGetUsers) // admin or alessors route only
	mux.HandleFunc("GET /user/{id}", uHndlr.HandleGetUser)
	mux.HandleFunc("POST /user/{id}", uHndlr.HandleCreateUser)
	mux.HandleFunc("PUT /user/{id}", uHndlr.HandleUpdateUser)
	mux.HandleFunc("DELETE /user/{id}", uHndlr.HandleDeleteUser)
	// need to update this to be more restful properties needs to be property/alsrId but
	// but alessor id is needed and is triggering property/pid api doesnt
	// know if its alessor id or propertyId so for now just use /properties/id
	mux.HandleFunc("GET /properties/{id}", pHndlr.HandleGetProperties)
	mux.HandleFunc("GET /property", pHndlr.HandleGetProperty)
	mux.HandleFunc("GET /property/{id}", pHndlr.HandleGetProperty)
	mux.HandleFunc("POST /property", pHndlr.HandleCreateProperty)
	mux.HandleFunc("PUT /property/{id}", pHndlr.HandleUpdateProperty)
	//mux.HandleFunc("DELETE /property/{id}", pHndlr.HandleDeleteProperty)
}

// make this unexported after jwt in use
func Authenticate(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cook, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				utils.WriteErr(w, http.StatusUnauthorized, err)
				return
			}
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		token := cook.Value
		claims := &auth.UserClaims{}
		tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return auth.GetJwtKey()
		})

		if err != nil || !tkn.Valid {
			utils.WriteErr(w, http.StatusUnauthorized, errors.New("unathorized"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func headerMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		crane.DefaultLogger.MustDebug(fmt.Sprintf("orign: %v", origin))

		if config.IsValidOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		//w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func contextMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeout := 10 * time.Minute
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func loggerMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &middlewares.WrappedWriter{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}
		next.ServeHTTP(wrapped, r)
		crane.DefaultLogger.MustDebug(fmt.Sprintf("Method: %s, URI: %s, IP: %s, Duration: %v, Status: %v", r.Method, r.RequestURI, r.RemoteAddr, start, wrapped.StatusCode))
	})
}

func HandleShutdown(server *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	crane.DefaultLogger.MustDebug("Shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		crane.DefaultLogger.MustFatal(fmt.Sprintf("Server forced shutdown: %v", err))
	}
	crane.DefaultLogger.MustDebug("Server exited")
}

func handlePanic(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				crane.DefaultLogger.MustDebug(fmt.Sprintf("panic recovered %v", rec))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}
