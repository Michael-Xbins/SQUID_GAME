package mongodb

import (
	"application/api/presenter"
	"application/api/presenter/auction"
	"application/api/presenter/compete"
	"application/api/presenter/glass"
	"application/api/presenter/ladder"
	rechargedb "application/api/presenter/recharge"
	"application/api/presenter/squid"
	"application/api/presenter/web"
	pb "application/pkg/proto/danmu/message"
	"application/pkg/utils"
	"application/pkg/utils/log"
	"application/wallet"
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var mongoInstance *mongo.Client

type DataStandard interface {
	TableName() string
	PrimaryId() interface{}
}

func NewMongoDBClient() (context.CancelFunc, error) {
	connectionString := viper.GetString("common.mongodb.connection_string")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(viper.GetInt64("common.mongodb.timeout"))*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		cancel()
		return nil, err
	}
	mongoInstance = client
	if err := initializeData(ctx); err != nil {
		cancel()
		return nil, err
	}
	return cancel, nil
}
func initializeData(ctx context.Context) error {
	if err := ensureDocumentExists(ctx, &squid.Game{Id: 0}); err != nil {
		return err
	}
	if err := ensureDocumentExists(ctx, &squid.Fund{Id: 0}); err != nil {
		return err
	}
	for i := 1; i <= squid.TotalRounds; i++ {
		round := &squid.GlobalRound{
			RoundID: i,
			Track1:  squid.Track{},
			Track2:  squid.Track{},
			Track3:  squid.Track{},
			Track4:  squid.Track{},
		}
		if err := ensureDocumentExists(ctx, round); err != nil {
			return err
		}
	}

	if err := ensureDocumentExists(ctx, &compete.Game{Id: 0}); err != nil {
		return err
	}

	//if err := ensureDocumentExists(ctx, &glass.Game{Id: 0}); err != nil {
	//	return err
	//}
	//if err := ensureDocumentExists(ctx, &glass.Fund{Id: 0}); err != nil {
	//	return err
	//}

	if err := ensureDocumentExists(ctx, &ladder.Game{Id: 0}); err != nil {
		return err
	}
	if err := ensureDocumentExists(ctx, &ladder.Fund{Id: 0}); err != nil {
		return err
	}

	if err := ensureDocumentExists(ctx, &rechargedb.Fund{Id: 0}); err != nil {
		return err
	}
	if err := ensureDocumentExists(ctx, &auction.Game{Id: 0}); err != nil {
		return err
	}
	//创建机器人
	hasRobot := viper.GetBool("common.has_robot")
	if hasRobot {
		configNum := int64(utils.LubanTables.TBWood.Get("robot_num").NumInt)
		curRobotCount, err := CountRobots(ctx)
		if err != nil {
			log.Error("CountRobots error: ", err)
		}
		if curRobotCount < configNum {
			robotCoin := int64(utils.LubanTables.TBWood.Get("robot_coin").NumInt)
			for i := curRobotCount + 1; i <= configNum; i++ {
				robotAccount := fmt.Sprintf("robot%d", i)
				robot := &presenter.UserInfo{
					Account:  robotAccount,
					Nickname: robotAccount,
					Balance:  robotCoin,
					Squid:    presenter.Squid{RoundId: 1, BetPricesPerRound: make([]int64, squid.TotalRounds)},
					IsRobot:  true,
				}
				err := Insert(context.Background(), robot)
				if err != nil {
					log.Error(err)
				}
			}
		}
	}
	return nil
}
func ensureDocumentExists(ctx context.Context, data DataStandard) error {
	err := Find(ctx, data, data.PrimaryId())
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		if err := Insert(ctx, data); err != nil {
			return err
		}
	}
	return nil
}

func Find(ctx context.Context, data DataStandard, primary interface{}) (err error) {
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(data.TableName())
	singleResult := collection.FindOne(ctx, bson.M{"_id": primary})
	if err = singleResult.Err(); err != nil {
		return err
	}
	return singleResult.Decode(data)
}

func Insert(ctx context.Context, data DataStandard) error {
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(data.TableName())
	if _, err := collection.InsertOne(ctx, data); err != nil {
		return err
	}
	return nil
}
func InsertMany(ctx context.Context, orders []*ladder.Order) error {
	if len(orders) == 0 {
		return nil
	}
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(orders[0].TableName())
	documents := make([]interface{}, len(orders))
	for i, order := range orders {
		documents[i] = order
	}
	if _, err := collection.InsertMany(ctx, documents); err != nil {
		return err
	}
	return nil
}

func Update(ctx context.Context, data DataStandard, update interface{}) error {
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(data.TableName())
	if update == nil {
		update = bson.M{"$set": data}
	}
	_, err := collection.UpdateByID(ctx, data.PrimaryId(), update)
	return err
}

func Delete(ctx context.Context, data DataStandard) error {
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(data.TableName())
	_, err := collection.DeleteOne(ctx, bson.M{"_id": data.PrimaryId()})
	return err
}

func DeleteOldOrders(ctx context.Context, days int32) (int64, error) {
	// 假设 mongoInstance 是一个全局的 *mongo.Client 实例
	historyList := &auction.HistoryList{}
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(historyList.TableName())

	// 获取当前时间
	now := time.Now()
	// 计算30天前的日期
	cutoffDate := now.AddDate(0, 0, -int(days))

	// 创建删除过滤器
	filter := bson.M{
		"timestamp": bson.M{"$lt": cutoffDate}, // 删除 orderDate 早于 cutoffDate 的订单
	}

	// 执行删除操作
	deleteResult, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	// 返回被删除的文档数量
	return deleteResult.DeletedCount, nil
}

