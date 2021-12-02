package svc

import (
	"context"
	"fmt"
	"math"
	"mfn/mmailer"
	"testing"
)

type TestService struct {
	name string
}

func (t *TestService) Name() string {
	return t.name
}
func (t *TestService) Send(ctx context.Context, email mmailer.Email) (res []mmailer.Response, err error) {
	return
}
func (t *TestService) UnmarshalPosthook(body []byte) ([]mmailer.Posthook, error) {
	return nil, nil
}

var s0 = WithWeight(2, &TestService{"2"})
var s1 = WithWeight(10, &TestService{"10"})
var s2 = WithWeight(20, &TestService{"20"})
var s3 = WithWeight(30, &TestService{"30"})
var sers = []mmailer.Service{s0, s1, s2, s3}

func TestNewWeightService(t *testing.T) {

	sum := 2 + 10 + 20 + 30

	c := map[string]int{}

	count := 20000

	for i := 0; i < count; i++ {
		ser := SelectWeighted(sers)
		v := c[ser.Name()]
		c[ser.Name()] = v + 1
	}

	for k, v := range c {
		var w uint
		for _, s := range sers {
			if s.Name() == k {
				w = s.(*weightService).weight
			}
		}
		got := 100 * v / count
		exp := int(100 * w / uint(sum))
		fmt.Println("Weight", k, "got", v, fmt.Sprintf("%d%%", got), "expected", fmt.Sprintf("%d%%", exp))

		if math.Abs(float64(got-exp)) > 2 {
			t.Fatal("got", got, ", expected ~", exp)
		}

	}
}

func TestRandomService(t *testing.T) {
	c := map[string]int{}

	count := 20000

	for i := 0; i < count; i++ {
		ser := mmailer.SelectRandom(sers)
		v := c[ser.Name()]
		c[ser.Name()] = v + 1
	}

	for k, v := range c {
		got := 100 * v / count
		exp := 100 / len(sers)
		fmt.Println("Name", k, "got", v, fmt.Sprintf("%d%%", got), "expected", fmt.Sprintf("%d%%", exp))

		if math.Abs(float64(got-exp)) > 2 {
			t.Fatal("got", got, ", expected ~", exp)
		}

	}
}

func TestSelectRoundRobin(t *testing.T) {
	c := map[string]int{}

	count := 20000

	rr := SelectRoundRobin()
	for i := 0; i < count; i++ {
		ser := rr(sers)
		v := c[ser.Name()]
		c[ser.Name()] = v + 1
	}

	for k, v := range c {
		got := 100 * v / count
		exp := 100 / len(sers)
		fmt.Println("Name", k, "got", v, fmt.Sprintf("%d%%", got), "expected", fmt.Sprintf("%d%%", exp))

		if v != count/len(sers) {
			t.Fatal("got", v, ", expected ", count/len(sers))
		}
	}
}
