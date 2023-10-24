package builder

import "encoding/json"

type Payload struct {
	Ref          *string `json:"ref"`
	RefType      *string `json:"ref_type"`
	MasterBranch *string `json:"master_branch"`
	Description  *string `json:"description"`
	PusherType   *string `json:"pusher_type"`
}

type Repo struct {
	ID   *int    `json:"id"`
	Name *string `json:"name"`
	URL  *string `json:"url"`
}

type Actor struct {
	ID           *int    `json:"id"`
	Login        *string `json:"login"`
	DisplayLogin *string `json:"display_login"`
	GravatarID   *string `json:"gravatar_id"`
	URL          *string `json:"url"`
	AvatarURL    *string `json:"avatar_url"`
}

type Event struct {
	ID        *string `json:"id"`
	Type      *string `json:"type"`
	Actor     Actor   `json:"actor"`
	Repo      Repo    `json:"repo"`
	Payload   Payload `json:"payload"`
	Public    *bool   `json:"public"`
	CreatedAt *string `json:"created_at"`
}

func Unmarshal(data []byte) (*Event, error) {
	var event Event
	err := json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}
