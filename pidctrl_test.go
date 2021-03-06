package pidctrl

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

var tests = []struct {
	p       float64
	i       float64
	d       float64
	min     float64
	max     float64
	updates []*testUpdate
}{
	// p-only controller
	{
		p: 0.5,
		updates: []*testUpdate{
			{setpoint: 10, input: 5, output: 2.5},
			{input: 10, output: 0},
			{input: 15, output: -2.5},
			{input: 100, output: -45},
			{setpoint: 1, input: 0, output: 0.5},
		},
	},
	// i-only controller
	{
		i: 0.5,
		updates: []*testUpdate{
			{setpoint: 10, input: 5, duration: time.Second, output: 2.5},
			{input: 5, duration: time.Second, output: 5},
			{input: 5, duration: time.Second, output: 7.5},
			{input: 15, duration: time.Second, output: 5},
			{input: 20, duration: time.Second, output: 0},
		},
	},
	// d-only controller
	{
		d: 0.5,
		updates: []*testUpdate{
			{setpoint: 10, input: 5, duration: time.Second, output: -2.5},
			{input: 5, duration: time.Second, output: 0},
			{input: 10, duration: time.Second, output: -2.5},
		},
	},
	// pid controller
	{
		p: 0.5,
		i: 0.5,
		d: 0.5,
		updates: []*testUpdate{
			{setpoint: 10, input: 5, duration: time.Second, output: 2.5},
			{input: 10, duration: time.Second, output: 0},
			{input: 15, duration: time.Second, output: -5},
			{input: 100, duration: time.Second, output: -132.5},
			{setpoint: 1, duration: time.Second, input: 0, output: 6},
		},
	},
	// Thermostat example
	{
		p:   0.6,
		i:   1.2,
		d:   0.075,
		max: 1, // on or off
		updates: []*testUpdate{
			{setpoint: 72, input: 50, duration: time.Second, output: 1},
			{input: 51, duration: time.Second, output: 1},
			{input: 55, duration: time.Second, output: 1},
			{input: 60, duration: time.Second, output: 1},
			{input: 75, duration: time.Second, output: 0},
			{input: 76, duration: time.Second, output: 0},
			{input: 74, duration: time.Second, output: 0},
			{input: 72, duration: time.Second, output: 1},
			{input: 71, duration: time.Second, output: 1},
		},
	},
	// panic test
	{
		p:       0.5,
		i:       0.5,
		d:       0.5,
		min:     100, // min and max are swapped
		max:     1,
		updates: []*testUpdate{},
	},
}

type testUpdate struct {
	setpoint float64
	input    float64
	duration time.Duration
	output   float64
}

func (u *testUpdate) check(c *PIDController) error {
	if u.setpoint != 0 {
		c.Set(u.setpoint)
	}
	output := c.UpdateDuration(u.input, u.duration)
	if output != u.output {
		return fmt.Errorf("Bad output: %f != %f (%#v)", output, u.output, u)
	}
	return nil
}

func TestUpdate_p(t *testing.T) {
	defer func() {
		if r := recover(); (reflect.TypeOf(r)).Name() == "MinMaxError" {
			fmt.Println("Recovered Error:", r)
		} else {
			t.Error(r)
		}
	}()
	for i, test := range tests {
		t.Logf("-- test #%d", i+1)
		c := NewPIDController(test.p, test.i, test.d)
		if test.min != 0 || test.max != 0 {
			c.SetOutputLimits(test.min, test.max)
		}
		for _, u := range test.updates {
			if err := u.check(c); err != nil {
				t.Error(err)
			}
		}
	}
}
