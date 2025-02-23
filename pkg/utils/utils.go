package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"unicode/utf8"

	"github.com/google/uuid"
)

var (
	EmlRgx             = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	PhneRgx            = regexp.MustCompile(`^\d{3}-\d{3}-\d{4}$/`)
	DefaultRecordLimit = 10
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
		return uuid.Nil
	}

	uid, err := uuid.Parse(str)
	if err != nil {
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
