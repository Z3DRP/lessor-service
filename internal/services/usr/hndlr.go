package usr

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Z3DRP/lessor-service/internal/auth"
	"github.com/Z3DRP/lessor-service/internal/cmerr"
	"github.com/Z3DRP/lessor-service/internal/dac"
	"github.com/Z3DRP/lessor-service/internal/dtos"
	"github.com/Z3DRP/lessor-service/internal/filters"
	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/Z3DRP/lessor-service/internal/services/alssr"
	"github.com/Z3DRP/lessor-service/internal/ztype"
	"github.com/Z3DRP/lessor-service/pkg/utils"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type UserExtendedService struct {
	UserService
	alssr.AlessorService
}

func (u *UserExtendedService) ServiceName() string {
	return "Extended User"
}

func (u UserExtendedService) ServiceNames() []string {
	names := []string{
		"user",
		"alessor",
	}
	return names
}

type UserHandler struct {
	UserService
	alssr.AlessorService
	//UserExtendedService
}

func NewHandler(service UserService) UserHandler {
	return UserHandler{UserService: service}
}

func (u UserHandler) HandlerName() string {
	return "User"
}

func (u UserHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var creds filters.Creds
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		utils.WriteErr(w, http.StatusBadRequest, err)
		return
	}

	authenticated, user, err := u.AuthenticateUser(r.Context(), creds)

	if err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err)
		log.Printf("auth err %v", err)
		return
	}

	if !authenticated {
		log.Printf("usr not authenticated")
		utils.WriteErr(w, http.StatusUnauthorized, errors.New("invalid credentials"))
		return
	}

	token, err := auth.GenerateToken(user.Uid.String(), creds.Email, user.ProfileType)

	if err != nil {
		log.Printf("error generating token %v\n", err)
		utils.WriteErr(w, http.StatusInternalServerError, err)
		return
	}

	// for session based
	// http.SetCookie(w, &http.Cookie{
	// 	Name:     "token",
	// 	Value:    token,
	// 	Expires:  auth.Expirey,
	// 	HttpOnly: true,
	// 	// Secure: true,
	// 	// SameSite: http.SameSiteStrictMode,
	// })

	uDto := dtos.NewSigninResponse(&user)
	if user.ProfileType == "worker" {
		var workerLessorId uuid.UUID
		workerLessorId, err = u.GetWorkerLessor(r.Context(), &user)
		if err != nil {
			log.Printf("error in handler for worker data fetch %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}
		uDto.LessorId = workerLessorId
	}

	res := ztype.JsonResponse{
		"accessToken": token,
		"user":        uDto,
	}

	if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
		u.logger.MustDebug(err.Error())
		utils.WriteErr(w, http.StatusInternalServerError, err)
		return
	}
}

