package api

import (
	"testing"
)

func TestRegexpArguments(t *testing.T) {
	type cases struct {
		res bool
		arg string
	}
	casess := []cases{
		{
			res: false,
			arg: "<fermi>",
		},
		{
			res: true,
			arg: "fermi",
		},
		{
			res: true,
			arg: "<ffads",
		},
		{
			res: true,
			arg: "fads>",
		},
	}

	for _, v := range casess {
		if regexpArguments(v.arg) != v.res {
			t.Error("fail regexpArguments!")
		}
	}
}

func TestInArray(t *testing.T) {
	type cases struct {
		res bool
		arg string
	}
	casess := []cases{
		{
			res: false,
			arg: "119",
		},
		{
			res: false,
			arg: "kaidd",
		},
		{
			res: true,
			arg: "ppgod",
		},
	}

	array := []string{"rapgod", "ppgod", "weigod", "110", "911"}
	for _, v := range casess {
		res := inArray(array, v.arg)
		if res != v.res {
			t.Error("fail inarray!")
		}
	}
}
