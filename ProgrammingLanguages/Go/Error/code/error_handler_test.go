package code

import (
	"fmt"
	"testing"
)

func TestReadConfig(t *testing.T) {
	existFileName := "test.txt"
	notExitsFileName := "test_none.txt"

	data, err := ReadConfig(existFileName)
	if err != nil {
		t.Errorf("Read exist file get error: %v", err)
	}
	fmt.Println(data)

	_, err = ReadConfig(notExitsFileName)
	if err == nil {
		t.Error("Read not exist file not get error")
	}
}
