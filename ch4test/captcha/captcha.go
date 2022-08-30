package captcha

import (
	"crypto/subtle"
	"image"
)

type Challenger interface {
	Challenge() (img image.Image, imgText string)
}

type Prompter interface {
	Prompt(img image.Image) string
}

func ChallengeUser(c Challenger, p Prompter) bool {
	img, expAnswer := c.Challenge()
	userAnswer := p.Prompt(img)

	if subtle.ConstantTimeEq(int32(len(expAnswer)), int32(len(userAnswer))) == 0 {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(userAnswer), []byte(expAnswer)) == 1
}
