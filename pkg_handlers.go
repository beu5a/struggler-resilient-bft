package main

import (
	getty "github.com/apache/dubbo-getty"
)

type DefaultPackageHandler struct{}

func (h *DefaultPackageHandler) Read(ss getty.Session, data []byte) (interface{}, int, error) {
	/*
		package handling done via node
	*/
	return data, len(data), nil
}

func (h *DefaultPackageHandler) Write(ss getty.Session, p interface{}) ([]byte, error) {

	/*
		package handling done via node
	*/
	return p.([]byte), nil
}
