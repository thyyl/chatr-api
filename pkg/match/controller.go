package match

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thyyl/chatr/pkg/common"
	"gopkg.in/olahol/melody.v1"
)

func (s *HttpServer) Match(ctx *gin.Context) {
	userId, ok := ctx.Request.Context().Value(common.UserKey).(uint64)
	if !ok {
		common.Response(ctx, http.StatusUnauthorized, common.ErrorUnauthorized)
		return
	}

	_, err := s.userService.GetUserById(ctx.Request.Context(), userId)
	if err != nil {
		if errors.Is(err, common.ErrorUserNotFound) {
			common.Response(ctx, http.StatusNotFound, common.ErrorUserNotFound)
			return
		}

		s.logger.Error(err.Error())
		common.Response(ctx, http.StatusInternalServerError, common.ErrorServer)
		return
	}

	if err := s.melodyMatch.HandleRequest(ctx.Writer, ctx.Request); err != nil {
		s.logger.Error("upgrade websocket error" + err.Error())
		common.Response(ctx, http.StatusInternalServerError, common.ErrorServer)
		return
	}
}

func (s *HttpServer) HandleMatchOnConnect(session *melody.Session) {
	userId, ok := session.Request.Context().Value(common.UserKey).(uint64)
	if !ok {
		s.logger.Error("user session not found")
		return
	}

	err := s.initializeMatchSession(session, userId)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}

	ctx := context.Background()
	matchResult, err := s.matchService.Match(ctx, userId)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}

	if !matchResult.Matched {
		return
	}

	if err := s.matchService.BroadcastMatchResult(ctx, matchResult); err != nil {
		s.logger.Error(err.Error())
		return
	}
}

func (s *HttpServer) initializeMatchSession(session *melody.Session, userId uint64) error {
	session.Set(common.SessionUidKey, userId)
	return nil
}

func (s *HttpServer) HandleClose(session *melody.Session, i int, str string) error {
	userId, ok := session.Request.Context().Value(common.UserKey).(uint64)
	if !ok {
		return nil
	}

	return s.matchService.RemoveUserFromWaitList(session.Request.Context(), userId)
}