func (u UserHandler) HandleSignUp(w http.ResponseWriter, r *http.Request) {
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		u.logger.MustDebug(timeoutErr.Error())
		log.Println("request timeout")
		utils.WriteErr(w, http.StatusRequestTimeout, &timeoutErr)
	default:
		// TODO change defalting to text communication preference
		var payload dtos.UserSignupRequest
		w.Header().Set("Content-Type", "application/json")

		if err := utils.ParseJSON(r, &payload); err != nil {
			u.logger.MustDebug("failed to parse request body")
			utils.WriteErr(w, http.StatusBadRequest, err)
			log.Println(err)
			return
		}

		if err := payload.Validate(); err != nil {
			u.logger.MustDebug(fmt.Sprintf("user create payload failed validation, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			log.Println(err)
			return
		}

		user, err := u.CreateUsr(r.Context(), payload)

		if err != nil {
			u.logger.MustDebug(fmt.Sprintf("database err %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			log.Println(err)
			return
		}

		// create a initial alessor profile for the user
		_, err = u.CreateAlessor(r.Context(), user)

		if err != nil {
			log.Printf("failed to create alessor profile %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		token, err := auth.GenerateToken(user.Uid.String(), user.Email, user.ProfileType)

		if err != nil {
			u.logger.MustDebug(fmt.Sprintf("auth error: %v", err))
			utils.WriteErr(w, http.StatusInternalServerError, err)
			log.Println("auth err")
			return
		}

		usrDto := dtos.NewSigninResponse(user)

		res := ztype.JsonResponse{
			"accessToken": token,
			"user":        usrDto,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			u.logger.MustDebug(fmt.Sprintf("json encoding err %v", err))
			log.Printf("error encoding json: %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (u UserHandler) HandleSignUpWorker(w http.ResponseWriter, r *http.Request) {
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		u.logger.MustDebug(timeoutErr.Error())
		log.Println("request timeout")
		utils.WriteErr(w, http.StatusRequestTimeout, &timeoutErr)
	default:
		// TODO change defalting to text communication preference
		log.Println("worker signup ep")
		var payload dtos.WorkerUserSignupRequest
		w.Header().Set("Content-Type", "application/json")

		if err := utils.ParseJSON(r, &payload); err != nil {
			u.logger.MustDebug("failed to parse request body")
			utils.WriteErr(w, http.StatusBadRequest, err)
			log.Println(err)
			return
		}

		if err := payload.Validate(); err != nil {
			u.logger.MustDebug(fmt.Sprintf("user create payload failed validation, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			log.Println(err)
			return
		}

		signupRequest := dtos.UserSignupRequest{
			FirstName:   payload.FirstName,
			LastName:    payload.LastName,
			ProfileType: payload.ProfileType,
			Username:    payload.Username,
			Password:    payload.Password,
			Phone:       payload.Phone,
			Email:       payload.Email,
		}

		user, err := u.CreateUsr(r.Context(), signupRequest)

		if err != nil {
			u.logger.MustDebug(fmt.Sprintf("database err %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			log.Println(err)
			return
		}

		// create a initial alessor profile for the user
		log.Println("")
		log.Println("")
		log.Printf("payload %+v", payload)
		_, err = u.CreateWorker(r.Context(), user, payload)

		if err != nil {
			log.Printf("failed to create worker profile %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		token, err := auth.GenerateToken(user.Uid.String(), user.Email, user.ProfileType)
		log.Printf("Generating token for UID: %s, Email: %s, Role: %s", user.Uid.String(), user.Email, user.ProfileType)

		if err != nil {
			u.logger.MustDebug(fmt.Sprintf("auth error: %v", err))
			utils.WriteErr(w, http.StatusInternalServerError, err)
			log.Println("auth err")
			return
		}

		usrDto := dtos.NewWorkerSignUpResponse(user, utils.ParseUuid(payload.LessorId))

		res := ztype.JsonResponse{
			"accessToken": token,
			"user":        usrDto,
		}

		log.Printf("returning login res: %+v", res)

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			u.logger.MustDebug(fmt.Sprintf("json encoding err %v", err))
			log.Printf("error encoding json: %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (u UserHandler) HandleGetDetails(w http.ResponseWriter, r *http.Request) {
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		u.logger.MustDebug(timeoutErr.Error())
		utils.WriteErr(w, http.StatusRequestTimeout, &timeoutErr)
	default:
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			u.logger.LogFields(logrus.Fields{
				"msg":     "request did not have auth token",
				"request": r.URL,
			})
			log.Printf("request did not have auth token: %v", r.URL)
			utils.WriteErr(w, http.StatusBadRequest, errors.New("request did not have auth token"))
			return
		}

		tokenParts := strings.Split(authHeader, " ")

		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			log.Printf("token was not on header")
			utils.WriteErr(w, http.StatusBadRequest, errors.New("header did not have token"))
			return
		}

		token := tokenParts[1]
		user, err := u.ValidateClaims(r.Context(), token)

		if err != nil {
			// TODO will need to check for expirey after is setup
			u.logger.LogFields(logrus.Fields{
				"msg": "token claim validation failed",
				"err": err,
			})
			log.Printf("token claim validation failed: %v", err)
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		uDto := dtos.NewSigninRequest(user)
		res := ztype.JsonResponse{
			"user": uDto,
		}

		if err := utils.WriteJSON(w, http.StatusOK, res); err != nil {
			log.Printf("failed to write json response %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (u UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		u.logger.MustDebug(timeoutErr.Error())
		utils.WriteErr(w, http.StatusRequestTimeout, &timeoutErr)
	default:
		var payload dtos.UserSignupRequest
		w.Header().Set("Content-Type", "application/json")

		if err := utils.ParseJSON(r, payload); err != nil {
			u.logger.MustDebug(fmt.Sprintf("failed to parse request body, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		if err := payload.Validate(); err != nil {
			u.logger.MustDebug(fmt.Sprintf("user create payload failed validation, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		user, err := u.CreateUsr(r.Context(), payload)
		if err != nil {
			u.logger.MustDebug(fmt.Sprintf("database err %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		res := ztype.JsonResponse{
			"user":    user,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			u.logger.MustDebug(err.Error())
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (u UserHandler) HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		u.logger.MustDebug(timeoutErr.Error())
		utils.WriteErr(w, http.StatusRequestTimeout, &timeoutErr)
	default:
		fltr, err := filters.GenFilter(r)
		if err != nil {
			u.logger.MustDebug(fmt.Sprintf("failed create user filter, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		usrs, err := u.GetUsrs(r.Context(), fltr)
		if err != nil {
			var noResults dac.ErrNoResults
			if errors.As(err, &noResults) {
				res := ztype.JsonResponse{
					"users":   make([]model.User, 0),
					"success": true,
				}
				if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
					utils.WriteErr(w, http.StatusInternalServerError, err)
					return
				}
			}
			u.logger.MustDebug(fmt.Sprintf("database err, %v", err))
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"users":   usrs,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (u UserHandler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		u.logger.MustDebug(timeoutErr.Error())
		utils.WriteErr(w, http.StatusRequestTimeout, &timeoutErr)
	default:
		fltr, err := filters.GenFilter(r)
		if err != nil {
			u.logger.MustDebug(fmt.Sprintf("failed to create user filter, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		usr, err := u.GetUsr(r.Context(), fltr)
		if err != nil {
			var noResults dac.ErrNoResults
			if errors.As(err, &noResults) {
				res := ztype.JsonResponse{
					"user":    nil,
					"success": true,
				}

				if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
					utils.WriteErr(w, http.StatusInternalServerError, err)
				}
				return
			}
			u.logger.MustDebug(fmt.Sprintf("database errr, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		res := ztype.JsonResponse{
			"user":    usr,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			u.logger.MustDebug(fmt.Sprintf("internal server error, %v", err))
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

	}
}

func (u UserHandler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		u.logger.MustDebug(timeoutErr.Error())
		utils.WriteErr(w, http.StatusRequestTimeout, &timeoutErr)
	default:
		var payload dtos.UserRequest
		if err := utils.ParseJSON(r, &payload); err != nil {
			u.logger.MustDebug(fmt.Sprintf("failed to parse user dto from request, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		if err := payload.Validate(); err != nil {
			u.logger.MustDebug(fmt.Sprintf("user dto failed validation, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		updateUsr, err := u.ModifyUser(r.Context(), payload)
		if err != nil {
			u.logger.MustDebug(fmt.Sprintf("database err, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		res := ztype.JsonResponse{
			"user":    updateUsr,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			u.logger.MustDebug(err.Error())
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (u UserHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		u.logger.MustDebug(timeoutErr.Error())
		utils.WriteErr(w, http.StatusRequestTimeout, &timeoutErr)
	default:
		delReq, err := dtos.BuildDeleteRequest(r)
		if err != nil {
			u.logger.MustDebug(fmt.Sprintf("invalid delete request, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		if err = delReq.Validate(); err != nil {
			u.logger.MustDebug(fmt.Sprintf("invalid delete request, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, fmt.Errorf("invalid delete request, %v", err))
			return
		}

		req, ok := delReq.(dtos.DeleteRequest)
		if !ok {
			invalidData := cmerr.ErrUnexpectedData{Wanted: dtos.DeleteRequest{}, Got: delReq}
			u.logger.MustDebug(fmt.Sprintf("invalid delete request, %v", invalidData))
			utils.WriteErr(w, http.StatusBadRequest, invalidData)
			return
		}

		if err = u.DeleteUsr(r.Context(), req); err != nil {
			u.logger.MustDebug(fmt.Sprintf("database err, %v", err))
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			u.logger.MustDebug(err.Error())
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}
