package usr

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/Z3DRP/lessor-service/internal/auth"
	"github.com/Z3DRP/lessor-service/internal/cmerr"
	"github.com/Z3DRP/lessor-service/internal/dtos"
	"github.com/Z3DRP/lessor-service/internal/filters"
	"github.com/Z3DRP/lessor-service/internal/ztype"
	"github.com/Z3DRP/lessor-service/pkg/utils"
)

type UserHandler struct {
	UserService
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
	log.Printf("sign-in hit, %v", creds)

	authenticated, user, err := u.AuthenticateUser(r.Context(), creds)
	log.Printf("usr authenticated: %v", user)

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

	log.Println("generating token")
	token, err := auth.GenerateToken(user.Uid.String(), creds.Email, user.ProfileType)

	if err != nil {
		log.Printf("error generating token %v\n", err)
		utils.WriteErr(w, http.StatusInternalServerError, err)
		return
	}

	log.Printf("token generated: %v", token)

	// for session based
	// http.SetCookie(w, &http.Cookie{
	// 	Name:     "token",
	// 	Value:    token,
	// 	Expires:  auth.Expirey,
	// 	HttpOnly: true,
	// 	// Secure: true,
	// 	// SameSite: http.SameSiteStrictMode,
	// })

	res := ztype.JsonResponse{
		"accessToken": token,
		"user":        user,
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

		token, err := auth.GenerateToken(user.Uid.String(), user.Username, user.ProfileType)

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

		log.Printf("returning login res: %+v", res)

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			u.logger.MustDebug(fmt.Sprintf("json encoding err %v", err))
			log.Printf("error encoding json: %v", err)
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
