package common

const (
	OAuthStateCookieName string = "oauthstate"
	SessionIdCookieName  string = "sid"
)

const OAuthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v3/userinfo?access_token="

const (
	JWTAuthHeader                  = "Authorization"
	JaegerHeader                   = "Uber-Trace-Id"
	ChannelIdHeader                = "X-Channel-Id"
	ChannelKey      HTTPContextKey = "channel_key"
	UserKey         HTTPContextKey = "user_key"
	ServiceIdHeader string         = "Service-Id"
	SessionUidKey                  = "SessionUid"
	SessionCidKey                  = "sesscid"
)

const (
	UserRcKey             = "rc:user"
	SessionRcKey          = "rc:session"
	MatchPubSubTopicRcKey = "rc.match"
	UserWaitListRcKey     = "rc:userwait"
	ForwardRcKey          = "rc:forward"
	ChannelUsersRcKey     = "rc:chanusers"
	OnlineUsersRcKey      = "rc:onlineusers"
	RateLimitRcKey        = "rc:ratelimit"
)

const (
	MessagePubTopic = "rc.msg.pub"
)
