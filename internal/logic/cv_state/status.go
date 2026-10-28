package cv_state

type IssueStatus string

const (
	Issued              IssueStatus = "issued"
	Reused              IssueStatus = "reused"
	Collision           IssueStatus = "collision"
	CaptchaNotFound     IssueStatus = "not_found"
	CaptchaNotSolved    IssueStatus = "not_solved"
	InvalidPassword     IssueStatus = "invalid_password"
	NoTriesLeft         IssueStatus = "no_tries"
	UnsupportedLanguage IssueStatus = "unsupported_language"
	InvalidIssueState   IssueStatus = "invalid_state"
)

type RedeemStatus string

const (
	Redeemed           RedeemStatus = "redeemed"
	TokenNotFound      RedeemStatus = "not_found"
	WrongLanguage      RedeemStatus = "wrong_language"
	InvalidRedeemState RedeemStatus = "invalid_state"
)
