package main

import (
	"Distributed-fileserver/config"
	"Distributed-fileserver/mq"
	"Distributed-fileserver/service/upload/customLog"
	"fmt"
	"go.uber.org/zap"
	"os"
	"time"

	"github.com/micro/cli"
	"github.com/micro/go-micro"
	_ "github.com/micro/go-plugins/registry/consul"
	_ "github.com/micro/go-plugins/registry/kubernetes"

	"Distributed-fileserver/common"
	dbproxy "Distributed-fileserver/service/dbproxy/client"
	cfg "Distributed-fileserver/service/upload/config"
	upProto "Distributed-fileserver/service/upload/proto"
	"Distributed-fileserver/service/upload/route"
	upRpc "Distributed-fileserver/service/upload/rpc"
)

func startRPCService() {
	service := micro.NewService(
		micro.Name("go.micro.service.upload"), // 服务名称
		micro.RegisterTTL(time.Second*10),     // TTL指定从上一次心跳间隔起，超过这个时间服务会被服务发现移除
		micro.RegisterInterval(time.Second*5), // 让服务在指定时间内重新注册，保持TTL获取的注册时间有效
		micro.Flags(common.CustomFlags...),
	)
	service.Init(
		micro.Action(func(c *cli.Context) {
			// 检查是否指定mqhost
			mqhost := c.String("mqhost")
			if len(mqhost) > 0 {
				customLog.Logger.Info(fmt.Sprintf("upload main custom mq address: %s" + mqhost))
				mq.UpdateRabbitHost(mqhost)
			}
		}),
	)

	// 初始化dbproxy client
	dbproxy.Init(service)
	// 初始化mq client
	mq.Init()

	err := upProto.RegisterUploadServiceHandler(service.Server(), new(upRpc.Upload))
	if err != nil{
		customLog.Logger.Error("upload main RegisterUploadServiceHandler error", zap.Error(err))
	}
	if err = service.Run(); err != nil {
		customLog.Logger.Error("upload main service run error", zap.Error(err))
	}
}

func startAPIService() {
	router := route.Router()
	err := router.Run(cfg.UploadServiceHost)
	if err != nil{
		customLog.Logger.Error("upload main startAPIService error", zap.Error(err))
	}
	// service := web.NewService(
	// 	web.Name("go.micro.web.upload"),
	// 	web.Handler(router),
	// 	web.RegisterTTL(10*time.Second),
	// 	web.RegisterInterval(5*time.Second),
	// )
	// if err := service.Init(); err != nil {
	// 	log.Fatal(err)
	// }

	// if err := service.Run(); err != nil {
	// 	log.Fatal(err)
	// }
}

func main() {
	err := os.MkdirAll(config.TempLocalRootDir, 0777)
	if err != nil{
		customLog.Logger.Error("upload main mkdir error", zap.Error(err))
	}
	err = os.MkdirAll(config.TempPartRootDir, 0777)
	if err != nil{
		customLog.Logger.Error("upload main mkdir error", zap.Error(err))
	}

	// api 服务
	go startAPIService()

	// rpc 服务
	startRPCService()
}
