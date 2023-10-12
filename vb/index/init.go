package index

import (
	"github.com/lgdzz/vingo-utils/vingo/config"
	"github.com/lgdzz/vingo-utils/vingo/db/mysql"
)

var Config *config.Config

func InitBackstage(option *config.Config) {
	Config = option

	mysql.InitClient(&option.Database)

}