func GetOrderHistoryByDate(ctx context.Context, account string, day int32) ([]*auction.HistoryList, error) {
	historyList := &auction.HistoryList{}
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(historyList.TableName())
	now := time.Now()
	// 计算截止日期
	endDate := now.AddDate(0, 0, -int(day)) // 注意这里用的是负数来向后计算
	endDateMillis := endDate.UnixNano() / 1e6
	// 创建查询过滤器
	filter := bson.M{
		"account":   account,
		"timestamp": bson.M{"$gte": endDateMillis},
	}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var list []*auction.HistoryList
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func GetBuyOrderWithMaxPrice(ctx context.Context) (*auction.BuyOrderList, error) {
	buyOrderList := &auction.BuyOrderList{}
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(buyOrderList.TableName())

	// 创建排序选项，按 price 降序排列
	sortOptions := options.FindOne().SetSort(bson.D{{Key: "price", Value: -1}})

	// 执行查询，只返回一条结果
	var result auction.BuyOrderList
	err := collection.FindOne(ctx, bson.M{}, sortOptions).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

func GetSellOrderWithMaxPrice(ctx context.Context) (*auction.SellOrderList, error) {
	buyOrderList := &auction.SellOrderList{}
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(buyOrderList.TableName())

	// 创建排序选项，按 price 降序排列
	sortOptions := options.FindOne().SetSort(bson.D{{Key: "price", Value: 1}})

	// 执行查询，只返回一条结果
	var result auction.SellOrderList
	err := collection.FindOne(ctx, bson.M{}, sortOptions).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

func GetBuyOrderListByAccount(ctx context.Context, account string) ([]*auction.BuyOrderList, error) {
	buyOrderList := &auction.BuyOrderList{}
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(buyOrderList.TableName())
	cursor, err := collection.Find(ctx, bson.M{"account": account})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var list []*auction.BuyOrderList
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func GetSellOrderListByAccount(ctx context.Context, account string) ([]*auction.SellOrderList, error) {
	sellOrderList := &auction.SellOrderList{}
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(sellOrderList.TableName())
	cursor, err := collection.Find(ctx, bson.M{"account": account})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var list []*auction.SellOrderList
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}
func GetTop5SellOrderList(ctx context.Context) ([]*pb.SellOrderInfo, error) {
	buyOrderList := &auction.SellOrderList{}
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(buyOrderList.TableName())
	// 创建排序和限制选项
	findOptions := options.Find()
	//findOptions.SetSort(bson.D{{Key: "timestamp", Value: -1}})                          // 按 timestamp 降序排列
	findOptions.SetSort(bson.D{{Key: "price", Value: 1}, {Key: "timestamp", Value: 1}}) // 价格升序，时间升序
	findOptions.SetLimit(5)                                                             // 限制结果数量为 5

	// 执行查询
	cursor, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// 解码结果到切片中
	var results []*auction.SellOrderList
	for cursor.Next(ctx) {
		var sellOrderList auction.SellOrderList
		err := cursor.Decode(&sellOrderList)
		if err != nil {
			return nil, err
		}
		results = append(results, &sellOrderList)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	var sellOrderInfo []*pb.SellOrderInfo
	if len(results) > 0 {
		for _, v := range results {
			sellOrderInfo = append(sellOrderInfo, &pb.SellOrderInfo{
				Id:    v.Id,
				Count: v.TotalCount - v.CompletedCount,
				Price: v.Price,
			})
		}
	}
	return sellOrderInfo, nil
}
func GetTop5BuyOrderList(ctx context.Context) ([]*pb.BuyOrderInfo, error) {
	buyOrderList := &auction.BuyOrderList{}
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(buyOrderList.TableName())
	// 创建排序和限制选项
	findOptions := options.Find()
	//findOptions.SetSort(bson.D{{Key: "timestamp", Value: -1}}) // 按 timestamp 降序排列
	findOptions.SetSort(bson.D{{Key: "price", Value: -1}, {Key: "timestamp", Value: 1}}) // 价格升序，时间升序
	findOptions.SetLimit(5)                                                              // 限制结果数量为 5

	// 执行查询
	cursor, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// 解码结果到切片中
	var results []*auction.BuyOrderList
	for cursor.Next(ctx) {
		var buyOrder auction.BuyOrderList
		err := cursor.Decode(&buyOrder)
		if err != nil {
			return nil, err
		}
		results = append(results, &buyOrder)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	var buyOrderInfo []*pb.BuyOrderInfo

	if len(results) > 0 {
		for _, v := range results {
			buyOrderInfo = append(buyOrderInfo, &pb.BuyOrderInfo{
				Id:    v.Id,
				Count: v.TotalCount - v.CompletedCount,
				Price: v.Price,
			})
		}
	}
	return buyOrderInfo, nil
}

func FindAllBuyOrders() ([]auction.BuyOrderList, error) {
	buyOrderList := &auction.BuyOrderList{}
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(buyOrderList.TableName())
	// 创建一个空的查询条件，表示查询所有文档
	filter := bson.D{{}}
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.TODO())
	var orders []auction.BuyOrderList
	for cur.Next(context.TODO()) {
		var order auction.BuyOrderList
		if err := cur.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}
func FindAllSellOrders() ([]auction.SellOrderList, error) {
	sellOrderList := &auction.SellOrderList{}
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(sellOrderList.TableName())
	// 创建一个空的查询条件，表示查询所有文档
	filter := bson.D{{}}
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.TODO())
	var orders []auction.SellOrderList
	for cur.Next(context.TODO()) {
		var order auction.SellOrderList
		if err := cur.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func GetAllRobots(ctx context.Context) ([]*presenter.UserInfo, error) {
	userinfo := &presenter.UserInfo{}
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(userinfo.TableName())
	cursor, err := collection.Find(ctx, bson.M{"isRobot": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var robots []*presenter.UserInfo
	if err = cursor.All(ctx, &robots); err != nil {
		return nil, err
	}
	return robots, nil
}
func CountRobots(ctx context.Context) (int64, error) {
	userinfo := &presenter.UserInfo{}
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(userinfo.TableName())
	count, err := collection.CountDocuments(ctx, bson.M{"isRobot": true})
	if err != nil {
		return 0, err
	}
	return count, nil
}

// BulkUpdateUserInfos 批量更新用户信息UserInfo
func BulkUpdateUserInfos(ctx context.Context, userInfos []*presenter.UserInfo) error {
	var userInfo presenter.UserInfo
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(userInfo.TableName())
	var operations []mongo.WriteModel // 写操作的切片
	for _, userInfo := range userInfos {
		update := mongo.NewUpdateOneModel()
		update.SetFilter(bson.M{"_id": userInfo.Account})
		update.SetUpdate(bson.M{"$set": userInfo})
		operations = append(operations, update)
	}
	// 执行批量写操作
	_, err := collection.BulkWrite(ctx, operations, options.BulkWrite().SetOrdered(false))
	if err != nil {
		return err
	}
	return nil
}
func FindUserInfoByToken(ctx context.Context, token interface{}) (userInfo *presenter.UserInfo, err error) {
	userInfo = &presenter.UserInfo{}
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(userInfo.TableName())
	singleResult := collection.FindOne(ctx, bson.M{"session_token": token})
	if err = singleResult.Err(); err != nil {
		return
	}
	err = singleResult.Decode(userInfo)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}

func CheckBalance(userinfo *presenter.UserInfo, value int64) bool {
	if userinfo.Balance < value {
		return false
	}
	return true
}
func AddAmount(userInfo *presenter.UserInfo, value int64) {
	userInfo.Balance += value
}
func DecrAmount(userInfo *presenter.UserInfo, value int64) {
	userInfo.Balance -= value
}

func CheckUSDT(userinfo *presenter.UserInfo, value int64) bool {
	if userinfo.USDT < value {
		return false
	}
	return true
}
func AddUSDT(userInfo *presenter.UserInfo, value int64) {
	userInfo.USDT += value
}
func DecrUSDT(userInfo *presenter.UserInfo, value int64) {
	userInfo.USDT -= value
}
func AddUsdtPool(fund *rechargedb.Fund, value int64) {
	fund.UsdtPool += value
}
func AddWithdraw(fund *rechargedb.Fund, amount int64, value int64) {
	fund.WithdrawnAmount += amount
	fund.ServiceFee += value
}

func CheckVoucher(userinfo *presenter.UserInfo, value int64) bool {
	if userinfo.Voucher < value {
		return false
	}
	return true
}
func AddVoucher(userInfo *presenter.UserInfo, value int64) {
	userInfo.Voucher += value
}
func DecrVoucher(userInfo *presenter.UserInfo, value int64) {
	userInfo.Voucher -= value
}

func TransferAgent(userInfo *presenter.UserInfo) {
	// 木头人
	transferAgentDetail(&userInfo.Agent.Squid, userInfo)
	//玻璃桥
	transferAgentDetail(&userInfo.Agent.Glass, userInfo)
	//拔河
	transferAgentDetail(&userInfo.Agent.Compete, userInfo)
}
func transferAgentDetail(agentDetail *presenter.AgentDetail, userInfo *presenter.UserInfo) {
	unclaimed := agentDetail.Unclaimed
	agentDetail.Claimed += unclaimed // 将未领取的佣金转移到已领取佣金
	AddAmount(userInfo, unclaimed)   // 更新用户余额
	agentDetail.Unclaimed = 0        // 重置未领取佣金为0
}

func TransferUsdtAgent(userInfo *presenter.UserInfo) {
	oldUsdt := userInfo.USDT
	unclaimed := userInfo.UsdtAgent.Unclaimed
	userInfo.UsdtAgent.Claimed += unclaimed // 将未领取的佣金转移到已领取佣金
	AddUSDT(userInfo, unclaimed)            // 更新用户余额
	userInfo.UsdtAgent.Unclaimed = 0        // 重置未领取佣金为0

	// UsdtFlow埋点
	log.InfoJson("USDT入口",
		zap.String("Account", userInfo.Account),
		zap.String("ActionType", log.Flow),
		zap.String("FlowType", log.UsdtFlow),
		zap.String("From", log.FromAgent),
		zap.String("Flag", log.FlagIn),
		zap.Int64("Amount", unclaimed),
		zap.Int64("Old", oldUsdt),
		zap.Int64("New", userInfo.USDT),
		zap.Int64("CreatedAt", time.Now().UnixMilli()),
	)
}

func UpdateUserinfo(userInfo *presenter.UserInfo) {
	today := time.Now().Format("2006-01-02")
	needUpdate := false
	if today != userInfo.TurnOver.LastDate {
		userInfo.TurnOver.Squid.DailyTurnOver = 0
		userInfo.TurnOver.Glass.DailyTurnOver = 0
		userInfo.TurnOver.Compete.DailyTurnOver = 0
		userInfo.TurnOver.LastDate = today
		needUpdate = true
	}
	if today != userInfo.Agent.LastDate {
		userInfo.Agent.Squid.DailyAgent = 0
		userInfo.Agent.Glass.DailyAgent = 0
		userInfo.Agent.Compete.DailyAgent = 0
		userInfo.Agent.LastDate = today
		needUpdate = true
	}
	if today != userInfo.UsdtRecharge.LastDate {
		userInfo.UsdtRecharge.DailyRecharge = 0
		userInfo.UsdtRecharge.LastDate = today
		userInfo.UsdtRecharge.DownLineDailyRecharge = make(map[string]bool)
		needUpdate = true
	}
	if today != userInfo.UsdtAgent.LastDate {
		userInfo.UsdtAgent.DailyAgent = 0
		userInfo.UsdtAgent.LastDate = today
		needUpdate = true
	}
	if today != userInfo.Welfare.LastDate {
		times := utils.LubanTables.TBApp.Get("poor_count").NumInt
		userInfo.Welfare.Times = times
		userInfo.Welfare.LastDate = today
		needUpdate = true
	}
	if today != userInfo.DailyTask.LastDate {
		userInfo.DailyTask.LastDate = today
		userInfo.DailyTask.Tasks = make(map[int32]*presenter.TaskDetail)
		needUpdate = true
	}
	if needUpdate {
		if err := Update(context.Background(), userInfo, nil); err != nil {
			log.Error("update error: ", err)
		}
	}
}

// ----------------------------------------------木头人------------------------------------------------------------
func AddSquidTurnOver(userInfo *presenter.UserInfo, value int64) {
	today := time.Now().Format("2006-01-02")
	if today != userInfo.TurnOver.LastDate {
		userInfo.TurnOver.Squid.DailyTurnOver = 0
		userInfo.TurnOver.LastDate = today
	}
	userInfo.TurnOver.Squid.DailyTurnOver += value
	userInfo.TurnOver.Squid.TotalTurnOver += value
}

func AddFirstPass(userInfo *presenter.UserInfo, squidFund *squid.Fund, value int64) {
	if userInfo.IsRobot {
		squidFund.RobotPool += value
	} else {
		userInfo.Squid.FirstPass.Pool += value
	}
}

func AddSquidAgent(userInfo *presenter.UserInfo, squidFund *squid.Fund, pumpDetails *presenter.PumpDetails) {
	// 更新 上级/上上级 代理的抽水
	if userInfo.UpLine != "" {
		upLineUserInfo := &presenter.UserInfo{}
		_ = Find(context.Background(), upLineUserInfo, userInfo.UpLine)
		log.Debugf("account:%v 有上线 %v, 上线佣金:%v, 上上线用户: %v, 上上线佣金: %v", userInfo.Account, upLineUserInfo.Account, pumpDetails.UpLineContribution, upLineUserInfo.UpLine, pumpDetails.UpUpLineContribution)
		updateSquidAgent(upLineUserInfo, pumpDetails.UpLineContribution)
		upLineUserInfo1 := &presenter.UserInfo{}
		_ = Find(context.Background(), upLineUserInfo1, userInfo.UpLine)
		if upLineUserInfo.UpLine != "" {
			upUpLineUserInfo := &presenter.UserInfo{}
			_ = Find(context.Background(), upUpLineUserInfo, upLineUserInfo.UpLine)
			updateSquidAgent(upUpLineUserInfo, pumpDetails.UpUpLineContribution)
		} else {
			AddSquidHouseCut(userInfo.IsRobot, squidFund, pumpDetails.UpUpLineContribution) // 没有上上线，则放到庄家抽水中
		}
	} else {
		AddSquidHouseCut(userInfo.IsRobot, squidFund, pumpDetails.AgentContribution) // 没有上线，则放到庄家抽水中
	}
}
func updateSquidAgent(userInfo *presenter.UserInfo, contribution int64) {
	today := time.Now().Format("2006-01-02")
	if userInfo.Agent.LastDate != today {
		userInfo.Agent.Squid.DailyAgent = 0
		userInfo.Agent.LastDate = today
	}
	userInfo.Agent.Squid.DailyAgent += contribution
	userInfo.Agent.Squid.TotalAgent += contribution
	userInfo.Agent.Squid.Unclaimed += contribution
	if err := Update(context.Background(), userInfo, nil); err != nil {
		log.Error("updateSquidAgent error: ", err)
	}
}

func AddSquidHouseCut(isRobot bool, fund *squid.Fund, value int64) {
	if isRobot {
		fund.RobotPool += value
	} else {
		fund.HouseCut += value
	}
}
func AddSquidJackpot(fund *squid.Fund, value int64) {
	fund.Jackpot += value
}
func DecrSquidJackpot(fund *squid.Fund, value int64) {
	fund.Jackpot -= value
}

func SquidNextRound(userInfo *presenter.UserInfo) {
	if userInfo.Squid.RoundId >= squid.TotalRounds {
		userInfo.Squid.RoundId = 1
		userInfo.Squid.Track = 0
		userInfo.Squid.BetPrices = 0
		userInfo.Squid.CanJackpot = true
		UpdateDailyTaskProgress(2, userInfo, squid.TotalRounds)
	} else {
		userInfo.Squid.RoundId += 1
		userInfo.Squid.Track = 0
		userInfo.Squid.BetPrices = 0
		userInfo.Squid.CanJackpot = false
		UpdateDailyTaskProgress(2, userInfo, userInfo.Squid.RoundId)
	}
}

func UpdateDailyTaskProgress(taskType int32, userInfo *presenter.UserInfo, progress int32) {
	if !userInfo.IsRobot {
		if taskType == 2 {
			// 直接设置特定任务的进度
			updateTask(userInfo, taskType, progress, false)
		} else {
			// 递增任务进度
			updateTask(userInfo, taskType, 1, true)
		}
	}
}
func updateTask(userInfo *presenter.UserInfo, taskType int32, progress int32, increment bool) {
	if taskDetail, ok := userInfo.DailyTask.Tasks[taskType]; ok {
		if increment {
			taskDetail.Progress += progress
		} else {
			taskDetail.Progress = max(taskDetail.Progress, progress)
		}
	} else {
		userInfo.DailyTask.Tasks[taskType] = &presenter.TaskDetail{Progress: progress}
	}
}

func SquidReset(userInfo *presenter.UserInfo) { //退赛,死亡,处理jackpot(完成7轮)
	userInfo.Squid.RoundId = 1
	userInfo.Squid.Track = 0
	userInfo.Squid.BetPrices = 0
	userInfo.Squid.BetPricesPerRound = make([]int64, squid.TotalRounds)
	userInfo.Squid.CanJackpot = false
}

func SquidOrder(userInfo *presenter.UserInfo, price int64, track int32) error {
	oldTrack := userInfo.Squid.Track
	curRoundId := userInfo.Squid.RoundId
	DecrAmount(userInfo, price)
	userInfo.Squid.Track = track
	userInfo.Squid.BetPrices += price
	userInfo.Squid.BetPricesPerRound[curRoundId-1] += price
	if err := Update(context.Background(), userInfo, nil); err != nil {
		log.Error("Update userInfo error: ", err)
		return err
	}
	globalSquidRound := &squid.GlobalRound{}
	if err := Find(context.Background(), globalSquidRound, curRoundId); err != nil {
		log.Error("Failed to find globalSquidRound: ", err)
		return err
	}
	trackField := fmt.Sprintf("track%d", track)
	var trackUpdate bson.D
	if userInfo.IsRobot {
		trackUpdate = bson.D{{"$inc", bson.D{{trackField + ".robot_total_bet_prices", price}}}}
	} else {
		trackUpdate = bson.D{{"$inc", bson.D{{trackField + ".total_bet_prices", price}}}}
	}
	// 新下注赛道,才增加玩家数量(同赛道加注不变)
	if oldTrack == 0 && oldTrack != track {
		trackUpdate = append(trackUpdate, bson.E{Key: "$inc", Value: bson.D{{trackField + ".player_nums", 1}}})
	} else if oldTrack != 0 && oldTrack != track {
		return errors.New("本轮已选赛道")
	}
	if err := Update(context.Background(), globalSquidRound, trackUpdate); err != nil {
		log.Error("Failed to update globalSquidRound: ", err)
		return err
	}
	return nil
}

func SquidCancel(userInfo *presenter.UserInfo) error {
	oldTrack := userInfo.Squid.Track
	betPrices := userInfo.Squid.BetPrices
	curRoundId := userInfo.Squid.RoundId
	globalSquidRound := &squid.GlobalRound{}
	if err := Find(context.Background(), globalSquidRound, curRoundId); err != nil {
		log.Error("Find globalSquidRound error: ", err)
		return err
	}
	AddAmount(userInfo, betPrices)
	userInfo.Squid.Track = 0
	userInfo.Squid.BetPrices = 0
	userInfo.Squid.BetPricesPerRound[curRoundId-1] = 0
	if err := Update(context.Background(), userInfo, nil); err != nil {
		return err
	}
	trackField := fmt.Sprintf("track%d", oldTrack)
	trackUpdate := bson.D{
		{"$inc", bson.D{
			{trackField + ".total_bet_prices", -betPrices}, // 减少赛道的总注额
			{trackField + ".player_nums", -1},              // 减少参与该赛道的玩家数量
		}},
	}
	if err := Update(context.Background(), globalSquidRound, trackUpdate); err != nil {
		log.Error("Update globalSquidRound error: ", err)
		return err
	}
	return nil
}

func SquidSwitch(userInfo *presenter.UserInfo, newTrack int32) error {
	oldTrack := userInfo.Squid.Track
	betPrices := userInfo.Squid.BetPrices
	if oldTrack != newTrack {
		globalSquidRound := &squid.GlobalRound{}
		if err := Find(context.Background(), globalSquidRound, userInfo.Squid.RoundId); err != nil {
			log.Error("Find globalSquidRound error: ", err)
		}
		oldTrackField := fmt.Sprintf("track%d", oldTrack)
		newTrackField := fmt.Sprintf("track%d", newTrack)
		update := bson.D{
			{"$inc", bson.D{
				{oldTrackField + ".total_bet_prices", -betPrices}, // 减少原赛道的总注额
				{oldTrackField + ".player_nums", -1},              // 减少原赛道的玩家数量
				{newTrackField + ".total_bet_prices", betPrices},  // 增加新赛道的总注额
				{newTrackField + ".player_nums", 1},               // 增加新赛道的玩家数量
			}},
		}
		if err := Update(context.Background(), globalSquidRound, update); err != nil {
			log.Error("Update globalSquidRound error: ", err)
			return err
		}
	}

	userInfo.Squid.Track = newTrack
	if err := Update(context.Background(), userInfo, nil); err != nil {
		log.Error("Update userInfo error: ", err)
		return err
	}
	return nil
}

// SquidRobotSupply 机器人余额不足, 库存来补充余额
func SquidRobotSupply(robot *presenter.UserInfo, betPrice int64) error {
	if !CheckBalance(robot, betPrice) {
		squidFund := &squid.Fund{}
		if err := Find(context.Background(), squidFund, 0); err != nil {
			log.Error("Error retrieving squid fund:", err)
			return err
		}
		oldBalance := robot.Balance
		oldRobotPool := squidFund.RobotPool
		configNum := int64(utils.LubanTables.TBWood.Get("robot_num").NumInt)
		supplement := squidFund.RobotPool / configNum
		squidFund.RobotPool -= supplement
		robot.Balance += supplement
		log.Debugf("机器人%s余额不足, 当前余额:%d, 当前机器人库存:%d, 随机下注额:%d, 库存补充:%d,  新余额:%d, 新机器人库存:%d", robot.Account, oldBalance, oldRobotPool, betPrice, supplement, robot.Balance, squidFund.RobotPool)
		if e := Update(context.Background(), squidFund, nil); e != nil {
			log.Error("Error updating squid fund:", e)
			return e
		}
	}
	return nil
}

func ResetAllGlobalSquidRound() {
	for i := 1; i <= squid.TotalRounds; i++ {
		globalSquidRound := &squid.GlobalRound{}
		if err := Find(context.Background(), globalSquidRound, i); err != nil {
			log.Error("Error finding round:", err)
			continue
		}
		tracks := []*squid.Track{&globalSquidRound.Track1, &globalSquidRound.Track2, &globalSquidRound.Track3, &globalSquidRound.Track4}
		for _, track := range tracks {
			track.TotalBetPrices = 0
			track.PlayerNums = 0
			track.RobotTotalBetPrices = 0
		}
		if err := Update(context.Background(), globalSquidRound, nil); err != nil {
			log.Error("Error updating round:", err)
			continue
		}
	}
}
func ResetGlobalSquidRound(globalSquidRound *squid.GlobalRound) {
	tracks := []*squid.Track{&globalSquidRound.Track1, &globalSquidRound.Track2, &globalSquidRound.Track3, &globalSquidRound.Track4}
	for _, track := range tracks {
		track.TotalBetPrices = 0
		track.PlayerNums = 0
		track.RobotTotalBetPrices = 0
	}
}

func SquidDailyFirstPass(userInfo *presenter.UserInfo) {
	today := time.Now().Format("2006-01-02")
	if userInfo.Squid.FirstPass.LastFirstPassDate == today {
		return // 今天已经领取过
	}
	oldBalance := userInfo.Balance
	AddAmount(userInfo, userInfo.Squid.FirstPass.Pool)
	if !userInfo.IsRobot {
		log.Debugf("领取firstpass, user: %v, 上次领取时间: %v, 奖金池: %v", userInfo.Account, userInfo.Squid.FirstPass.LastFirstPassDate, userInfo.Squid.FirstPass.Pool)
		// coinFlow埋点
		log.InfoJson("金币入口",
			zap.String("Account", userInfo.Account),
			zap.String("ActionType", log.Flow),
			zap.String("FlowType", log.CoinFlow),
			zap.String("From", log.FromFirstPass),
			zap.String("Flag", log.FlagIn),
			zap.Int64("Amount", userInfo.Squid.FirstPass.Pool), //兑换了
			zap.Int64("Old", oldBalance),                       //旧游戏币
			zap.Int64("New", userInfo.Balance),                 //新游戏币
			zap.Int64("CreatedAt", time.Now().UnixMilli()),
		)
	}
	userInfo.Squid.FirstPass.Pool = 0
	userInfo.Squid.FirstPass.LastFirstPassDate = today
}

func FindSquidUserOrders(ctx context.Context, account string) ([]squid.Order, error) {
	var squidOrder *squid.Order
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(squidOrder.TableName())
	filter := bson.M{}
	if account != "" {
		filter["account"] = account
	}
	opts := options.Find().SetSort(bson.D{{"timestamp", -1}}) // Sorting by timestamp descending
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var orders []squid.Order
	for cursor.Next(ctx) {
		var order squid.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

// ----------------------------------------------玻璃桥------------------------------------------------------------
func FindGlassRoundNumOrders(ctx context.Context, roundNum interface{}) ([]*glass.Order, error) {
	var glassOrder *glass.Order
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(glassOrder.TableName())
	filter := bson.M{}
	switch v := roundNum.(type) {
	case int, int64:
		filter["roundNum"] = v
	case string:
		filter["roundNum"] = v
	default:
		return nil, fmt.Errorf("unsupported type for roundNum: %v", reflect.TypeOf(roundNum))
	}
	opts := options.Find().SetSort(bson.D{{"timestamp", -1}}) // Sorting by timestamp descending
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []*glass.Order
	for cursor.Next(ctx) {
		var order glass.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

// BulkUpdateOrderInfos 批量更新订单信息lottery
func BulkUpdateOrderInfos(ctx context.Context, orders []*glass.Order) error {
	var glassOrder *glass.Order
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(glassOrder.TableName())
	var operations []mongo.WriteModel // 写操作的切片
	for _, order := range orders {
		filter := bson.M{"_id": order.OrderId}
		update := bson.M{
			"$set": bson.M{
				"isWin": order.IsWin,
				"bonus": order.Bonus,
			},
		}
		operation := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update)
		operations = append(operations, operation)
	}
	// 执行批量写操作
	opts := options.BulkWrite().SetOrdered(false)
	_, err := collection.BulkWrite(ctx, operations, opts)
	return err
}

func FindGlassUserOrders(ctx context.Context, account string, roundNum interface{}) ([]*glass.Order, error) {
	var glassOrder *glass.Order
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(glassOrder.TableName())
	filter := bson.M{}
	if account != "" {
		filter["account"] = account
	}
	if roundNum != nil {
		switch v := roundNum.(type) {
		case int, int64:
			filter["roundNum"] = v
		case string:
			filter["roundNum"] = v
		default:
			return nil, fmt.Errorf("unsupported type for roundNum: %v", reflect.TypeOf(roundNum))
		}
	}
	opts := options.Find().SetSort(bson.D{{"timestamp", -1}}) // Sorting by timestamp descending
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var orders []*glass.Order
	for cursor.Next(ctx) {
		var order glass.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func FindAllLotteries(ctx context.Context) ([]*glass.Lottery, error) {
	var lottery *glass.Lottery
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(lottery.TableName())
	opts := options.Find().SetSort(bson.D{{"timestamp", -1}}) // Sorting by timestamp descending
	cursor, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var lotteries []*glass.Lottery
	for cursor.Next(ctx) {
		var lottery glass.Lottery
		if err := cursor.Decode(&lottery); err != nil {
			return nil, err
		}
		lotteries = append(lotteries, &lottery)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return lotteries, nil
}

func FindAllBatchCdk(ctx context.Context) ([]*presenter.BatchCdk, error) {
	batchCdk := &presenter.BatchCdk{}
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(batchCdk.TableName())
	opts := options.Find().SetSort(bson.D{{"timestamp", -1}}) // Sorting by timestamp descending
	filter := bson.M{"_id": bson.M{"$ne": batchCdk.TableName()}}
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var batchCdks []*presenter.BatchCdk
	for cursor.Next(ctx) {
		var batchCdk presenter.BatchCdk // 创建新的实例
		if err := cursor.Decode(&batchCdk); err != nil {
			return nil, err
		}
		batchCdks = append(batchCdks, &batchCdk)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return batchCdks, nil
}

func AddGlassHouseCut(fund *glass.Fund, value int64) {
	fund.HouseCut += value
}

func AddGlassAgent(userInfo *presenter.UserInfo, glassFund *glass.Fund, pumpDetails *presenter.PumpDetails) {
	// 更新 上级/上上级 代理的抽水
	if userInfo.UpLine != "" {
		upLineUserInfo := &presenter.UserInfo{}
		_ = Find(context.Background(), upLineUserInfo, userInfo.UpLine)
		log.Debugf("account:%v 有上线 %v, 上线佣金:%v, 上上线用户: %v, 上上线佣金: %v", userInfo.Account, upLineUserInfo.Account, pumpDetails.UpLineContribution, upLineUserInfo.UpLine, pumpDetails.UpUpLineContribution)
		updateGlassAgent(upLineUserInfo, pumpDetails.UpLineContribution)
		upLineUserInfo1 := &presenter.UserInfo{}
		_ = Find(context.Background(), upLineUserInfo1, userInfo.UpLine)
		if upLineUserInfo.UpLine != "" {
			upUpLineUserInfo := &presenter.UserInfo{}
			_ = Find(context.Background(), upUpLineUserInfo, upLineUserInfo.UpLine)
			updateGlassAgent(upUpLineUserInfo, pumpDetails.UpUpLineContribution)
		} else {
			AddGlassHouseCut(glassFund, pumpDetails.UpUpLineContribution) // 没有上上线，则放到庄家抽水中
		}
	} else {
		AddGlassHouseCut(glassFund, pumpDetails.AgentContribution) // 没有上线，则放到庄家抽水中
	}
}
func updateGlassAgent(userInfo *presenter.UserInfo, contribution int64) {
	today := time.Now().Format("2006-01-02")
	if userInfo.Agent.LastDate != today {
		userInfo.Agent.Glass.DailyAgent = 0
		userInfo.Agent.LastDate = today
	}
	userInfo.Agent.Glass.DailyAgent += contribution
	userInfo.Agent.Glass.TotalAgent += contribution
	userInfo.Agent.Glass.Unclaimed += contribution
	if err := Update(context.Background(), userInfo, nil); err != nil {
		log.Error("updateGlassAgent error: ", err)
	}
}

func AddGlassTurnOver(userInfo *presenter.UserInfo, value int64) {
	today := time.Now().Format("2006-01-02")
	if today != userInfo.TurnOver.LastDate {
		userInfo.TurnOver.Glass.DailyTurnOver = 0
		userInfo.TurnOver.LastDate = today
	}
	userInfo.TurnOver.Glass.DailyTurnOver += value
	userInfo.TurnOver.Glass.TotalTurnOver += value
}

// ----------------------------------------------梯子游戏------------------------------------------------------------
func AddLadderHouseCut(fund *ladder.Fund, value int64) {
	fund.HouseCut += value
}

func AddLadderAvailableFund(fund *ladder.Fund, value int64) {
	fund.AvailableFund += value
}

func AddLadderAgent(userInfo *presenter.UserInfo, ladderFund *ladder.Fund, pumpDetails *presenter.PumpDetails) {
	// 更新 上级/上上级 代理的抽水
	if userInfo.UpLine != "" {
		upLineUserInfo := &presenter.UserInfo{}
		_ = Find(context.Background(), upLineUserInfo, userInfo.UpLine)
		log.Debugf("account:%v 有上线 %v, 上线佣金:%v, 上上线用户: %v, 上上线佣金: %v", userInfo.Account, upLineUserInfo.Account, pumpDetails.UpLineContribution, upLineUserInfo.UpLine, pumpDetails.UpUpLineContribution)
		updateLadderAgent(upLineUserInfo, pumpDetails.UpLineContribution)
		upLineUserInfo1 := &presenter.UserInfo{}
		_ = Find(context.Background(), upLineUserInfo1, userInfo.UpLine)
		if upLineUserInfo.UpLine != "" {
			upUpLineUserInfo := &presenter.UserInfo{}
			_ = Find(context.Background(), upUpLineUserInfo, upLineUserInfo.UpLine)
			updateLadderAgent(upUpLineUserInfo, pumpDetails.UpUpLineContribution)
		} else {
			AddLadderHouseCut(ladderFund, pumpDetails.UpUpLineContribution) // 没有上上线，则放到庄家抽水中
		}
	} else {
		AddLadderHouseCut(ladderFund, pumpDetails.AgentContribution) // 没有上线，则放到庄家抽水中
	}
}

func updateLadderAgent(userInfo *presenter.UserInfo, contribution int64) {
	today := time.Now().Format("2006-01-02")
	if userInfo.Agent.LastDate != today {
		userInfo.Agent.Ladder.DailyAgent = 0
		userInfo.Agent.LastDate = today
	}
	userInfo.Agent.Ladder.DailyAgent += contribution
	userInfo.Agent.Ladder.TotalAgent += contribution
	userInfo.Agent.Ladder.Unclaimed += contribution
	if err := Update(context.Background(), userInfo, nil); err != nil {
		log.Error("updateLadderAgent error: ", err)
	}
}

func AddLadderTurnOver(userInfo *presenter.UserInfo, value int64) {
	today := time.Now().Format("2006-01-02")
	if today != userInfo.TurnOver.LastDate {
		userInfo.TurnOver.Ladder.DailyTurnOver = 0
		userInfo.TurnOver.LastDate = today
	}
	userInfo.TurnOver.Ladder.DailyTurnOver += value
	userInfo.TurnOver.Ladder.TotalTurnOver += value
}

func FindLadderUserOrders(ctx context.Context, account string, roundNum interface{}) ([]*ladder.Order, error) {
	var ladderOrder *ladder.Order
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(ladderOrder.TableName())
	filter := bson.M{}
	if account != "" {
		filter["account"] = account
	}
	if roundNum != nil {
		switch v := roundNum.(type) {
		case int, int64:
			filter["roundNum"] = v
		case string:
			filter["roundNum"] = v
		default:
			return nil, fmt.Errorf("unsupported type for roundNum: %v", reflect.TypeOf(roundNum))
		}
	}
	opts := options.Find().SetSort(bson.D{{"timestamp", -1}}) // Sorting by timestamp descending
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var orders []*ladder.Order
	for cursor.Next(ctx) {
		var order ladder.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

// BulkUpdateLadderOrderInfos 批量更新梯子订单信息lottery
func BulkUpdateLadderOrderInfos(ctx context.Context, orders []*ladder.Order) error {
	var ladderOrder *ladder.Order
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(ladderOrder.TableName())
	var operations []mongo.WriteModel // 写操作的切片
	for _, order := range orders {
		filter := bson.M{"_id": order.OrderId}
		update := bson.M{
			"$set": bson.M{
				"isWin": order.IsWin,
				"bonus": order.Bonus,
			},
		}
		operation := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update)
		operations = append(operations, operation)
	}
	// 执行批量写操作
	opts := options.BulkWrite().SetOrdered(false)
	_, err := collection.BulkWrite(ctx, operations, opts)
	return err
}

func FindLadderLotteries(ctx context.Context) ([]*ladder.Lottery, error) {
	var lottery *ladder.Lottery
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(lottery.TableName())
	opts := options.Find().SetSort(bson.D{{"timestamp", -1}}) // Sorting by timestamp descending
	cursor, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var lotteries []*ladder.Lottery
	for cursor.Next(ctx) {
		var lottery ladder.Lottery
		if err := cursor.Decode(&lottery); err != nil {
			return nil, err
		}
		lotteries = append(lotteries, &lottery)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return lotteries, nil
}

func SumTotalPricesByCategory(ctx context.Context, roundNum int64) (int64, int64, error) {
	var ladderOrder *ladder.Order
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(ladderOrder.TableName())
	pipeline := mongo.Pipeline{
		// 第一步：过滤出指定 roundNum 的订单
		{{"$match", bson.M{"roundNum": roundNum}}},
		// 第二步：根据 category 分组，并计算每组的 totalPrice 总和
		{{"$group", bson.M{
			"_id":           "$category",
			"totalPriceSum": bson.M{"$sum": "$orderPrice"},
		}}},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, 0, fmt.Errorf("error executing aggregate query: %v", err)
	}
	defer cursor.Close(ctx)
	var results []struct {
		Category      int32 `bson:"_id"`
		TotalPriceSum int64 `bson:"totalPriceSum"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return 0, 0, fmt.Errorf("error getting aggregate results: %v", err)
	}
	var sumCategory1, sumCategory2 int64
	for _, result := range results {
		if result.Category == 1 {
			sumCategory1 = result.TotalPriceSum
		} else if result.Category == 2 {
			sumCategory2 = result.TotalPriceSum
		}
	}
	return sumCategory1, sumCategory2, nil
}

// ----------------------------------------------拔河------------------------------------------------------------
func FindCompeteUserOrders(ctx context.Context, account string) ([]compete.OrderInfo, error) {
	var competeOrder *compete.OrderInfo
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(competeOrder.TableName())
	filter := bson.M{}
	if account != "" {
		filter["account"] = account
	}
	opts := options.Find().SetSort(bson.D{{"timestamp", -1}}) // Sorting by timestamp descending
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var orders []compete.OrderInfo
	for cursor.Next(ctx) {
		var order compete.OrderInfo
		if err := cursor.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func AddCompeteAgent(userInfo *presenter.UserInfo, game *compete.Game, pumpDetails presenter.PumpDetails) {
	// 更新 上级/上上级 代理的抽水
	if userInfo.UpLine != "" {
		upLineUserInfo := &presenter.UserInfo{}
		_ = Find(context.Background(), upLineUserInfo, userInfo.UpLine)
		log.Debugf("account:%v 有上线 %v, 上线佣金:%v, 上上线用户: %v, 上上线佣金: %v", userInfo.Account, upLineUserInfo.Account, pumpDetails.UpLineContribution, upLineUserInfo.UpLine, pumpDetails.UpUpLineContribution)
		updateCompeteAgent(upLineUserInfo, pumpDetails.UpLineContribution)
		upLineUserInfo1 := &presenter.UserInfo{}
		_ = Find(context.Background(), upLineUserInfo1, userInfo.UpLine)
		if upLineUserInfo.UpLine != "" {
			upUpLineUserInfo := &presenter.UserInfo{}
			_ = Find(context.Background(), upUpLineUserInfo, upLineUserInfo.UpLine)
			updateCompeteAgent(upUpLineUserInfo, pumpDetails.UpUpLineContribution)
		} else {
			game.TakeAmount += pumpDetails.UpUpLineContribution // 没有上上线，则放到庄家抽水中
		}
	} else {
		game.TakeAmount += pumpDetails.AgentContribution // 没有上线，则放到庄家抽水中
	}
}
func updateCompeteAgent(userInfo *presenter.UserInfo, contribution int64) {
	today := time.Now().Format("2006-01-02")
	if userInfo.Agent.LastDate != today {
		userInfo.Agent.Compete.DailyAgent = 0
		userInfo.Agent.LastDate = today
	}
	userInfo.Agent.Compete.DailyAgent += contribution
	userInfo.Agent.Compete.TotalAgent += contribution
	userInfo.Agent.Compete.Unclaimed += contribution
	if err := Update(context.Background(), userInfo, nil); err != nil {
		log.Error("updateSquidAgent error: ", err)
	}
}

func AddCompeteTurnOver(userInfo *presenter.UserInfo, value int64) {
	today := time.Now().Format("2006-01-02")
	if today != userInfo.TurnOver.LastDate {
		userInfo.TurnOver.Compete.DailyTurnOver = 0
		userInfo.TurnOver.LastDate = today
	}
	userInfo.TurnOver.Compete.DailyTurnOver += value
	userInfo.TurnOver.Compete.TotalTurnOver += value
}

// ----------------------------------------------------交易------------------------------------------------------------------
func StoreUserWallet(ctx context.Context, address string, privateKey string, userInfo *presenter.UserInfo) error {
	if len(userInfo.Address) != 0 {
		log.Error("%s钱包地址已存在, address:%s", userInfo.Account, address)
		return errors.New("钱包地址已存在")
	}
	rechargeAddressInfo := &rechargedb.AddressInfo{
		Address:    address,
		UserId:     userInfo.Account,
		PrivateKey: privateKey,
		CreatedAt:  time.Now().UnixMilli(),
	}
	if err := Insert(context.Background(), rechargeAddressInfo); err != nil {
		log.Error(err)
		return err
	}

	userInfo.Address = address
	if err := Update(ctx, userInfo, nil); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func IsTransactionProcessed(ctx context.Context, transactionID string) bool {
	var payInfo rechargedb.PayInfo
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(payInfo.TableName())

	// 计算匹配的文档数量
	filter := bson.M{"_id": transactionID}
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Error("Error checking if transaction is processed: %v", err)
		return false
	}
	return count > 0
}

func RechargeUsdt(account string, value string, decimals int) (int64, int64, int64, error) {
	// 计算实际的代币数量
	valueBigInt, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return 0, 0, 0, errors.New("error converting value to big.Int")
	}
	decimalsBigInt := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil) // 计算10的decimals次方
	actualValue := new(big.Float).Quo(new(big.Float).SetInt(valueBigInt), new(big.Float).SetInt(decimalsBigInt))
	allocation := float64(utils.LubanTables.TBApp.Get("allocation").NumInt) // USDT兑换游戏币汇率
	multiplier := new(big.Float).SetFloat64(allocation)
	gameCoins := new(big.Float).Mul(actualValue, multiplier)

	// 实际充值的 usdt额度
	usdtAmount, _ := actualValue.Float64()
	// 充值得到的 游戏币额度(向下取整)：1美元:1000000
	gameCoinsInt, _ := gameCoins.Int64()
	// 充值得到的 兑换券额度(向下取整)：1美元:10
	voucher := int64(usdtAmount * float64(utils.LubanTables.TBApp.Get("allocatioon").NumInt))

	userInfo := &presenter.UserInfo{}
	if err := Find(context.Background(), userInfo, account); err != nil {
		log.Error(err)
		return 0, 0, 0, err
	}
	oldBalance := userInfo.Balance
	oldVoucher := userInfo.Voucher
	fund := &rechargedb.Fund{}
	if err := Find(context.Background(), fund, 0); err != nil {
		log.Error(err)
		return 0, 0, 0, err
	}

	// 更新充值流水
	AddUsdtRecharge(userInfo, int64(usdtAmount*100))

	// 更新 上级/上上级 代理 USDT抽水
	const scaleFactor = int64(1000)
	usdtSum := float64(utils.LubanTables.TBApp.Get("usdt_sum").NumInt)             //千分比
	usdtUp := float64(utils.LubanTables.TBApp.Get("usdt_up").NumInt)               //千分比
	agent := int64(usdtAmount*100*usdtSum) / scaleFactor                           //美分为单位 (*100)
	upLineUSDT := int64(usdtAmount*100*usdtSum*usdtUp) / scaleFactor / scaleFactor //美分为单位,千分比例
	upUpLineUSDT := agent - upLineUSDT

	agentGameCoins := int64(0)
	agentVoucher := int64(0)

	if userInfo.UpLine != "" {
		upLineUserInfo := &presenter.UserInfo{}
		_ = Find(context.Background(), upLineUserInfo, userInfo.UpLine)
		// 更新下线充值流水
		updateDownLineRecharge(upLineUserInfo, userInfo.Account)
		// 更新下线代理
		updateUsdtAgent(upLineUserInfo, upLineUSDT)
		if err := Update(context.Background(), upLineUserInfo, nil); err != nil {
			log.Error("error: ", err)
		}
		agentUSD := float64(upLineUSDT) / 100.0
		agentUSDTFloat := new(big.Float).SetFloat64(agentUSD)
		agentGameCoins, _ = new(big.Float).Mul(agentUSDTFloat, multiplier).Int64()
		agentVoucher = int64(agentUSD * float64(utils.LubanTables.TBApp.Get("allocatioon").NumInt))
		log.Infof("用户:%v 充值代理佣金有 上线 %v, 上线usdt: %v, 上上线用户: %v, 上上线usdt: %v", userInfo.Account, upLineUserInfo.Account, upLineUSDT, upLineUserInfo.UpLine, upUpLineUSDT)

		// 更新下下线代理
		upLineUserInfo1 := &presenter.UserInfo{}
		_ = Find(context.Background(), upLineUserInfo1, userInfo.UpLine)
		if upLineUserInfo.UpLine != "" {
			upUpLineUserInfo := &presenter.UserInfo{}
			_ = Find(context.Background(), upUpLineUserInfo, upLineUserInfo.UpLine)
			updateUsdtAgent(upUpLineUserInfo, upUpLineUSDT)
			if err := Update(context.Background(), upUpLineUserInfo, nil); err != nil {
				log.Error("error: ", err)
			}
			agentUSD := float64(agent) / 100.0
			agentUSDTFloat := new(big.Float).SetFloat64(agentUSD)
			agentGameCoins, _ = new(big.Float).Mul(agentUSDTFloat, multiplier).Int64()
			agentVoucher = int64(agentUSD * float64(utils.LubanTables.TBApp.Get("allocatioon").NumInt))

		} else {
			//AddUsdtPool(fund, upUpLineUSDT) // 没有上上线, 则放到usdt库存
		}
	} else {
		//AddUsdtPool(fund, agent) // 没有上线, 则放到usdt库存
	}

	// 更新游戏币
	retGameCoins := gameCoinsInt - agentGameCoins
	AddAmount(userInfo, retGameCoins)
	// 更新兑换券
	retVoucher := voucher - agentVoucher
	AddVoucher(userInfo, retVoucher)

	if err := Update(context.Background(), userInfo, nil); err != nil {
		log.Error(err)
		return 0, 0, 0, err
	}
	if e := Update(context.Background(), fund, nil); e != nil {
		log.Error("Error updating recharge fund:", e)
		return 0, 0, 0, e
	}
	//log.Infof("玩家%s 充值USDT, value:%s, Decimals:%d, 实际到账USDT:%f, 兑换游戏币:%d,玩家旧余额:%d,新余额:%d, 兑换券:%d,玩家旧余额:%d,新余额:%d",
	//	account, value, decimals, usdtAmount, retGameCoins, oldBalance, userInfo.Balance, retVoucher, oldVoucher, userInfo.Voucher)

	// coinFlow埋点
	log.InfoJson("金币入口",
		zap.String("Account", account),
		zap.String("ActionType", log.Flow),
		zap.String("FlowType", log.CoinFlow),
		zap.String("From", log.FromRecharge),
		zap.String("Flag", log.FlagIn),
		zap.Int64("Amount", retGameCoins),
		zap.Int64("Old", oldBalance),
		zap.Int64("New", userInfo.Balance),
		zap.Int64("CreatedAt", time.Now().UnixMilli()),
	)
	// VoucherFlow埋点
	log.InfoJson("凭证入口",
		zap.String("Account", account),
		zap.String("ActionType", log.Flow),
		zap.String("FlowType", log.VoucherFlow),
		zap.String("From", log.FromRecharge),
		zap.String("Flag", log.FlagIn),
		zap.Int64("Amount", retVoucher),
		zap.Int64("Old", oldVoucher),
		zap.Int64("New", userInfo.Voucher),
		zap.Int64("CreatedAt", time.Now().UnixMilli()),
	)

	return int64(usdtAmount * 100), retGameCoins, retVoucher, nil
}
func updateUsdtAgent(userInfo *presenter.UserInfo, contribution int64) {
	today := time.Now().Format("2006-01-02")
	if userInfo.UsdtAgent.LastDate != today {
		userInfo.UsdtAgent.DailyAgent = 0
		userInfo.UsdtAgent.LastDate = today
	}
	userInfo.UsdtAgent.DailyAgent += contribution
	userInfo.UsdtAgent.TotalAgent += contribution
	userInfo.UsdtAgent.Unclaimed += contribution
}

func UsdtToSqu(userInfo *presenter.UserInfo, amount int64) error {
	if !CheckUSDT(userInfo, amount) {
		return errors.New("USDT not enough")
	}
	oldBalance := userInfo.Balance
	oldUSDT := userInfo.USDT
	oldVoucher := userInfo.Voucher

	DecrUSDT(userInfo, amount)
	allocation := int64(utils.LubanTables.TBApp.Get("allocation").NumInt)   // USDT 转换 游戏币汇率 (1美元 : 1000000游戏币)
	allocatioon := int64(utils.LubanTables.TBApp.Get("allocatioon").NumInt) // USDT 转换 兑换券汇率 (1美元 : 10兑换券)
	transferCoin := amount * allocation / 100                               // 计算汇率美元为单位
	transferVoucher := amount * allocatioon / 100
	AddAmount(userInfo, transferCoin)
	AddVoucher(userInfo, transferVoucher)
	if err := Update(context.Background(), userInfo, nil); err != nil {
		log.Error("Update userInfo error: ", err)
		return err
	}

	exchangeInfo := &rechargedb.ExchangeInfo{
		Account:   userInfo.Account,
		Type:      rechargedb.UsdtToSqu,
		Amount:    amount,
		Coin:      transferCoin,
		Voucher:   transferVoucher,
		CreatedAt: time.Now().UnixMilli(),
	}
	if err := Insert(context.Background(), exchangeInfo); err != nil {
		log.Error(err)
		return err
	}

	fund := &rechargedb.Fund{}
	if err := Find(context.Background(), fund, 0); err != nil {
		log.Error(err)
		return err
	}
	fund.ExchangeCoins += transferCoin
	if e := Update(context.Background(), fund, nil); e != nil {
		log.Error("Error updating recharge fund:", e)
	}

	//log.Infof("用户:%s, 美分==>游戏币/兑换券, 旧美分:%d,消耗了:%d,新美分:%d, 旧游戏币:%d,兑换了:%d,新游戏币:%d, 旧券:%d,兑换了:%d,新券:%d",
	//	userInfo.Account, oldUSDT, amount, userInfo.USDT, oldBalance, transferCoin, userInfo.Balance, oldVoucher, transferVoucher, userInfo.Voucher)

	// 美分==>游戏币/兑换券
	// UsdtFlow埋点
	log.InfoJson("USDT出口",
		zap.String("Account", userInfo.Account),
		zap.String("ActionType", log.Flow),
		zap.String("FlowType", log.UsdtFlow),
		zap.String("From", log.FromUsdtToSqu),
		zap.String("Flag", log.FlagOut),
		zap.Int64("Amount", amount),     //消耗了
		zap.Int64("Old", oldUSDT),       //旧美分
		zap.Int64("New", userInfo.USDT), //新美分
		zap.Int64("CreatedAt", time.Now().UnixMilli()),
	)
	// coinFlow埋点
	log.InfoJson("金币入口",
		zap.String("Account", userInfo.Account),
		zap.String("ActionType", log.Flow),
		zap.String("FlowType", log.CoinFlow),
		zap.String("From", log.FromUsdtToSqu),
		zap.String("Flag", log.FlagIn),
		zap.Int64("Amount", transferCoin),  //兑换了
		zap.Int64("Old", oldBalance),       //旧游戏币
		zap.Int64("New", userInfo.Balance), //新游戏币
		zap.Int64("CreatedAt", time.Now().UnixMilli()),
	)
	// VoucherFlow埋点
	log.InfoJson("凭证入口",
		zap.String("Account", userInfo.Account),
		zap.String("ActionType", log.Flow),
		zap.String("FlowType", log.VoucherFlow),
		zap.String("From", log.FromUsdtToSqu),
		zap.String("Flag", log.FlagIn),
		zap.Int64("Amount", transferVoucher), //兑换了
		zap.Int64("Old", oldVoucher),         //旧券
		zap.Int64("New", userInfo.Voucher),   //新券
		zap.Int64("CreatedAt", time.Now().UnixMilli()),
	)
	return nil
}

func AddUsdtRecharge(userInfo *presenter.UserInfo, value int64) {
	today := time.Now().Format("2006-01-02")
	if today != userInfo.UsdtRecharge.LastDate {
		userInfo.UsdtRecharge.DailyRecharge = 0
		userInfo.UsdtRecharge.LastDate = today
	}
	userInfo.UsdtRecharge.DailyRecharge += value
	userInfo.UsdtRecharge.TotalRecharge += value
}
func updateDownLineRecharge(userInfo *presenter.UserInfo, account string) {
	today := time.Now().Format("2006-01-02")
	if today != userInfo.UsdtRecharge.LastDate {
		userInfo.UsdtRecharge.DownLineDailyRecharge = make(map[string]bool)
		userInfo.UsdtRecharge.LastDate = today
	}
	// 记录今日下线充值账户，每个账户每天只记录一次
	if _, exists := userInfo.UsdtRecharge.DownLineDailyRecharge[account]; !exists {
		userInfo.UsdtRecharge.DownLineDailyRecharge[account] = true
	}
	// 记录总下线充值账户，每个账户只记录一次
	if _, exists := userInfo.UsdtRecharge.DownLineTotalRecharge[account]; !exists {
		userInfo.UsdtRecharge.DownLineTotalRecharge[account] = true
	}
}

func FindAllRechargeOrder(ctx context.Context, account string) ([]*rechargedb.PayInfo, error) {
	var payInfo rechargedb.PayInfo
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(payInfo.TableName())
	filter := bson.M{"account": account}
	opts := options.Find().SetSort(bson.D{{"created_at", -1}}) // 根据创建时间降序排序

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var payInfos []*rechargedb.PayInfo
	if err = cursor.All(ctx, &payInfos); err != nil {
		return nil, err
	}
	return payInfos, nil
}

func FindAllWithdrawOrder(ctx context.Context, account string) ([]*rechargedb.WithDrawInfo, error) {
	var withDrawInfo rechargedb.WithDrawInfo
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(withDrawInfo.TableName())
	filter := bson.M{"account": account}
	opts := options.Find().SetSort(bson.D{{"created_at", -1}}) // 根据创建时间降序排序

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var withDrawInfos []*rechargedb.WithDrawInfo
	if err = cursor.All(ctx, &withDrawInfos); err != nil {
		return nil, err
	}
	// 检查每个订单的TxHash字段
	for _, info := range withDrawInfos {
		if info.TxHash == "" {
			txHash := wallet.WithdrawHistoryById(info.OrderId)
			log.Infof("wallet.HistoryById, orderId:%v, txHash:%v", info.OrderId, txHash)
			if txHash != "" && txHash != wallet.ErrNoTransactionFound && txHash != wallet.ErrFailedToRetrieve {
				info.TxHash = txHash
				update := bson.M{"$set": bson.M{"txHash": txHash}}
				if err := Update(ctx, info, update); err != nil {
					log.Error(err)
				}
			}
		}
	}
	return withDrawInfos, nil
}

func FindAllExchangeOrder(ctx context.Context, account string) ([]*rechargedb.ExchangeInfo, error) {
	var exchangeInfo rechargedb.ExchangeInfo
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(exchangeInfo.TableName())
	filter := bson.M{"account": account}
	opts := options.Find().SetSort(bson.D{{"created_at", -1}}) // 根据创建时间降序排序

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var exchangeInfos []*rechargedb.ExchangeInfo
	if err = cursor.All(ctx, &exchangeInfos); err != nil {
		return nil, err
	}
	return exchangeInfos, nil
}

func FindLabels() ([]web.Label, error) {
	label := &web.Label{}
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(label.TableName())
	// 创建一个空的查询条件，表示查询所有文档
	filter := bson.D{{}}
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.TODO())
	var labels []web.Label
	for cur.Next(context.TODO()) {
		var webLabel web.Label
		if err := cur.Decode(&webLabel); err != nil {
			return nil, err
		}
		labels = append(labels, webLabel)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return labels, nil
}

func ExistAdv(adv string) (bool, error) {
	label := &web.Label{}
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(label.TableName())
	filter := bson.D{{"adv", adv}}
	err := collection.FindOne(context.Background(), filter).Decode(label)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func GetBatchCdkNextSeq(key string) (int64, error) {
	filter := bson.M{"_id": key}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var updatedDoc struct {
		Seq int64 `bson:"seq"`
	}
	label := &presenter.BatchCdk{}
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(label.TableName())
	err := collection.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&updatedDoc)
	if err != nil {
		return 0, err
	}
	return updatedDoc.Seq, nil
}

// ----------------------------------------------用于核对账单------------------------------------------------------------
// 有多少用户(不包含机器人)
func CountUsers(ctx context.Context) (int64, error) {
	var userInfo presenter.UserInfo
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(userInfo.TableName())
	filter := bson.M{"isRobot": false}
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// 所有用户的余额之和
func SumBalances(ctx context.Context) (int64, error) {
	var userInfo presenter.UserInfo
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(userInfo.TableName())
	pipeline := mongo.Pipeline{
		{{"$group", bson.M{"_id": nil, "totalBalance": bson.M{"$sum": "$balance"}}}},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	var results []struct {
		TotalBalance int64 `bson:"totalBalance"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return 0, err
	}
	if len(results) > 0 {
		return results[0].TotalBalance, nil
	}
	return 0, nil
}

// 所有机器人的余额之和
func SumRobotBalances(ctx context.Context) (int64, error) {
	var userInfo presenter.UserInfo
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(userInfo.TableName())
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{"isRobot": true}}}, // 只选择机器人用户
		{{"$group", bson.M{"_id": nil, "totalBalance": bson.M{"$sum": "$balance"}}}},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	var results []struct {
		TotalBalance int64 `bson:"totalBalance"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return 0, err
	}
	if len(results) > 0 {
		return results[0].TotalBalance, nil
	}
	return 0, nil
}

// 所有用户已领取cdk之和
func SumCdk(ctx context.Context) (int64, error) {
	var cdkInfo presenter.Cdk
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(cdkInfo.TableName())

	pipeline := mongo.Pipeline{
		{{"$match", bson.M{"Received": true}}},
		{{"$group", bson.M{"_id": nil, "totalBalance": bson.M{"$sum": "$exchangeAmount"}}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}

	var results []struct {
		TotalBalance int64 `bson:"totalBalance"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return 0, err
	}
	if len(results) > 0 {
		return results[0].TotalBalance, nil
	}
	return 0, nil
}

// 所有用户邀请奖励之和
func SumInvite(ctx context.Context) (int64, error) {
	var userInfo presenter.UserInfo
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(userInfo.TableName())
	cursor, err := collection.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"completedTasks": 1}))
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var sum int64
	for cursor.Next(ctx) {
		var user presenter.UserInfo
		if err := cursor.Decode(&user); err != nil {
			log.Error("Error decoding user:", err)
			continue
		}
		for _, inviteID := range user.CompletedTasks {
			reward := int64(utils.LubanTables.TBApp.Get(inviteID).NumInt)
			sum += reward
		}
	}
	if err := cursor.Err(); err != nil {
		return 0, err
	}
	return sum, nil
}

// 所有用户福利金总额
func SumTotalWelfare(ctx context.Context) (int64, error) {
	var userInfo presenter.UserInfo
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(userInfo.TableName())
	pipeline := mongo.Pipeline{
		{
			{"$group", bson.M{
				"_id":                nil,
				"totalWelfareAmount": bson.M{"$sum": "$welfare.totalAmount"},
			}},
		},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Error("Error executing aggregate query for total welfare amount: ", err)
		return 0, err
	}
	defer cursor.Close(ctx)
	var results []struct {
		TotalWelfareAmount int64 `bson:"totalWelfareAmount"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		log.Error("Error getting aggregate results for total welfare amount: ", err)
		return 0, err
	}
	if len(results) > 0 {
		return results[0].TotalWelfareAmount, nil
	}
	return 0, nil
}

// 所有用户充值兑换游戏币总额
func SumTotalGameCoins(ctx context.Context) (int64, error) {
	var payInfo rechargedb.PayInfo
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(payInfo.TableName())
	pipeline := mongo.Pipeline{
		{
			{"$group", bson.M{
				"_id":            nil,
				"totalGameCoins": bson.M{"$sum": "$gameCoins"},
			}},
		},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Errorf("Error executing aggregate query for total game coins: %v", err)
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		TotalGameCoins int64 `bson:"totalGameCoins"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		log.Errorf("Error getting aggregate results for total game coins: %v", err)
		return 0, err
	}
	if len(results) > 0 {
		return results[0].TotalGameCoins, nil
	}
	return 0, nil
}

// 所有用户木头人的首通之和
func SumFirstPassPools(ctx context.Context) (int64, error) {
	var userInfo presenter.UserInfo
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(userInfo.TableName())
	pipeline := mongo.Pipeline{
		{{"$group", bson.M{"_id": nil, "totalFirstPassPool": bson.M{"$sum": "$squid.first_pass.pool"}}}},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	var results []struct {
		TotalFirstPassPool int64 `bson:"totalFirstPassPool"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return 0, err
	}
	if len(results) > 0 {
		return results[0].TotalFirstPassPool, nil
	}
	return 0, nil
}

// 所有用户的未领取代理之和
func SumUnclaimedAgents(ctx context.Context) (int64, error) {
	var userInfo presenter.UserInfo
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(userInfo.TableName())
	pipeline := mongo.Pipeline{
		{{"$group", bson.M{"_id": nil, "totalUnclaimed": bson.M{"$sum": bson.M{"$add": []interface{}{"$agent.squid.unclaimed", "$agent.glass.unclaimed", "$agent.compete.unclaimed"}}}}}},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	var results []struct {
		TotalUnclaimed int64 `bson:"totalUnclaimed"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return 0, err
	}
	if len(results) > 0 {
		return results[0].TotalUnclaimed, nil
	}
	return 0, nil
}

// 所有用户本轮下注之和(还没到结算):木头人, 玻璃桥, 拔河
func SumSquidBets(ctx context.Context) (int64, error) {
	var userInfo presenter.UserInfo
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(userInfo.TableName())
	pipeline := mongo.Pipeline{
		{{"$group", bson.M{"_id": nil, "totalSquidBets": bson.M{"$sum": "$squid.bet_prices"}}}},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	var results []struct {
		TotalSquidBets int64 `bson:"totalSquidBets"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return 0, err
	}
	if len(results) > 0 {
		return results[0].TotalSquidBets, nil
	}
	return 0, nil
}
func SumCompeteBets(ctx context.Context) (int64, error) {
	var userInfo presenter.UserInfo
	collection := mongoInstance.Database(viper.GetString("common.mongodb.database")).Collection(userInfo.TableName())
	pipeline := mongo.Pipeline{
		{{"$group", bson.M{"_id": nil, "totalCompeteBets": bson.M{"$sum": "$competeLastBet"}}}},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	var results []struct {
		TotalCompeteBets int64 `bson:"totalCompeteBets"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return 0, err
	}
	if len(results) > 0 {
		return results[0].TotalCompeteBets, nil
	}
	return 0, nil
}

const SquidType = "木头人"
const GlassType = "玻璃桥"
const LadderType = "梯子"
const CompeteType = "拔河"

func Check(gameType string) {
	userNum, err := CountUsers(context.Background())
	if err != nil {
		log.Error("CountUsers error: ", err)
	}
	robotNums := int64(utils.LubanTables.TBWood.Get("robot_num").NumInt)
	robotCoin := int64(utils.LubanTables.TBWood.Get("robot_coin").NumInt)
	squidFund := &squid.Fund{}
	if err := Find(context.Background(), squidFund, 0); err != nil {
		log.Error(err)
	}
	//glassFund := &glass.Fund{}
	//if err := Find(context.Background(), glassFund, 0); err != nil {
	//	log.Error(err)
	//}
	ladderFund := &ladder.Fund{}
	if err := Find(context.Background(), ladderFund, 0); err != nil {
		log.Error(err)
	}
	game := &compete.Game{}
	if e := Find(context.Background(), game, 0); e != nil {
		log.Error(err)
	}
	rechargeFund := &rechargedb.Fund{}
	if err := Find(context.Background(), rechargeFund, 0); err != nil {
		log.Error(err)
	}

	// 用户初始资金之和
	initMoney := userNum*utils.InitBalance + robotNums*robotCoin

	// 用户当前余额之和
	balances, err := SumBalances(context.Background())
	if err != nil {
		log.Error("SumBalances error: ", err)
	}

	// 所有用户已领取cdk之和
	cdk, err := SumCdk(context.Background())
	if err != nil {
		log.Error("SumCdk error: ", err)
	}

	// 所有用户邀请奖励之和
	invite, err := SumInvite(context.Background())
	if err != nil {
		log.Error("SumInvite error: ", err)
	}

	// 所有用户扶贫金之和
	welfare, err := SumTotalWelfare(context.Background())
	if err != nil {
		log.Error("SumTotalWelfare error: ", err)
	}

	// 所有充值兑换之和
	rechargeGameCoins, err := SumTotalGameCoins(context.Background())
	if err != nil {
		log.Error("SumTotalGameCoins error: ", err)
	}

	// 用户本轮下注之和(还没到结算)
	sumSquidBets, err := SumSquidBets(context.Background())
	if err != nil {
		log.Error("SumSquidBets error: ", err)
	}
	sumCompeteBets, err := SumCompeteBets(context.Background())
	if err != nil {
		log.Error("SumCompeteBets error: ", err)
	}
	bets := sumSquidBets + sumCompeteBets

	// 用户未领取代理之和
	unclaimedAgents, err := SumUnclaimedAgents(context.Background())
	if err != nil {
		log.Error("SumUnclaimedAgents error: ", err)
	}

	// 用户木头人首通池之和
	firstPassPools, err := SumFirstPassPools(context.Background())
	if err != nil {
		log.Error("SumFirstPassPools error: ", err)
	}
	// 木头人玩家抽水
	houseCutSquid := squidFund.HouseCut
	// 木头人机器人抽水
	robotPool := squidFund.RobotPool
	// 木头人jackpot
	jackpot := squidFund.Jackpot

	//// 玻璃桥玩家抽水
	//houseCutGlass := glassFund.HouseCut
	//// 玻璃桥可赔付库存
	//availableFundGlass := glassFund.AvailableFund

	// 梯子玩家抽水
	houseCutLadder := ladderFund.HouseCut
	// 梯子可赔付库存
	availableFundLadder := ladderFund.AvailableFund

	// 拔河玩家抽水
	houseCutCompete := game.TakeAmount
	// 拔河可赔付库存
	availableFundCompete := game.Amount

	// 闪兑游戏币
	exchangeCoins := rechargeFund.ExchangeCoins

	loss := (initMoney + invite + cdk + welfare + rechargeGameCoins + exchangeCoins - balances) - (bets + unclaimedAgents + firstPassPools + houseCutSquid + robotPool + jackpot + houseCutLadder + availableFundLadder + houseCutCompete + availableFundCompete)
	if loss == 0 {
		log.Infof("%s核对总账单成功,误差为:%d, 初始资金%d, 邀请奖励%d, 已领取cdk:%d, 福利金:%d, 余额:%d, 用户本轮未结算下注额之和:%d, 未领取代理%d, 首通池:%d, 木头人玩家抽水:%d, 木头人机器人抽水:%d, 木头人jackpot:%d, 梯子抽水:%d, 梯子可赔付库存:%d, 拔河抽水:%d, 拔河可赔付库存:%d",
			gameType, loss, initMoney, invite, cdk, welfare, balances, bets, unclaimedAgents, firstPassPools, houseCutSquid, robotPool, jackpot, houseCutLadder, availableFundLadder, houseCutCompete, availableFundCompete)
	} else {
		log.Errorf("%s核对总账失败,误差为:%d, 初始资金%d, 邀请奖励%d, 已领取cdk:%d, 福利金:%d, 余额:%d, 用户本轮未结算下注额之和:%d, 未领取代理%d, 首通池:%d, 木头人玩家抽水:%d, 木头人机器人抽水:%d, 木头人jackpot:%d, 梯子抽水:%d, 梯子可赔付库存:%d,  拔河抽水:%d, 拔河可赔付库存:%d",
			gameType, loss, initMoney, invite, cdk, welfare, balances, bets, unclaimedAgents, firstPassPools, houseCutSquid, robotPool, jackpot, houseCutLadder, availableFundLadder, houseCutCompete, availableFundCompete)
	}
}
