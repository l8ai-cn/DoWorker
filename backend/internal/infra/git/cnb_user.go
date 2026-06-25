package git

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

func (p *CNBProvider) GetCurrentUser(ctx context.Context) (*User, error) {
	resp, err := p.doRequest(ctx, http.MethodGet, "/user", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var cnbUser struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Nickname  string `json:"nickname"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cnbUser); err != nil {
		return nil, err
	}

	name := cnbUser.Name
	if name == "" {
		name = cnbUser.Nickname
	}
	return &User{
		ID:        strconv.FormatInt(cnbUser.ID, 10),
		Username:  cnbUser.Username,
		Name:      name,
		Email:     cnbUser.Email,
		AvatarURL: cnbUser.AvatarURL,
	}, nil
}
