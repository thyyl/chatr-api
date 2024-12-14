package chat

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/thyyl/chatr/pkg/common"
	"gopkg.in/olahol/melody.v1"
)

func (s *HttpServer) StartChat(ctx *gin.Context) {
	uid := ctx.Query("uid")
	userId, err := strconv.ParseUint(uid, 10, 64)
	if err != nil {
		common.Response(ctx, http.StatusBadRequest, common.ErrorInvalidParam)
		return
	}

	_, err = s.userService.GetUser(ctx.Request.Context(), userId)
	if err != nil {
		if errors.Is(err, common.ErrorUserNotFound) {
			common.Response(ctx, http.StatusNotFound, common.ErrorUserNotFound)
			return
		}
		s.logger.Error(err.Error())
		common.Response(ctx, http.StatusInternalServerError, common.ErrorServer)
		return
	}

	accessToken := ctx.Query("access_token")
	authResult, err := common.Auth(&common.AuthPayload{
		AccessToken: accessToken,
	})
	if err != nil {
		common.Response(ctx, http.StatusUnauthorized, common.ErrorUnauthorized)
		return
	}

	channelId := authResult.ChannelId
	exist, err := s.userService.IsChannelUserExists(ctx.Request.Context(), channelId, userId)
	if err != nil {
		s.logger.Error(err.Error())
		common.Response(ctx, http.StatusInternalServerError, common.ErrorServer)
		return
	}

	if !exist {
		common.Response(ctx, http.StatusNotFound, common.ErrorChannelOrUserNotFound)
		return
	}

	if err := s.melodyChat.HandleRequest(ctx.Writer, ctx.Request); err != nil {
		s.logger.Error("upgrade websocket error: " + err.Error())
		common.Response(ctx, http.StatusInternalServerError, common.ErrorServer)
		return
	}
}

func (s *HttpServer) ForwardAuth(ctx *gin.Context) {
	channelId, ok := ctx.Request.Context().Value(common.ChannelKey).(uint64)
	if !ok {
		common.Response(ctx, http.StatusUnauthorized, common.ErrorUnauthorized)
		return
	}

	ctx.Writer.Header().Set(common.ChannelIdHeader, strconv.FormatUint(channelId, 10))
	ctx.Status(http.StatusOK)
}

func (s *HttpServer) GetChannelUsers(ctx *gin.Context) {
	channelId, ok := ctx.Request.Context().Value(common.ChannelKey).(uint64)
	if !ok {
		common.Response(ctx, http.StatusUnauthorized, common.ErrorUnauthorized)
		return
	}

	userIds, err := s.userService.GetChannelUserIds(ctx.Request.Context(), channelId)
	if err != nil {
		s.logger.Error(err.Error())
		common.Response(ctx, http.StatusInternalServerError, common.ErrorServer)
		return
	}

	userIdsDto := []string{}
	for _, userId := range userIds {
		userIdsDto = append(userIdsDto, strconv.FormatUint(userId, 10))
	}

	ctx.JSON(http.StatusOK, &UserIdsDto{UserIds: userIdsDto})
}

func (s *HttpServer) GetOnlineUsers(ctx *gin.Context) {
	channelId, ok := ctx.Request.Context().Value(common.ChannelKey).(uint64)
	if !ok {
		common.Response(ctx, http.StatusUnauthorized, common.ErrorUnauthorized)
		return
	}

	userIds, err := s.userService.GetOnlineUserIds(ctx.Request.Context(), channelId)
	if err != nil {
		s.logger.Error(err.Error())
		common.Response(ctx, http.StatusInternalServerError, common.ErrorServer)
		return
	}

	userIdsDto := []string{}
	for _, userId := range userIds {
		userIdsDto = append(userIdsDto, strconv.FormatUint(userId, 10))
	}
	ctx.JSON(http.StatusOK, &UserIdsDto{UserIds: userIdsDto})
}

func (s *HttpServer) ListMessages(ctx *gin.Context) {
	channelId, ok := ctx.Request.Context().Value(common.ChannelKey).(uint64)
	if !ok {
		common.Response(ctx, http.StatusUnauthorized, common.ErrorUnauthorized)
		return
	}

	pageState := ctx.Query("ps")
	messages, nextPageState, err := s.chatService.ListMessages(ctx.Request.Context(), channelId, pageState)
	if err != nil {
		s.logger.Error(err.Error())
		common.Response(ctx, http.StatusInternalServerError, common.ErrorServer)
		return
	}

	messageDtos := []MessageDto{}
	for _, message := range messages {
		messageDtos = append(messageDtos, *message.ToPresenter())
	}

	ctx.JSON(http.StatusOK, &MessagesDto{
		Messages:      messageDtos,
		NextPageState: nextPageState,
	})
}

