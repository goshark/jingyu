package Uuid

import (
	"github.com/gofrs/uuid"
	"github.com/gogf/gf/os/glog"
)


/*
	生成uuid
*/
func Generate_uuid() string {
	// Create a Version 4 UUID.
	u2, err := uuid.NewV4()
	if err != nil {
		glog.Error(err)
		return ""
	}
	return u2.String()
}

