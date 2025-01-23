package auction

import (
	"application/api/presenter"
	"application/api/presenter/auction"
	"application/mongodb"
	pb "application/pkg/proto/danmu/message"
	"application/pkg/utils"
	"application/pkg/utils/log"
	"application/session"
	"context"
	"go.uber.org/zap"
	"time"
)

func StartAuctionGame() {
	if e := mongodb.Find(context.Background(), game, 0); e != nil {
		log.Panic(e)
	}
	if game.ClosePrice == 0 {
		game.ClosePrice = int64(utils.LubanTables.TBApp.Get("firstprice").NumInt)
	}
	game.MinutesOfPrice = 0
	game.MinutesOfCount = 0
	utils.LoopSafeGo(EventLoop)
}

var game = &auction.Game{}

var orderChanged = false

var isClose = false

var fchan = make(chan func(), 256)

func EventLoop() {
	startDate := time.Now()
	day := startDate.Day()
	minute := startDate.Minute()
	secondTicker := time.NewTicker(time.Second)
	defer secondTicker.Stop()

	var tradePrice = int64(0)
	for {
		if checkOpenTime() == nil {
			isClose = false
			currentDate := time.Now()
			var tradeSucc = false
			for {
				startTime := time.Now() // 开始计时
				var err error
				var buyOrder *auction.BuyOrderList
				var sellOrder *auction.SellOrderList
				buyOrder, err = mongodb.GetBuyOrderWithMaxPrice(context.Background())
				if err != nil {
					break
				}
				sellOrder, err = mongodb.GetSellOrderWithMaxPrice(context.Background())
				if err != nil {
					break
				}

				// 先按价格撮合
				if buyOrder.Price < sellOrder.Price {
					break
				}
				buyUser := &presenter.UserInfo{}
				err = mongodb.Find(context.Background(), buyUser, buyOrder.Account)
				if err != nil {
					break
				}

				oldBuyVoucher := buyUser.Voucher
				tradePrice = sellOrder.Price
				buyCount := buyOrder.TotalCount - buyOrder.CompletedCount
				sellCount := sellOrder.TotalCount - sellOrder.CompletedCount
				var cnt int64
				if buyCount >= sellCount {
					cnt = sellCount
				} else {
					cnt = buyCount
				}

				// 再按时间撮合
				var finalPrice int64
				refundAmount := (buyOrder.Price - sellOrder.Price) * cnt
				if buyOrder.Timestamp > sellOrder.Timestamp {
					// 卖单先挂出，按卖的价格成交
					finalPrice = sellOrder.Price
				} else {
					// 买单先挂出，按买的价格成交
					finalPrice = buyOrder.Price
				}

				// 扣除卖家手续费
				tradeAmount := finalPrice * cnt
				takePercent := int64(utils.LubanTables.TBApp.Get("market").NumInt) //手续费
				takeAmount := tradeAmount * takePercent / 1000
				completedAmount := tradeAmount - takeAmount

				if buyOrder.Timestamp > sellOrder.Timestamp {
					// 卖单先挂出，按卖的价格成交, 买家多出的钱 给买家
					buyUser.Balance += refundAmount
				}
				buyUser.Voucher += cnt
				err = mongodb.Update(context.Background(), buyUser, nil)
				if err != nil {
					break
				}
				log.InfoJson("金币出口", // coinFlow埋点
					zap.String("Account", buyUser.Account),
					zap.String("ActionType", log.Flow),
					zap.String("FlowType", log.CoinFlow),
					zap.String("From", log.FromAuction),
					zap.String("Flag", log.FlagOut),
					zap.Int64("Amount", tradeAmount),
					zap.Int64("Old", buyUser.Balance+tradeAmount),
					zap.Int64("New", buyUser.Balance),
					zap.Int64("CreatedAt", time.Now().UnixMilli()),
				)
				// VoucherFlow埋点
				log.InfoJson("凭证入口",
					zap.String("Account", buyUser.Account),
					zap.String("ActionType", log.Flow),
					zap.String("FlowType", log.VoucherFlow),
					zap.String("From", log.FromAuction),
					zap.String("Flag", log.FlagIn),
					zap.Int64("Amount", cnt),
					zap.Int64("Old", oldBuyVoucher),
					zap.Int64("New", buyUser.Voucher),
					zap.Int64("CreatedAt", time.Now().UnixMilli()),
				)

				sellUser := &presenter.UserInfo{}
				err = mongodb.Find(context.Background(), sellUser, sellOrder.Account)
				if err != nil {
					break
				}
				oldSellBalance := sellUser.Balance
				sellUser.Balance += completedAmount
				err = mongodb.Update(context.Background(), sellUser, nil)
				if err != nil {
					break
				}
				log.InfoJson("金币入口", // coinFlow埋点
					zap.String("Account", sellUser.Account),
					zap.String("ActionType", log.Flow),
					zap.String("FlowType", log.CoinFlow),
					zap.String("From", log.FromAuction),
					zap.String("Flag", log.FlagIn),
					zap.Int64("Amount", completedAmount),
					zap.Int64("Old", oldSellBalance),
					zap.Int64("New", sellUser.Balance),
					zap.Int64("CreatedAt", time.Now().UnixMilli()),
				)
				// VoucherFlow埋点
				log.InfoJson("凭证出口",
					zap.String("Account", sellUser.Account),
					zap.String("ActionType", log.Flow),
					zap.String("FlowType", log.VoucherFlow),
					zap.String("From", log.FromAuction),
					zap.String("Flag", log.FlagOut),
					zap.Int64("Amount", cnt),
					zap.Int64("Old", sellUser.Voucher+cnt),
					zap.Int64("New", sellUser.Voucher),
					zap.Int64("CreatedAt", time.Now().UnixMilli()),
				)

				game.MinutesOfPrice += tradeAmount
				game.MinutesOfCount += cnt
				buyOrder.CompletedCount += cnt
				sellOrder.CompletedCount += cnt
				mongodb.Insert(context.Background(), &auction.HistoryList{
					OrderId:        buyOrder.Id,
					Type:           "buy",
					Account:        buyOrder.Account,
					Timestamp:      time.Now().UnixMilli(),
					Price:          finalPrice,
					CompletedCount: cnt,
					Amount:         finalPrice * cnt,
				})
				mongodb.Insert(context.Background(), &auction.HistoryList{
					OrderId:        sellOrder.Id,
					Type:           "sell",
					Account:        sellOrder.Account,
					Timestamp:      time.Now().UnixMilli(),
					Price:          finalPrice,
					CompletedCount: cnt,
					Amount:         finalPrice * cnt,
				})

				if buyOrder.CompletedCount == buyOrder.TotalCount {
					mongodb.Delete(context.Background(), buyOrder)
				} else {
					mongodb.Update(context.Background(), buyOrder, nil)
				}
				if sellOrder.CompletedCount == sellOrder.TotalCount {
					mongodb.Delete(context.Background(), sellOrder)
				} else {
					mongodb.Update(context.Background(), sellOrder, nil)
				}

				session.S2CMessage(buyUser.Account, &pb.ClientResponse{
					Type:    pb.MessageType_BuySuccessNotify_,
					Message: &pb.ClientResponse_BuySuccessNotify{BuySuccessNotify: &pb.BuySuccessNotify{Count: cnt, Price: finalPrice, Balance: buyUser.Balance, Voucher: buyUser.Voucher}},
				})
				session.S2CMessage(sellUser.Account, &pb.ClientResponse{
					Type:    pb.MessageType_SellSuccessNotify_,
					Message: &pb.ClientResponse_SellSuccessNotify{SellSuccessNotify: &pb.SellSuccessNotify{Count: cnt, Price: finalPrice, Balance: sellUser.Balance, Voucher: sellUser.Voucher}},
				})
				tradeSucc = true
				log.Infof("撮合成功, 耗时:%v, buyOrderId:%s, sellOrderId:%s, ", time.Since(startTime), buyOrder.Id, sellOrder.Id)
			}
			if tradeSucc || orderChanged {
				orderChanged = false
				sellListTop5, _ := mongodb.GetTop5SellOrderList(context.Background())
				session.S2AllMessage(&pb.ClientResponse{
					Type:    pb.MessageType_SellOrderListNotify_,
					Message: &pb.ClientResponse_SellOrderListNotify{SellOrderListNotify: &pb.SellOrderListNotify{SellOrderInfo: sellListTop5}},
				})
				buyListTop5, _ := mongodb.GetTop5BuyOrderList(context.Background())
				session.S2AllMessage(&pb.ClientResponse{
					Type:    pb.MessageType_BuyOrderListNotify_,
					Message: &pb.ClientResponse_BuyOrderListNotify{BuyOrderListNotify: &pb.BuyOrderListNotify{BuyOrderInfo: buyListTop5}},
				})
			}
			if minute != currentDate.Minute() {
				minute = currentDate.Minute()
				avgPrice := int64(0)
				if game.MinutesOfCount > 0 {
					avgPrice = game.MinutesOfPrice / (game.MinutesOfCount)
				}
				game.TradeRecordList = append(game.TradeRecordList, auction.TradeRecord{
					Timestamp: currentDate.UnixMilli(),
					AvgPrice:  avgPrice,
				})
				game.MinutesOfPrice = 0
				game.MinutesOfCount = 0
				session.S2AllMessage(&pb.ClientResponse{
					Type: pb.MessageType_TradeRecordNotify_,
					Message: &pb.ClientResponse_TradeRecordNotify{
						TradeRecordNotify: &pb.TradeRecordNotify{
							Timestamp: currentDate.UnixMilli(),
							AvgPrice:  avgPrice,
						}},
				})
				mongodb.Update(context.Background(), game, nil)
			}
			if day != currentDate.Day() {
				if tradePrice > 0 {
					game.ClosePrice = tradePrice
				}
				mongodb.DeleteOldOrders(context.Background(), 30) // 记录保存30天
				game.TradeRecordList = nil
				day = currentDate.Day()
				mongodb.Update(context.Background(), game, nil)
				session.S2AllMessage(&pb.ClientResponse{
					Type: pb.MessageType_OpenMarketNotify_,
					Message: &pb.ClientResponse_OpenMarketNotify{OpenMarketNotify: &pb.OpenMarketNotify{
						ClosePrice: game.ClosePrice,
					}},
				})
			}
		} else {
			if err := cancelOrder(); err != nil {
				log.Error(err)
			}
		}
		select {
		case <-secondTicker.C:
		case f := <-fchan:
			f()
		}
	}

}

