package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/go_final_project/config"
)

type AuthHandler struct {
	Config *config.Сonfig
}

func NewAuthHandler(cfg *config.Сonfig) AuthHandler {
	return AuthHandler{Config: cfg}
}

type getAuthByPassword struct {
	Token string `json:"token"`
}

type errResponse struct {
	Error string `json:"error"`
}

type bodyRequest struct {
	Password string `json:"password"`
}

// GetAuthByPassword ... Получение токена
// @Summary Получение токена по паролю
// @Description Получение токена по паролю
// @Accept json
// @Param password body bodyRequest true "Пароль профиля"
// @Success 200 {object} getAuthByPassword
// @Failure 400,401,500 {object} errResponse
// @Router /api/signin [post]
func (a *AuthHandler) GetAuthByPassword(w http.ResponseWriter, r *http.Request) {
	var (
		buf  bytes.Buffer
		body bodyRequest
	)

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Errorf("http.GetAuthByPassword: %+v", err)

		errResp := errResponse{
			Error: err.Error(),
		}
		returnErr(http.StatusBadRequest, errResp, w)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &body); err != nil {
		log.Errorf("http.GetAuthByPassword: %+v", err)

		errResp := errResponse{
			Error: err.Error(),
		}
		returnErr(http.StatusBadRequest, errResp, w)
		return
	}

	if body.Password != a.Config.Password {
		err := fmt.Errorf("password mismatch")
		log.Errorf("http.GetAuthByPassword: %+v", err)

		errResp := errResponse{
			Error: err.Error(),
		}
		returnErr(http.StatusUnauthorized, errResp, w)
		return
	}

	jwtToken := jwt.New(jwt.SigningMethodHS256)
	signedToken, err := jwtToken.SignedString([]byte(body.Password))
	if err != nil {
		log.Errorf("http.GetAuthByPassword: %+v", err)

		errResp := errResponse{
			Error: err.Error(),
		}
		returnErr(http.StatusUnauthorized, errResp, w)
		return
	}

	authResp := getAuthByPassword{
		Token: signedToken,
	}

	resp, err := json.Marshal(authResp)
	if err != nil {
		log.Errorf("http.GetAuthByPassword: %+v", err)

		errResp := errResponse{
			Error: err.Error(),
		}
		returnErr(http.StatusInternalServerError, errResp, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Errorf("http.GetAuthByPassword: %+v", err)

		errResp := errResponse{
			Error: err.Error(),
		}
		returnErr(http.StatusInternalServerError, errResp, w)
	}
}

func returnErr(status int, message interface{}, w http.ResponseWriter) {
	messageJson, err := json.Marshal(message)
	if err != nil {
		status = http.StatusInternalServerError
		messageJson = []byte("{\"error\":\"" + err.Error() + "\"}")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(messageJson)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		log.Errorf("get wallet balance by UUID error: %+v", err)
	}
}
