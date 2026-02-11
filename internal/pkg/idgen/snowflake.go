package idgen

import (
	"github.com/bwmarrin/snowflake"
	"go.uber.org/zap"
)

var node *snowflake.Node

func Init(nodeId int64, logger *zap.Logger) error {
	var err error

	node, err = snowflake.NewNode(nodeId)
	if err != nil {
		logger.Error("Failed to initialize snowflake node", zap.Error(err))
		return err
	}

	return nil
}

func NextId() int64 {
	return node.Generate().Int64()
}