func cancelOrder() error {
	if isClose == false {
		//撤购买单
		buyOrderList, err := mongodb.FindAllBuyOrders()
		if err != nil {
			log.Error(err)
			return err
		} else {
			for _, order := range buyOrderList {
				err = mongodb.Delete(context.Background(), &order)
				if err != nil {
					log.Error(err)
				}
				cnt := order.TotalCount - order.CompletedCount
				if cnt > 0 {
					userInfo := &presenter.UserInfo{}
					_ = mongodb.Find(context.Background(), userInfo, order.Account)
					restitutionAmount := cnt * order.Price
					userInfo.Balance += restitutionAmount
					err = mongodb.Update(context.Background(), userInfo, nil)
					if err != nil {
						log.Error(err)
					}
				}
			}
		}

		//撤出售单
		sellOrderList, err := mongodb.FindAllSellOrders()
		if err != nil {
			log.Error(err)
			return err
		} else {
			for _, order := range sellOrderList {
				err = mongodb.Delete(context.Background(), &order)
				if err != nil {
					log.Error(err)
				}
				cnt := order.TotalCount - order.CompletedCount
				userInfo := &presenter.UserInfo{}
				_ = mongodb.Find(context.Background(), userInfo, order.Account)
				userInfo.Voucher += cnt
				err = mongodb.Update(context.Background(), userInfo, nil)
				if err != nil {
					log.Error(err)
				}
			}
		}

		session.S2AllMessage(&pb.ClientResponse{
			Type:    pb.MessageType_CloseMarketNotify_,
			Message: &pb.ClientResponse_CloseMarketNotify{CloseMarketNotify: &pb.CloseMarketNotify{}},
		})
		isClose = true
		game.MinutesOfPrice = 0
		game.MinutesOfCount = 0
	}
	return nil
}
