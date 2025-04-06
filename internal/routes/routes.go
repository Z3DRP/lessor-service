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
	rentalproperty "github.com/Z3DRP/lessor-service/internal/services/rentalProperty"
	"github.com/Z3DRP/lessor-service/internal/services/task"
	"github.com/Z3DRP/lessor-service/internal/services/usr"
	"github.com/Z3DRP/lessor-service/internal/services/worker"
	"github.com/Z3DRP/lessor-service/pkg/utils"
	"github.com/golang-jwt/jwt/v4"
)

func NewServer(
	sconfig *config.ZServerConfig,
	alessorHndlr alssr.AlessorHandler,
	usrHndlr usr.UserHandler,
	propertyHndlr property.PropertyHandler,
	taskHndlr task.TaskHandler,
	rentalPropertyHndlr rentalproperty.RentalPropertyHandler,
	workerHndlr worker.WorkerHandler,
) (*http.Server, error) {

	mux := http.NewServeMux()
	registerRoutes(mux, alessorHndlr, usrHndlr, propertyHndlr, taskHndlr, rentalPropertyHndlr, workerHndlr)

	mwChain := middlewares.MiddlewareChain(handlePanic, loggerMiddleware, headerMiddleware, contextMiddleware)
	server := &http.Server{
		Addr:         sconfig.Address,
		ReadTimeout:  time.Second * time.Duration(sconfig.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(sconfig.WriteTimeout),
		Handler:      mwChain(mux),
	}

	return server, nil
}

func registerRoutes(
	mux *http.ServeMux,
	aHandler alssr.AlessorHandler,
	uHandler usr.UserHandler,
	pHandler property.PropertyHandler,
	tHandler task.TaskHandler,
	rpHandler rentalproperty.RentalPropertyHandler,
	wHandler worker.WorkerHandler,
) {
	mux.HandleFunc("POST /sign-in", uHandler.HandleLogin)
	mux.HandleFunc("POST /sign-up", uHandler.HandleSignUp)
	mux.HandleFunc("POST /sign-up/worker", uHandler.HandleSignUpWorker)

	mux.HandleFunc("GET /alessor", aHandler.HandleGetAlessors)
	mux.HandleFunc("GET /alessor/{id}", aHandler.HandleGetAlessor)
	mux.HandleFunc("POST /alessor/{id}", aHandler.HandleCreateAlessor)
	mux.HandleFunc("PUT /alessor/{id}", aHandler.HandleUpdateAlessor)
	mux.HandleFunc("DELETE /alessor/{id}", aHandler.HandleDeleteAlessor)
	mux.HandleFunc("GET /alessor/{id}/task", tHandler.HandleGetTasks)
	mux.HandleFunc("GET /alessor/{id}/worker", wHandler.HandleGetWorkers)
	// need to add this and remove from below and change to property
	//mux.HandleFunc("GET alessor/{id}/property", pHandler.HandleGetProperties)

	mux.HandleFunc("GET /user", uHandler.HandleGetUsers) // admin or alessors route only
	mux.HandleFunc("GET /user/{id}", uHandler.HandleGetUser)
	mux.HandleFunc("POST /user/{id}", uHandler.HandleCreateUser)
	mux.HandleFunc("PUT /user/{id}", uHandler.HandleUpdateUser)
	mux.HandleFunc("DELETE /user/{id}", uHandler.HandleDeleteUser)
	mux.HandleFunc("GET /user-details", uHandler.HandleGetDetails)

	// need to update this to be more restful properties needs to be property/alsrId but
	// but alessor id is needed and is triggering property/pid api doesnt
	// know if its alessor id or propertyId so for now just use /properties/id
	mux.HandleFunc("GET /property", pHandler.HandleGetProperty)
	mux.HandleFunc("GET /properties/{id}", pHandler.HandleGetProperties)
	mux.HandleFunc("GET /property/{id}", pHandler.HandleGetProperty)
	mux.HandleFunc("POST /property", pHandler.HandleCreateProperty)
	mux.HandleFunc("PUT /property/{id}", pHandler.HandleUpdateProperty)
	mux.HandleFunc("DELETE /property/{id}", pHandler.HandleDeleteProperty)

	mux.HandleFunc("GET /task/{id}", tHandler.HandleGetTask)
	mux.HandleFunc("POST /task", tHandler.HandleCreateTask)
	mux.HandleFunc("PUT /task/{id}", tHandler.HandleUpdateTask)
	mux.HandleFunc("DELETE /task/{id}", tHandler.HandleDeleteTask)
	mux.HandleFunc("PUT /task/{id}/priority", tHandler.HandleUpdatePriority)
	mux.HandleFunc("PUT /task/{id}/assign", tHandler.HandleAssignTask)
	mux.HandleFunc("PUT /task/{id}/complete", tHandler.HandleCompleteTask)
	mux.HandleFunc("PUT /task/{id}/pause", tHandler.HandlePauseTask)
	mux.HandleFunc("PUT /task/{id}/unpause", tHandler.HandleUnPauseTask)

	mux.HandleFunc("GET /rental-property", rpHandler.HandleGetRentalProperties)
	mux.HandleFunc("GET /rental-property/{id}", rpHandler.HandleDeleteRentalProperty)
	mux.HandleFunc("POST /rental", rpHandler.HandleCreateRentalProperty)
	mux.HandleFunc("PUT /rental/{id}", rpHandler.HandleUpdateRentalProperty)
	mux.HandleFunc("DELETE /rental/{id}", rpHandler.HandleDeleteRentalProperty)

	mux.HandleFunc("POST /worker", wHandler.HandleCreateWorker)
	mux.HandleFunc("GET /worker/{id}", wHandler.HandleGetWorker)
	mux.HandleFunc("PUT /worker/{id}", wHandler.HandleUpdateWorker)
	mux.HandleFunc("DELETE /worker/{id}", wHandler.HandleDeleteWorker)
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
		// log.Printf("🔍 Incoming request: %s %s", r.Method, r.URL.Path)
		// log.Printf("Headers: %v", r.Header)

		if config.IsValidOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
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
