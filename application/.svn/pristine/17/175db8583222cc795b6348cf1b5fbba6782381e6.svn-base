package squid

import (
	"application/api/presenter"
	"application/api/presenter/squid"
	"application/mongodb"
	pb "application/pkg/proto/danmu/message"
	"application/pkg/utils"
	"application/pkg/utils/log"
	"application/redis"
	"context"
	"errors"
	"math/rand"
	"sort"
	"time"
)

func RobotOrder(game *squid.Game) error {
	if game.State != 0 {
		return errors.New("已封盘")
	}
	ctx := context.Background()
	robots, err := mongodb.GetAllRobots(ctx)
	if err != nil {
		log.Error("GetAllRobots error:", err)
		return err
	}
	numMin := int64(utils.LubanTables.TBWood.Get("random_coin_numin").NumInt)
	numMax := int64(utils.LubanTables.TBWood.Get("random_coin_numax").NumInt)
	minBetPrice := int64(utils.LubanTables.TBWood.Get("price_one").NumInt)

	// 随机每个机器人的投注时间
	betTimes, start := generateRandomTimes(game.StartTime, game.CloseTime, len(robots))
	sort.Slice(betTimes, func(i, j int) bool { return betTimes[i] < betTimes[j] })

	timeIndex := 0
	for _, robot := range robots {
		if robot.Squid.BetPrices != 0 {
			continue // 当前机器人已下注
		}
		track := rand.Int31n(4) + 1                    // 随机选择赛道(1,2,3,4)
		chips := rand.Int63n(numMax-numMin+1) + numMin // 随机筹码个数
		betPrice := chips * minBetPrice

		delay := time.Duration((betTimes[timeIndex] - start) * int64(time.Millisecond))
		if delay < 0 {
			delay = 0
		}
		timeIndex++

		robotCopy := robot
		betPriceCopy := betPrice
		trackCopy := track
		time.AfterFunc(delay, func() {
			err := SimulateBetRequest(game, robotCopy, trackCopy, betPriceCopy)
			if err != nil {
				log.Error("Error simulating bet request for robot:", robotCopy.Account, err)
			}
		})
	}
	return nil
}
func calculateChips(remainingChips, numRobotsLeft int64) int64 { // 计算单个机器人应该获得的筹码数
	if numRobotsLeft <= 0 {
		return remainingChips
	}
	// 确保每个机器人至少能分到1个筹码
	maxPossible := remainingChips - numRobotsLeft
	if maxPossible > 0 {
		return rand.Int63n(maxPossible) + 1
	}
	return 1
}

func generateRandomTimes(start, close int64, n int) ([]int64, int64) {
	var times []int64

	now, _, err := redis.BPop(redis.SquidPrefix)
	if err != nil {
		log.Error(err)
	}
	if start < now {
		start = now // 确保开始时间不早于当前时间
	}
	openDuration := close - start
	if openDuration <= 0 {
		return nil, start
	}
	for i := 0; i < n; i++ {
		randomOffset := rand.Int63n(openDuration)
		times = append(times, start+randomOffset)
	}
	return times, start
}

func SimulateBetRequest(game *squid.Game, robot *presenter.UserInfo, track int32, betPrice int64) error {
	if game.State != 0 {
		log.Error("已封盘")
		return errors.New("已封盘")
	}

	// 检查机器人余额是否充足
	if err := mongodb.SquidRobotSupply(robot, betPrice); err != nil {
		log.Error("SquidRobotSupply error: ", err)
		return err
	}

	// 玩家Round值在结算时更新, 首次进入游戏、死亡、退赛置1
	if robot.Squid.RoundId == 0 {
		robot.Squid.RoundId = 1
	}

	// 下注
	if err := mongodb.SquidOrder(robot, betPrice, track); err != nil {
		log.Error("SquidOrder error: ", err)
		return err
	}

	log.Debugf("机器人%s, 选择赛道: %d, 下注金额: %d, 当前轮次: %d, 实际下注时间: %s",
		robot.Account, track, betPrice, robot.Squid.RoundId, time.Now().Format("2006-01-02 15:04:05"))

	// 添加下注的玩家账号 ID 到 Redis
	if err := redis.AddSquidPlayer(robot.Account); err != nil {
		log.Error("Failed to add player ID to Redis:", err)
		return err
	}

	// 通知client
	if err := sendSquidBetInfoNotify(robot.Squid.RoundId, pb.SquidBetInfoEnumType_RobotOrderType); err != nil {
		log.Error("sendSquidBetInfoNotify err: ", err)
	}

	return nil
}
