package redis

import (
	"application/api/presenter"
	"application/api/presenter/squid"
	"application/pkg/utils/log"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"math"
	"strconv"
	"strings"
	"time"
)

const MaxListLength int64 = 30
const Prefix = "app:"
const SquidPrefix = "squid:"
const GlassPrefix = "glass:"
const LadderPrefix = "ladder:"
const RechargePrefix = "recharge:"
const MaxRecords = 20000

var redisInstance *redis.Client

func NewRedisClient() error {
	var redisConfig presenter.RedisConfig
	if err := viper.UnmarshalKey("common.redis", &redisConfig); err != nil {
		return err
	}
	redisInstance = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		//Password: redisConfig.Password,
		DB: 0, // 默认DB 0
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	if err := redisInstance.Ping(ctx).Err(); err != nil {
		return err
	}
	return nil
}

func TryLock(ctx context.Context, key string, sec int64) error {
	timeout := time.Duration(sec) * time.Second
	deadline := time.Now().Add(timeout)
	retryInterval := time.Second // 重试间隔

	for {
		nowStr := time.Now().Format("2006-01-02 15:04:05") + fmt.Sprintf(".%03d", time.Now().Nanosecond()/1e6)
		succ, err := redisInstance.SetNX(ctx, key, nowStr, timeout).Result()
		if err != nil {
			return err
		}
		if succ {
			return nil
		}

		if time.Now().After(deadline) {
			return errors.New("lock acquisition timeout")
		}
		time.Sleep(retryInterval) // 等待一段时间后重试
	}
}
func UnLock(ctx context.Context, key string) error {
	_, err := redisInstance.Del(ctx, key).Result()
	return err
}

func CloseRedisClient() {
	err := redisInstance.Close()
	if err != nil {
		log.Errorf("Failed to close Redis client: %v", err)
	}
}

func GetFetchTime() (int64, error) {
	val, err := redisInstance.Get(context.Background(), Prefix+"fetchtime").Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, err
	}
	return val, err
}

func Push(timestamp int64, vol float64, prefixs ...string) error {
	if len(prefixs) == 0 {
		return nil
	}
	for _, prefix := range prefixs {
		err := redisInstance.Set(context.Background(), prefix+"fetchtime", timestamp, 0).Err()
		if err != nil {
			return err
		}
		_, err = redisInstance.LPush(context.Background(), prefix+"list", fmt.Sprintf("%d:%f", timestamp, vol)).Result()
		if err != nil {
			return err
		}
	}
	return nil
}
func Pop() (int64, float64, error) {
	result, err := redisInstance.RPop(context.Background(), Prefix+"list").Result()
	if err != nil {
		return 0, 0, err
	}
	//log.Info("Pop:", result)
	arrs := strings.Split(result, ":")
	ts, _ := strconv.ParseInt(arrs[0], 10, 64)
	v, _ := strconv.ParseFloat(arrs[1], 64)
	return ts, math.Round(v*1000) / 1000, nil
}
func BPop(prefix string) (int64, float64, error) {
	result, err := redisInstance.BRPop(context.Background(), time.Second*3, prefix+"list").Result()
	if err != nil {
		return 0, 0, err
	}
	//log.Info("BPop:", result)
	arrs := strings.Split(result[1], ":")
	ts, _ := strconv.ParseInt(arrs[0], 10, 64)
	v, _ := strconv.ParseFloat(arrs[1], 64)
	return ts, math.Round(v*1000) / 1000, nil
}

func SetSessionToken(sessionToken string, account string, duration time.Duration) {
	redisInstance.Set(context.Background(), Prefix+"token:"+sessionToken, account, duration)
}

func DelSessionToken(sessionToken string) {
	redisInstance.Del(context.Background(), Prefix+"token:"+sessionToken)
}

func GetAccountBySessionToken(sessionToken string) (string, error) {
	return redisInstance.Get(context.Background(), Prefix+"token:"+sessionToken).Result()
}

func AddInviteToRedis(account string, downstream string, date string) error {
	key := Prefix + "invites:" + account + ":" + date
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err := redisInstance.SAdd(ctx, key, downstream).Result()
	if err != nil {
		return err
	}
	redisInstance.Expire(ctx, key, 24*time.Hour)
	return nil
}

func GetInvitesFromToday(account string, date string) ([]string, error) {
	key := Prefix + "invites:" + account + ":" + date
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	invites, err := redisInstance.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return invites, nil
}

