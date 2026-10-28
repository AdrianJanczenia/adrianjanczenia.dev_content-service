package redis

import (
	_ "embed"

	"github.com/go-redis/redis/v8"
)

//go:embed scripts/issue_token.lua
var issueTokenSource string

//go:embed scripts/redeem_token.lua
var redeemTokenSource string

var (
	issueTokenScript  = redis.NewScript(issueTokenSource)
	redeemTokenScript = redis.NewScript(redeemTokenSource)
)
