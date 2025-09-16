package mmailer

import (
	"cmp"
	"github.com/modfin/henry/slicez"
	"github.com/stretchr/testify/assert"
	"slices"
	"testing"
)

func TestWhiteList(t *testing.T) {
	emailTo := func(emails ...string) Email {
		var tos []Address
		for _, email := range emails {
			tos = append(tos, Address{
				Name:  email,
				Email: email,
			})
		}
		return Email{
			Headers: nil,
			From: Address{
				Name:  "John Doe",
				Email: "john.doe@example.com",
			},
			To:      tos,
			Cc:      nil,
			Subject: "Test whitelist",
			Text:    "",
			Html:    "",
		}
	}

	testCases := []struct {
		email     Email
		whitelist []string
		expected  []string
	}{
		{
			email:     emailTo("markus.johansson@modularfinance.se", "andreas.elers@modularfinance.se"),
			whitelist: []string{},
			expected:  []string{"andreas.elers@modularfinance.se", "markus.johansson@modularfinance.se"},
		},
		{
			email:     emailTo("markus.johansson@modularfinance.se", "andreas.elers@modularfinance.se"),
			whitelist: []string{"markus.johansson@modularfinance.se"},
			expected:  []string{"markus.johansson@modularfinance.se"},
		},
		{
			email:     emailTo("markus.johansson@modularfinance.se", "andreas.elers@modularfinance.se"),
			whitelist: []string{"henrik.norrman@modularfinance.se"},
			expected:  []string{},
		},
		{
			email:     emailTo("joel.edstrom@modularfinance.se", "markus.johansson@modularfinance.se", "andreas.elers@modularfinance.se"),
			whitelist: []string{"henrik.norrman@modularfinance.se", "joel.edstrom@modularfinance.se"},
			expected:  []string{"joel.edstrom@modularfinance.se"},
		},
		{
			email:     emailTo("markus.johansson@modularfinance.se", "andreas.elers@modularfinance.se"),
			whitelist: []string{"markus.johansson@modularfinance.se", "andreas.elers@modularfinance.se"},
			expected:  []string{"markus.johansson@modularfinance.se", "andreas.elers@modularfinance.se"},
		},
	}

	for _, tc := range testCases {
		recipients := whitelist(tc.email.To, tc.whitelist)
		slices.SortFunc(recipients, func(a, b Address) int {
			return cmp.Compare(a.Email, b.Email)
		})
		slices.SortFunc(tc.expected, func(a, b string) int {
			return cmp.Compare(a, b)
		})
		assert.Equal(t, tc.expected, slicez.Map(recipients, func(a Address) string {
			return a.Email
		}))
	}

}
