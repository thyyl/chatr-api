package match

type User struct {
	Id   uint64
	Name string
}

type MatchResult struct {
	Matched     bool
	UserId      uint64
	PeerId      uint64
	ChannelId   uint64
	AccessToken string
}

func (r *MatchResult) ToDto() *MatchResultDto {
	return &MatchResultDto{
		AccessToken: r.AccessToken,
	}
}