func (s *HttpServer) DeleteChannel(ctx *gin.Context) {
	channelId, ok := ctx.Request.Context().Value(common.ChannelKey).(uint64)
	if !ok {
		common.Response(ctx, http.StatusUnauthorized, common.ErrorUnauthorized)
		return
	}

	uid := ctx.Query("delby")
	userId, err := strconv.ParseUint(uid, 10, 64)
	if err != nil {
		common.Response(ctx, http.StatusBadRequest, common.ErrorInvalidParam)
		return
	}

	exist, err := s.userService.IsChannelUserExists(ctx.Request.Context(), channelId, userId)
	if err != nil {
		s.logger.Error(err.Error())
		common.Response(ctx, http.StatusInternalServerError, common.ErrorServer)
		return
	}

	if !exist {
		common.Response(ctx, http.StatusBadRequest, common.ErrorChannelOrUserNotFound)
		return
	}

	err = s.channelService.DeleteChannel(ctx.Request.Context(), channelId)
	if err != nil {
		s.logger.Error(err.Error())
		common.Response(ctx, http.StatusInternalServerError, common.ErrorServer)
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessMessage{
		Message: "ok",
	})
}

func (s *HttpServer) HandleChatOnConnect(session *melody.Session) {
	uid := session.Request.URL.Query().Get("uid")
	userId, err := strconv.ParseUint(uid, 10, 64)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}

	accessToken := session.Request.URL.Query().Get("access_token")
	authResult, err := common.Auth(&common.AuthPayload{
		AccessToken: accessToken,
	})

	if err != nil {
		s.logger.Error(err.Error())
	}
	if authResult.Expired {
		s.logger.Error(common.ErrorTokenExpired.Error())
	}

	channelId := authResult.ChannelId
	err = s.initializeChatSession(session, channelId, userId)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}

	if err := s.chatService.BroadcastConnectMessage(context.Background(), channelId, userId); err != nil {
		s.logger.Error(err.Error())
		return
	}
}

func (s *HttpServer) initializeChatSession(session *melody.Session, channelID, userID uint64) error {
	ctx := context.Background()
	if err := s.userService.AddOnlineUser(ctx, channelID, userID); err != nil {
		return err
	}
	if err := s.forwarderService.RegisterChannelSession(ctx, channelID, userID, s.messageSubscriber.subscriberId); err != nil {
		return err
	}
	session.Set(common.SessionCidKey, channelID)
	return nil
}

func (s *HttpServer) HandleChatOnMessage(session *melody.Session, data []byte) {
	chatMessageDto, err := DecodeToMessageDto(data)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}

	message, err := chatMessageDto.ToMessage(session.Request.URL.Query().Get("access_token"))
	if err != nil {
		s.logger.Error(err.Error())
		return
	}

	switch message.Event {
	case EventText:
		if err := s.chatService.BroadcastTextMessage(context.Background(), message.ChannelId, message.UserId, message.Payload); err != nil {
			s.logger.Error(err.Error())
		}
	case EventAction:
		if err := s.chatService.BroadcastActionMessage(context.Background(), message.ChannelId, message.UserId, Action(message.Payload)); err != nil {
			s.logger.Error(err.Error())
		}
	case EventSeen:
		messageId, err := strconv.ParseUint(message.Payload, 10, 64)
		if err != nil {
			s.logger.Error(err.Error())
			return
		}

		if err := s.chatService.MarkMessageSeen(context.Background(), message.ChannelId, message.UserId, messageId); err != nil {
			s.logger.Error(err.Error())
		}
	case EventFile:
		if err := s.chatService.BroadcastFileMessage(context.Background(), message.ChannelId, message.UserId, message.Payload); err != nil {
			s.logger.Error(err.Error())
		}
	default:
		s.logger.Error("unknown event type")
	}
}

func (s *HttpServer) HandleChatOnClose(session *melody.Session, i int, str string) error {
	userID, err := strconv.ParseUint(session.Request.URL.Query().Get("uid"), 10, 64)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	accessToken := session.Request.URL.Query().Get("access_token")
	authResult, err := common.Auth(&common.AuthPayload{
		AccessToken: accessToken,
	})
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	if authResult.Expired {
		s.logger.Error(common.ErrorTokenExpired.Error())
		return common.ErrorTokenExpired
	}
	channelID := authResult.ChannelId
	err = s.userService.DeleteOnlineUser(context.Background(), channelID, userID)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	err = s.forwarderService.RemoveChannelSession(context.Background(), channelID, userID)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	return s.chatService.BroadcastActionMessage(context.Background(), channelID, userID, OfflineMessage)
}
