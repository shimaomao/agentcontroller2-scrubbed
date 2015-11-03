package redisdata
import (
	"github.com/Jumpscale/agentcontroller2/core"
	"time"
	"encoding/json"
	"fmt"
)


func (redisData *RedisData) RespondToCommandAsJustQueued(agentID core.AgentID, command *core.Command) error {

	db := redisData.pool.Get()
	defer db.Close()

	resultPlaceholder := core.CommandResult{
		ID:        command.ID,
		Gid:       int(agentID.GID),
		Nid:       int(agentID.NID),
		State:     core.STATE_QUEUED,
		StartTime: int64(time.Duration(time.Now().UnixNano()) / time.Millisecond),
	}

	data, err := json.Marshal(&resultPlaceholder)

	if err != nil {
		panic(fmt.Sprintf("Failed to marshal JSON for some reason!! %s", err))
	}

	_, err =
	db.Do("HSET", fmt.Sprintf(hashCmdResults, command.ID), fmt.Sprintf("%d:%d", agentID.GID, agentID.NID), data)

	if err != nil {
		return fmt.Errorf("%s: %v", redisErrorMessage, err)
	}

	return nil
}

func getAgentResultQueue(result *core.CommandResult) string {
	return fmt.Sprintf(cmdQueueAgentResponse, result.ID, result.Gid, result.Nid)
}

func (redisData *RedisData) SetCommandResult(result *core.CommandResult) error {

	db := redisData.pool.Get()
	defer db.Close()

	resultJson, err := json.Marshal(&result)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal JSON for some reason!! %s", err.Error()))
	}

	key := fmt.Sprintf("%d:%d", result.Gid, result.Nid)
	_, err = db.Do("HSET", fmt.Sprintf(hashCmdResults, result.ID), key, resultJson)
	if err != nil {
		return fmt.Errorf("%s: %v", redisErrorMessage, err)
	}

	// push message to client result queue queue
	_, err = db.Do("RPUSH", getAgentResultQueue(result), resultJson)
	if err != nil {
		return fmt.Errorf("%s: %v", redisErrorMessage, err)
	}

	return nil
}

func (redisData *RedisData) SignalCommandAsQueued(commandID string) error {
	db := redisData.pool.Get()
	defer db.Close()

	_, err := db.Do("RPUSH", fmt.Sprintf(cmdQueueCmdQueued, commandID), "queued")

	if err != nil {
		return fmt.Errorf("%s: %v", redisErrorMessage, err)
	}

	return nil
}