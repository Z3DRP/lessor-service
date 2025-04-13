package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"unicode/utf8"

	"github.com/Z3DRP/lessor-service/internal/api"
	"github.com/google/uuid"
)

var (
	EmlRgx             = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	PhneRgx            = regexp.MustCompile(`^\d{3}-\d{3}-\d{4}$/`)
	DefaultRecordLimit = 10
	maxSize            = int64(1024000)
)

func WriteTimeoutResponse(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusRequestTimeout)

	errMsg := map[string]interface{}{
		"message": "request timeout",
		"status":  fmt.Sprintf("%v", http.StatusRequestTimeout),
		"success": false,
	}

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(errMsg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

func ParseJSON(r *http.Request, payload any) error {
	log.Printf("parsing body: %v", r.Body)
	defer r.Body.Close()
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}

	return json.NewDecoder(r.Body).Decode(payload)
}

func WriteJSON(w http.ResponseWriter, status int, msg any) error {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(msg)
}

func WriteErr(w http.ResponseWriter, status int, err error) {
	err = WriteJSON(w, status, map[string]string{"error": err.Error()})
	if err != nil {
		http.Error(w, err.Error(), status)
	}
}

func FormatErrMsg(msg string, err error) string {
	return fmt.Sprintf(msg, err)
}

func IsValidEmail(email string) bool {
	return EmlRgx.MatchString(email)
}

func IsValidPhone(phne string) bool {
	return PhneRgx.MatchString(phne)
}

func CharCount(str string) int {
	return utf8.RuneCountInString(str)
}

func ParseUuid(str string) uuid.UUID {
	if str == "" {
		log.Println("uuuid is empty")
		return uuid.Nil
	}

	uid, err := uuid.Parse(str)
	if err != nil {
		log.Println("failed to parse uuid setting to nil")
		return uuid.Nil
	}

	return uid
}

func DeterminRecordLimit(limt int) int {
	if limt <= 0 {
		return DefaultRecordLimit
	}

	return limt
}

func ParseFile(r *http.Request) (multipart.File, *multipart.FileHeader, error) {
	err := r.ParseMultipartForm(maxSize)

	if err != nil {
		return nil, nil, api.ErrMaxSize{Err: err}
	}

	file, header, err := r.FormFile("image")

	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			return nil, nil, err
		} else {
			return nil, nil, nil
		}
	}

	return file, header, nil
}

func ParseFloatOrZero(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return -1, err
	}

	return f, nil
}

func ParseIntOrZero(s string) (int, error) {
	if s == "" {
		return 0, nil
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return -1, err
	}

	return i, nil
}

func ParseBool(s string) bool {
	return s == "true" || s == "on"
}
