package mmailer

import "math/rand"

type SelectStrategy func([]Service) Service

func SelectRandom(s []Service) Service {
	i := rand.Int31n(int32(len(s)))
	return s[i]
}
