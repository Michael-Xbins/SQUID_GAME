package log

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"time"
)

var (
	debugLogger *zap.Logger
	infoLogger  *zap.Logger
	warnLogger  *zap.Logger
	errorLogger *zap.Logger
	fatalLogger *zap.Logger
	panicLogger *zap.Logger
)

// 埋点数据
var (
	Register              = "register"
	Login                 = "login"
	Recharge              = "recharge"
	Flow                  = "flow"
	CoinFlow              = "coinFlow"
	VoucherFlow           = "voucherFlow"
	UsdtFlow              = "usdtFlow"
	FlagIn                = "in"
	FlagOut               = "out"
	FromRecharge          = "充值"
	FromAgent             = "代理" // 领取 下/下下线 充值代理贡献
	FromWelfare           = "每日扶贫"
	DailyTask             = "每日任务"
	FromUsdtToSqu         = "USDT闪兑SQU"
	FromSquToUsdt         = "SQU闪兑USDT"
	FromWithdraw          = "提现"
	FromAuction           = "拍卖行"
	FromSquidSettlement   = "鱿鱼结算"
	FromCompeteSettlement = "拔河结算"
	FromLadderSettlement  = "梯子结算"
	FromFirstPass         = "鱿鱼每日首通"
	FromJackpot           = "鱿鱼jackpot"
	CDK                   = "兑换cdk"
)

func InitConfig() {
	logDir := "log"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.Mkdir(logDir, 0755)
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	level := zapcore.InfoLevel
	switch viper.GetString("common.log.logLevel") {
	case "Debug":
		level = zapcore.DebugLevel
	case "Info":
		level = zapcore.InfoLevel
	case "Warn":
		level = zapcore.WarnLevel
	case "Error":
		level = zapcore.ErrorLevel
	case "Panic":
		level = zapcore.PanicLevel
	case "Fatal":
		level = zapcore.FatalLevel
	}

	// --------------------------------------- 输出控制台 ---------------------------------------
	//consoleCore := zapcore.NewCore(
	//	zapcore.NewConsoleEncoder(encoderConfig),
	//	zapcore.AddSync(os.Stdout),
	//	zap.NewAtomicLevelAt(level),
	//)
	//logger := zap.New(consoleCore, zap.AddCaller(), zap.AddCallerSkip(1))
	//debugLogger = logger
	//infoLogger = logger
	//warnLogger = logger
	//errorLogger = logger
	//fatalLogger = logger
	//panicLogger = logger

	// ---------------------------------------- 写入文件 ----------------------------------------
	core := func(filename string) zapcore.Core {
		lumberjackLog := &lumberjack.Logger{
			Filename: filepath.Join(logDir, filename), // 日志文件的完整路径
			MaxSize:  10,                              // 日志文件达到10MB后将进行分割
			MaxAge:   60,                              // 保留分割文件的最长天数为60天
			//MaxBackups: 5,                               // 只保留最新的5个分割文件
			//Compress:   true,                            // 分割的旧文件将被压缩
		}
		return zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(lumberjackLog),
			zap.NewAtomicLevelAt(level),
		)
	}
	debugLogger = zap.New(core("debug.log"), zap.AddCaller(), zap.AddCallerSkip(1))
	infoLogger = zap.New(core("info.log"), zap.AddCaller(), zap.AddCallerSkip(1))
	warnLogger = zap.New(core("warn.log"), zap.AddCaller(), zap.AddCallerSkip(1))
	errorLogger = zap.New(core("error.log"), zap.AddCaller(), zap.AddCallerSkip(1))
	panicLogger = zap.New(core("panic.log"), zap.AddCaller(), zap.AddCallerSkip(1))
	fatalLogger = zap.New(core("fatal.log"), zap.AddCaller(), zap.AddCallerSkip(1))
}
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func Debug(msg ...interface{}) {
	debugLogger.Sugar().Debug(msg)
}
func Info(msg ...interface{}) {
	infoLogger.Sugar().Info(msg)
}
func Warn(msg ...interface{}) {
	warnLogger.Sugar().Warn(msg)
}
func Error(msg ...interface{}) {
	errorLogger.Sugar().Error(msg)
}
func Panic(msg ...interface{}) {
	panicLogger.Sugar().Panic(msg)
}
func Fatal(msg ...interface{}) {
	fatalLogger.Sugar().Fatal(msg)
}
func Debugf(msg string, args ...interface{}) {
	debugLogger.Sugar().Debugf(msg, args...)
}
func Infof(msg string, args ...interface{}) {
	infoLogger.Sugar().Infof(msg, args...)
}
func Warnf(msg string, args ...interface{}) {
	warnLogger.Sugar().Warnf(msg, args...)
}
func Errorf(msg string, args ...interface{}) {
	errorLogger.Sugar().Errorf(msg, args...)
}
func Panicf(msg string, args ...interface{}) {
	panicLogger.Sugar().Panicf(msg, args...)
}
func Fatalf(msg string, args ...interface{}) {
	fatalLogger.Sugar().Fatalf(msg, args...)
}

// json形式日志
func DebugJson(msg string, fields ...zap.Field) {
	debugLogger.Debug(msg, fields...)
}
func InfoJson(msg string, fields ...zap.Field) {
	infoLogger.Info(msg, fields...)
}
func WarnJson(msg string, fields ...zap.Field) {
	infoLogger.Warn(msg, fields...)
}
func ErrorJson(msg string, fields ...zap.Field) {
	infoLogger.Error(msg, fields...)
}
func PanicJson(msg string, fields ...zap.Field) {
	infoLogger.Panic(msg, fields...)
}
func FatalJson(msg string, fields ...zap.Field) {
	infoLogger.Fatal(msg, fields...)
}
