package vote

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestVoteValidate(t *testing.T) {
	for _, tt := range []struct {
		name        string
		poll        pollConfig
		vote        string
		expectValid bool
	}{
		// Test Method Y and N.
		{
			"Method Y, Global Y, Vote Y",
			pollConfig{
				Method:    "Y",
				GlobalYes: true,
			},
			`"Y"`,
			true,
		},
		{
			"Method Y, Vote Y",
			pollConfig{
				Method:    "Y",
				GlobalYes: false,
			},
			`"Y"`,
			false,
		},
		{
			"Method Y, Vote N",
			pollConfig{
				Method:   "Y",
				GlobalNo: false,
			},
			`"N"`,
			false,
		},
		{
			// The poll config is invalid. A poll with method Y should not allow global_no.
			"Method Y, Global N, Vote N",
			pollConfig{
				Method:   "Y",
				GlobalNo: true,
			},
			`"N"`,
			true,
		},
		{
			"Method N, Global N, Vote N",
			pollConfig{
				Method:   "N",
				GlobalNo: true,
			},
			`"N"`,
			true,
		},
		{
			"Method Y, Vote Option",
			pollConfig{
				Method:  "Y",
				Options: []int{1, 2},
			},
			`{"1":1}`,
			true,
		},
		{
			"Method Y, Vote on to many Options",
			pollConfig{
				Method:  "Y",
				Options: []int{1, 2},
			},
			`{"1":1,"2":1}`,
			false,
		},
		{
			"Method Y, Vote on one option with to high amount",
			pollConfig{
				Method:  "Y",
				Options: []int{1, 2},
			},
			`{"1":5}`,
			false,
		},
		{
			"Method Y, Vote on many option with to high amount",
			pollConfig{
				Method:    "Y",
				Options:   []int{1, 2},
				MaxAmount: 2,
			},
			`{"1":1,"2":2}`,
			false,
		},
		{
			"Method Y, Vote on one option with correct amount",
			pollConfig{
				Method:    "Y",
				Options:   []int{1, 2},
				MaxAmount: 5,
			},
			`{"1":5}`,
			true,
		},
		{
			"Method Y, Vote on one option with to less amount",
			pollConfig{
				Method:    "Y",
				Options:   []int{1, 2},
				MinAmount: 10,
			},
			`{"1":5}`,
			false,
		},
		{
			"Method Y, Vote on many options with to less amount",
			pollConfig{
				Method:    "Y",
				Options:   []int{1, 2},
				MinAmount: 10,
			},
			`{"1":1,"2":1}`,
			false,
		},
		{
			"Method Y, Vote on one option with -1 amount",
			pollConfig{
				Method:  "Y",
				Options: []int{1, 2},
			},
			`{"1":-1}`,
			false,
		},
		{
			"Method Y, Vote wrong option",
			pollConfig{
				Method:  "Y",
				Options: []int{1, 2},
			},
			`{"5":1}`,
			false,
		},

		// Test Method YN and YNA
		{
			"Method YN, Global Y, Vote Y",
			pollConfig{
				Method:    "YN",
				GlobalYes: true,
			},
			`"Y"`,
			true,
		},
		{
			"Method YN, Not Global Y, Vote Y",
			pollConfig{
				Method:    "YN",
				GlobalYes: false,
			},
			`"Y"`,
			false,
		},
		{
			"Method YNA, Global N, Vote N",
			pollConfig{
				Method:   "YNA",
				GlobalNo: true,
			},
			`"N"`,
			true,
		},
		{
			"Method YNA, Not Global N, Vote N",
			pollConfig{
				Method:    "YNA",
				GlobalYes: false,
			},
			`"N"`,
			false,
		},
		{
			"Method YNA, Y on Option",
			pollConfig{
				Method:  "YNA",
				Options: []int{1, 2},
			},
			`{"1":"Y"}`,
			true,
		},
		{
			"Method YNA, N on Option",
			pollConfig{
				Method:  "YNA",
				Options: []int{1, 2},
			},
			`{"1":"N"}`,
			true,
		},
		{
			"Method YNA, A on Option",
			pollConfig{
				Method:  "YNA",
				Options: []int{1, 2},
			},
			`{"1":"A"}`,
			true,
		},
		{
			"Method YN, A on Option",
			pollConfig{
				Method:  "YN",
				Options: []int{1, 2},
			},
			`{"1":"A"}`,
			false,
		},
		{
			"Method YN, Y on wrong Option",
			pollConfig{
				Method:  "YN",
				Options: []int{1, 2},
			},
			`{"3":"Y"}`,
			false,
		},
		{
			"Method YNA, Vote on many Options",
			pollConfig{
				Method:  "YNA",
				Options: []int{1, 2, 3},
			},
			`{"1":"Y","2":"N","3":"A"}`,
			true,
		},
		{
			"Method YNA, Amount on Option",
			pollConfig{
				Method:  "YNA",
				Options: []int{1, 2, 3},
			},
			`{"1":1}`,
			false,
		},

		// Unknown method
		{
			"Method Unknown",
			pollConfig{
				Method: "XXX",
			},
			`"Y"`,
			false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var b ballot
			if err := json.Unmarshal([]byte(tt.vote), &b.Value); err != nil {
				t.Fatalf("decoding vote: %v", err)
			}

			err := b.validate(tt.poll)

			if tt.expectValid {
				if err != nil {
					t.Fatalf("Validate returned unexpected error: %v", err)
				}
				return
			}

			if !errors.Is(err, ErrInvalid) {
				t.Fatalf("Expected ErrInvalid, got: %v", err)
			}
		})
	}
}