func AddSquidPlayer(account string) error {
	return redisInstance.SAdd(context.Background(), SquidPrefix+"bet_players", account).Err()
}
func RemoveSquidPlayer(account string) error { // close封盘, 取消下注
	return redisInstance.SRem(context.Background(), SquidPrefix+"bet_players", account).Err()
}
func GetSquidPlayers() ([]string, error) { // 获取最近所有下注的玩家账号
	members, err := redisInstance.SMembers(context.Background(), SquidPrefix+"bet_players").Result()
	if err != nil {
		return nil, err
	}
	return members, nil
}

func AddSquidData(roundTime int64, data squid.Data) error {
	key := SquidPrefix + "close_time_data"
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error("Error marshalling data:", err)
		return err
	}
	// 添加数据到ZSet，使用时间戳作为分数
	_, err = redisInstance.ZAdd(context.Background(), key, redis.Z{
		Score:  float64(roundTime),
		Member: jsonData,
	}).Result()
	if err != nil {
		log.Error("Error adding data to ZSet:", err)
		return err
	}
	currentSize, err := redisInstance.ZCard(context.Background(), key).Result()
	if err != nil {
		log.Error("Error getting ZSet size:", err)
		return err
	}
	if currentSize > MaxRecords {
		_, err = redisInstance.ZRemRangeByRank(context.Background(), key, 0, currentSize-MaxRecords-1).Result()
		if err != nil {
			log.Error("Error removing old records:", err)
			return err
		}
	}
	return nil
}
func GetSquidData() (*squid.Data, []string, error) {
	ctx := context.Background()
	key := SquidPrefix + "close_time_data"
	// 获取ZSet中的所有元素，从分数最高到最低
	results, err := redisInstance.ZRevRange(ctx, key, 0, -1).Result()
	if err != nil {
		log.Error("Error retrieving all data from ZSet:", err)
		return nil, nil, err
	}
	if len(results) == 0 {
		log.Info("No data available in ZSet")
		return nil, nil, nil
	}
	var latest squid.Data
	err = json.Unmarshal([]byte(results[0]), &latest)
	if err != nil {
		log.Error("Error unmarshalling the latest JSON data:", err)
		return nil, nil, err
	}
	reorderedData := make([]string, len(results))
	for i, jsonData := range results {
		var result squid.Data
		err := json.Unmarshal([]byte(jsonData), &result)
		if err != nil {
			log.Error("Error unmarshalling JSON data:", err)
			continue
		}
		reorderedJson, err := json.Marshal(result)
		if err != nil {
			log.Error("Error marshalling reordered data:", err)
			continue
		}
		// 将重新排序的JSON字符串添加到结果数组
		reorderedData[i] = string(reorderedJson)
	}
	return &latest, reorderedData, nil
}

func AddGlassPlayer(account string) error {
	return redisInstance.SAdd(context.Background(), GlassPrefix+"bet_players", account).Err()
}
func RemoveGlassPlayer(account string) error { // 结算后移除
	return redisInstance.SRem(context.Background(), GlassPrefix+"bet_players", account).Err()
}
func GetGlassPlayers() ([]string, error) { // 获取本轮所有下注的玩家账号
	members, err := redisInstance.SMembers(context.Background(), GlassPrefix+"bet_players").Result()
	if err != nil {
		return nil, err
	}
	return members, nil
}

