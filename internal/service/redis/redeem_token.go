package redis

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/cv_state"
	"github.com/go-redis/redis/v8"
)

func (c *Client) RedeemToken(ctx context.Context, token, lang string) (cv_state.RedeemStatus, error) {
	if c.rdb == nil {
		return "", fmt.Errorf("redis script client is not initialized")
	}

	tokenKey := "cv:token:" + token

	for range 4 {
		raw, err := c.rdb.Get(ctx, tokenKey).Result()
		if errors.Is(err, redis.Nil) {
			return cv_state.TokenNotFound, nil
		}
		if err != nil {
			return "", err
		}

		separator := strings.IndexByte(raw, ':')
		if separator < 1 || separator == len(raw)-1 {
			return cv_state.InvalidRedeemState, nil
		}

		resultKey := "cv:result:" + raw[separator+1:]

		status, err := redeemTokenScript.Run(ctx, c.rdb, []string{tokenKey, resultKey}, lang, raw).Text()
		if err != nil || status != "retry" {
			return cv_state.RedeemStatus(status), err
		}
	}

	return "", fmt.Errorf("token state changed repeatedly")
}
