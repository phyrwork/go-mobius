package filter_test

import (
	"testing"
	"regexp"
	. "github.com/phyrwork/mobius/filter"
)

func TestOrFilter(t *testing.T) {
	tests := []struct {
		name string
		item string
		list []string
		match bool
	}{
		{
			"any match",
			"item",
			[]string{
				"not item",
				"item",
			},
			true,
		}, {
			"no match",
			"item",
			[]string{
				"not item 1",
				"not item 2",
			},
			false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			list := make([]Filter, 0)
			for _, pattern := range test.list {
				re := regexp.MustCompile(pattern)
				filt := RegexpFilter{Regexp: re}
				list = append(list, filt)
			}
			filt := OrFilter{List: list}
			match, err := filt.Filter(test.item)
			if err != nil {
				t.Fatalf("unexpected filter error: %v", err)
			}
			if match != test.match {
				t.Fatalf("result not equals expected: expected %v, got %v", test.match, match)
			}
		})
	}
}

func TestNorFilter(t *testing.T) {
	tests := []struct {
		name string
		item string
		list []string
		match bool
	}{
		{
			"any match",
			"item",
			[]string{
				"not item",
				"item",
			},
			false,
		}, {
			"no match",
			"item",
			[]string{
				"not item 1",
				"not item 2",
			},
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			list := make([]Filter, 0)
			for _, pattern := range test.list {
				re := regexp.MustCompile(pattern)
				filt := RegexpFilter{Regexp: re}
				list = append(list, filt)
			}
			filt := NorFilter{List: list}
			match, err := filt.Filter(test.item)
			if err != nil {
				t.Fatalf("unexpected filter error: %v", err)
			}
			if match != test.match {
				t.Fatalf("result not equals expected: expected %v, got %v", test.match, match)
			}
		})
	}
}

func TestAndFilter(t *testing.T) {
	tests := []struct {
		name string
		item string
		list []string
		match bool
	}{
		{
			"any not match",
			"item",
			[]string{
				"not item",
				"item",
			},
			false,
		}, {
			"all match",
			"item",
			[]string{
				"item",
				"item",
			},
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			list := make([]Filter, 0)
			for _, pattern := range test.list {
				re := regexp.MustCompile(pattern)
				filt := RegexpFilter{Regexp: re}
				list = append(list, filt)
			}
			filt := AndFilter{List: list}
			match, err := filt.Filter(test.item)
			if err != nil {
				t.Fatalf("unexpected filter error: %v", err)
			}
			if match != test.match {
				t.Fatalf("result not equals expected: expected %v, got %v", test.match, match)
			}
		})
	}
}