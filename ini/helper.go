package ini

import (
	"github.com/cnfree/props/v3/kvs"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
)

func ByIni(content string) *kvs.MapProperties {
	props, err := ReadIni(ioutil.NopCloser(strings.NewReader(content)))
	if err != nil {
		logrus.Error(err)
		return nil
	}
	return &props.MapProperties
}
