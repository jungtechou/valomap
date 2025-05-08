package service

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

type Service interface {
}

var (
	OsStat        = os.Stat
	JsonUnmarshal = json.Unmarshal
	StrconvAtoi   = strconv.Atoi
	Logger        = logrus.StandardLogger()
)
