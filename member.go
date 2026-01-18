package appie

import (
	"context"
	"fmt"
	"net/http"
)

// GetMember retrieves the member profile.
func (c *Client) GetMember(ctx context.Context) (*Member, error) {
	var member Member
	if err := c.doRequest(ctx, http.MethodGet, "/mobile-services/member/v1/member", nil, &member); err != nil {
		return nil, fmt.Errorf("get member failed: %w", err)
	}

	return &member, nil
}

// GetBonusCard retrieves the bonus card information.
func (c *Client) GetBonusCard(ctx context.Context) (*BonusCard, error) {
	var card BonusCard
	if err := c.doRequest(ctx, http.MethodGet, "/mobile-services/bonuscard/v1/card", nil, &card); err != nil {
		return nil, fmt.Errorf("get bonus card failed: %w", err)
	}

	return &card, nil
}
