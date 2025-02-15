package config

import (
	"go.uber.org/zap"
)

var Log *zap.SugaredLogger

func init() {
	sugar, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer sugar.Sync()

	Log = sugar.Sugar()
}
