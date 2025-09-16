package mmailer

import (
	"fmt"
	"log"
	"strings"
)

func whitelist(recipients []Address, whiteList []string) []Address {
	if len(whiteList) == 0 {
		return recipients
	}
	fmt.Println("applying whitelist: ", whiteList)
	var tmp []Address
	for _, whitelisted := range whiteList {
		for _, recipient := range recipients {
			if strings.ToLower(recipient.Email) == strings.ToLower(whitelisted) {
				tmp = append(tmp, recipient)
				break
			}
		}
	}
	log.Println(fmt.Sprintf("White list recipients : %v", tmp))
	return tmp
}
