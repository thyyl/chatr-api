package match

import "encoding/json"

type MatchResultDto struct {
	AccessToken string `json:"accessToken"`
}

func (m *MatchResultDto) Encode() []byte {
	result, _ := json.Marshal(m)
	return result
}
