package yam

import (
	"github.com/cnfree/props/v3/kvs"
	log "github.com/sirupsen/logrus"
	"strings"
)

func ByYaml(content string) *kvs.MapProperties {
	y := NewYamlProperties()
	err := y.Load(strings.NewReader(content))
	if err != nil {
		log.Error(err)
		return nil
	}
	return &y.MapProperties
}