func AddLadderPlayer(account string, roundNum int64) error {
	key := fmt.Sprintf("%s%d_bet_players", LadderPrefix, roundNum)
	return redisInstance.SAdd(context.Background(), key, account).Err()
}
func RemoveLadderPlayer(account string, roundNum int64) error { // 结算后移除
	key := fmt.Sprintf("%s%d_bet_players", LadderPrefix, roundNum)
	return redisInstance.SRem(context.Background(), key, account).Err()
}
func GetLadderPlayers(roundNum int64) ([]string, error) { // 获取本轮所有下注的玩家账号
	key := fmt.Sprintf("%s%d_bet_players", LadderPrefix, roundNum)
	members, err := redisInstance.SMembers(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	return members, nil
}

//	func IncrementLadderBets(ctx context.Context, roundNum int64, betId string, amount int64) error {
//		redisKey := fmt.Sprintf("%sround:%d", LadderPrefix, roundNum)
//		_, err := redisInstance.HIncrBy(ctx, redisKey, betId, amount).Result()
//		if err != nil {
//			log.Errorf("Failed to increment bet amount in Redis for round %d, betId %s: %v", roundNum, betId, err)
//			return err
//		}
//
//		openTime := int64(utils.LubanTables.TBLadder.Get("open_time").NumInt)
//		closeTime := int64(utils.LubanTables.TBLadder.Get("close_time").NumInt)
//		settlementTime := int64(utils.LubanTables.TBLadder.Get("settlement_time").NumInt)
//		// 将秒转换为 time.Duration (纳秒)
//		expiration := time.Duration(openTime+closeTime+settlementTime) * time.Second
//		if _, err := redisInstance.Expire(ctx, redisKey, expiration).Result(); err != nil {
//			log.Errorf("Failed to set expiration for Redis key %s: %v", redisKey, err)
//			return err
//		}
//		return nil
//	}
//
// IncrementLadderBets 增加指定轮次和 BetId 的下注金额
func IncrementLadderBets(ctx context.Context, roundNum int64, betId string, amount int64) error {
	redisKey := fmt.Sprintf("%sround:%d", LadderPrefix, roundNum)
	_, err := redisInstance.HIncrBy(ctx, redisKey, betId, amount).Result()
	if err != nil {
		log.Errorf("Failed to increment bet amount in Redis for round %d, betId %s: %v", roundNum, betId, err)
		return err
	}
	return nil
}
func RemoveLadderBets(ctx context.Context, roundNum int64) error {
	redisKey := fmt.Sprintf("%sround:%d", LadderPrefix, roundNum)
	_, err := redisInstance.Del(ctx, redisKey).Result()
	if err != nil {
		log.Errorf("Failed to remove bets for round %d: %v", roundNum, err)
		return err
	}
	log.Debugf("Successfully removed bets for round %d", roundNum)
	return nil
}
func GetLadderBets(ctx context.Context, roundNum int64) (map[string]int64, error) {
	redisKey := fmt.Sprintf("%sround:%d", LadderPrefix, roundNum)
	result, err := redisInstance.HGetAll(ctx, redisKey).Result()
	if err != nil {
		log.Errorf("Failed to get all bet amounts from Redis for round %d: %v", roundNum, err)
		return nil, err
	}
	bets := make(map[string]int64)
	for betId, amountStr := range result {
		amount, err := strconv.ParseInt(amountStr, 10, 64)
		if err != nil {
			log.Errorf("Failed to parse bet amount for betId %s in round %d: %v", betId, roundNum, err)
			continue
		}
		bets[betId] = amount
	}
	return bets, nil
}

//func StoreHashValue(field int64, value string) error {
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	redisKey := fmt.Sprintf("%sroundHashes", LadderPrefix)
//	fieldStr := strconv.FormatInt(field, 10)
//	_, err := redisInstance.HSet(ctx, redisKey, fieldStr, value).Result()
//	if err != nil {
//		log.Errorf("Failed to set hash value for key %s, field %s: %v", redisKey, fieldStr, err)
//		return err
//	}
//	return nil
//}
//func GetAllRoundHashes() (map[int64]string, error) {
//	redisKey := fmt.Sprintf("%sroundHashes", LadderPrefix)
//	result, err := redisInstance.HGetAll(context.Background(), redisKey).Result()
//	if err != nil {
//		return nil, fmt.Errorf("failed to retrieve hashes from Redis: %v", err)
//	}
//	roundHashes := make(map[int64]string)
//	for roundStr, hashValue := range result {
//		roundNum, err := strconv.ParseInt(roundStr, 10, 64)
//		if err != nil {
//			return nil, fmt.Errorf("failed to parse round number %s: %v", roundStr, err)
//		}
//		roundHashes[roundNum] = hashValue
//	}
//	return roundHashes, nil
//}
//func RemoveHashField(field int64) error {
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	redisKey := fmt.Sprintf("%sroundHashes", LadderPrefix)
//	fieldStr := strconv.FormatInt(field, 10)
//	_, err := redisInstance.HDel(ctx, redisKey, fieldStr).Result()
//	if err != nil {
//		log.Errorf("Failed to delete field %s from Redis key %s: %v", fieldStr, redisKey, err)
//		return err
//	}
//	return nil
//}
