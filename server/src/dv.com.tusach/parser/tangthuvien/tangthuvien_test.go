package main

import (
	"dv.com.tusach/util"
	"fmt"
	"testing"
)

func TestPackage(t *testing.T) {
	util.LoadConfig("/home/dvan/vshared/dv/tusach/server/tusach-config.json")

	str, err := Parse("/home/dvan/vshared/dv/tusach/chapter1-raw.html", "/home/dvan/vshared/dv/tusach/chapter1.html")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(str)
}
