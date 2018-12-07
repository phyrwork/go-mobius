package filter

import (
	"regexp"
	"fmt"
	"reflect"
)

type Filter interface {
	Filter(interface{}) (bool, error)
}

type FnFilter struct {
	Fn func (item interface{}) (bool, error)
}

func (filt FnFilter) Filter(item interface{}) (bool, error) {
	return filt.Fn(item)
}

type RegexpFilter struct {
	Regexp *regexp.Regexp
}

func (filt RegexpFilter) Filter(item interface{}) (bool, error) {
	if filt.Regexp == nil {
		return false, fmt.Errorf("regexp filter error: nil regexp")
	}
	var s string
	switch t := item.(type) {
	case fmt.Stringer:
		s = t.String()
	case string:
		s = t
	default:
		return false, fmt.Errorf("%v not supported", reflect.TypeOf(t))
	}
	return filt.Regexp.MatchString(s), nil
}

type NotFilter struct {
	Filt Filter
}

func (filt NotFilter) Filter(item interface{}) (bool, error) {
	if filt.Filt == nil {
		return false, fmt.Errorf("not filter error: nil filter")
	}
	match, err := filt.Filt.Filter(item)
	return !match, err
}

type OrFilter struct {
	List []Filter
}

func (filt OrFilter) Filter(item interface{}) (bool, error) {
	for _, filt := range filt.List {
		match, err := filt.Filter(item)
		if err != nil {
			return false, fmt.Errorf("or filter error: %v", err)
		}
		if match == true {
			return true, nil
		}
	}
	return false, nil
}

type NorFilter struct {
	List []Filter
}

func (filt NorFilter) Filter(item interface{}) (bool, error) {
	for _, filt := range filt.List {
		match, err := filt.Filter(item)
		if err != nil {
			return false, fmt.Errorf("nor filter error: %v", err)
		}
		if match == true {
			return false, nil
		}
	}
	return true, nil
}

type AndFilter struct {
	List []Filter
}

func (filt AndFilter) Filter(item interface{}) (bool, error) {
	for _, filt := range filt.List {
		match, err := filt.Filter(item)
		if err != nil {
			return false, fmt.Errorf("and filter error: %v", err)
		}
		if match == false {
			return false, nil
		}
	}
	return true, nil
}