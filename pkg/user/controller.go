package user

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/thyyl/chatr/pkg/common"
)

func (s *HttpServer) CreateLocalUser(context *gin.Context) {
	var createLocalUserRequest CreateLocalUserRequest
	if err := context.ShouldBindJSON(&createLocalUserRequest); err != nil {
		common.Response(context, http.StatusBadRequest, common.ErrorInvalidParam)
		return
	}

	user, err := s.userService.CreateUser(context.Request.Context(), &User{
		Name:     createLocalUserRequest.Name,
		AuthType: LocalAuth,
	})
	if err != nil {
		s.logger.Error(err.Error())
		common.Response(context, http.StatusInternalServerError, common.ErrorServer)
		return
	}

	session, err := s.userService.SetUserSession(context.Request.Context(), user.Id)
	if err != nil {
		s.logger.Error(err.Error())
		common.Response(context, http.StatusInternalServerError, common.ErrorServer)
		return
	}

	common.SetAuthCookie(context, session, s.authCookieConfig.MaxAge, s.authCookieConfig.Path, s.authCookieConfig.Domain)
	context.JSON(http.StatusCreated, &UserDto{
		Id:   strconv.FormatUint(user.Id, 10),
		Name: user.Name,
	})
}

func (s *HttpServer) GetUser(context *gin.Context) {
	_, ok := context.Request.Context().Value(common.UserKey).(uint64)
	if !ok {
		common.Response(context, http.StatusUnauthorized, common.ErrorUnauthorized)
		return
	}

	var getUserRequest GetUserRequest
	if err := context.ShouldBindQuery(&getUserRequest); err != nil {
		common.Response(context, http.StatusBadRequest, common.ErrorInvalidParam)
		return
	}

	userId, err := strconv.ParseUint(getUserRequest.Id, 10, 64)
	if err != nil {
		common.Response(context, http.StatusBadRequest, common.ErrorInvalidParam)
		return
	}

	user, err := s.userService.GetUserById(context.Request.Context(), userId)
	if err != nil {
		if errors.Is(err, common.ErrorUserNotFound) {
			common.Response(context, http.StatusNotFound, common.ErrorUserNotFound)
			return
		}

		s.logger.Error(err.Error())
		common.Response(context, http.StatusInternalServerError, common.ErrorServer)
		return
	}

	context.JSON(http.StatusOK, &UserDto{
		Id:    strconv.FormatUint(user.Id, 10),
		Name:  user.Name,
		Photo: user.Photo,
	})
}

func (s *HttpServer) GetUserMe(context *gin.Context) {
	userId, ok := context.Request.Context().Value(common.UserKey).(uint64)
	if !ok {
		common.Response(context, http.StatusUnauthorized, common.ErrorUnauthorized)
		return
	}

	user, err := s.userService.GetUserById(context.Request.Context(), userId)
	if err != nil {
		if errors.Is(err, common.ErrorUserNotFound) {
			common.Response(context, http.StatusNotFound, common.ErrorUserNotFound)
			return
		}

		s.logger.Error(err.Error())
		common.Response(context, http.StatusInternalServerError, common.ErrorServer)
		return
	}

	context.JSON(http.StatusOK, &UserDto{
		Id:    strconv.FormatUint(user.Id, 10),
		Name:  user.Name,
		Photo: user.Photo,
	})
}

func (s *HttpServer) OAuthGoogleLogin(context *gin.Context) {
	state, err := common.GenerateStateOauthCookie(context, s.oAuthCookieConfig.MaxAge, s.oAuthCookieConfig.Path, s.oAuthCookieConfig.Domain)
	if err != nil {
		s.logger.Error(err.Error())
		common.Response(context, http.StatusInternalServerError, common.ErrorServer)
		return
	}

	url := s.googleOAuthConfig.AuthCodeURL(state)
	context.Redirect(http.StatusTemporaryRedirect, url)
}

func (s *HttpServer) OAuthGoogleCallback(context *gin.Context) {
	oauthState, err := common.GetCookie(context, common.OAuthStateCookieName)
	if err != nil {
		s.logger.Error(err.Error())
		context.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}
	if context.Query("state") != oauthState {
		s.logger.Error("invalid oauth google state")
		context.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	token, err := s.googleOAuthConfig.Exchange(context.Request.Context(), context.Request.FormValue("code"))
	if err != nil {
		s.logger.Error("code exchange wrong: " + err.Error())
		context.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}
	googleUser, err := s.userService.GetGoogleUser(context.Request.Context(), token.AccessToken)
	if err != nil {
		s.logger.Error(err.Error())
		context.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}
	user, err := s.userService.GetOrCreateUserByOAuth(context.Request.Context(), &User{
		Email:    googleUser.Email,
		Name:     googleUser.Name,
		Photo:    googleUser.Photo,
		AuthType: GoogleAuth,
	})
	if err != nil {
		s.logger.Error(err.Error())
		context.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}
	sid, err := s.userService.SetUserSession(context.Request.Context(), user.Id)
	if err != nil {
		s.logger.Error(err.Error())
		common.Response(context, http.StatusInternalServerError, common.ErrorServer)
		return
	}
	common.SetAuthCookie(context, sid, s.authCookieConfig.MaxAge, s.authCookieConfig.Path, s.authCookieConfig.Domain)

	context.Redirect(http.StatusTemporaryRedirect, "/")
}
