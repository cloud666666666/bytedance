package main

import (
	"bytedance/model"
	"bytedance/service"
	"context"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func main() {
	// 连接数据库
	dsn := "root:root@tcp(127.0.0.1:3306)/dingdan?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 自动迁移数据库表
	err = db.AutoMigrate(&model.Order{})
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 初始化订单服务
	orderService := service.NewOrderService(db)

	// server.Default() creates a Hertz with recovery middleware.
	// If you need a pure hertz, you can use server.New()
	h := server.Default()

	// 添加创建订单的路由
	h.POST("/orders", func(ctx context.Context, c *app.RequestContext) {
		var req service.PlaceOrderReq

		// 打印接收到的原始数据
		body := c.Request.Body()
		log.Printf("收到的请求体: %s", string(body))

		// 添加更详细的绑定日志
		if err := c.Bind(&req); err != nil {
			log.Printf("绑定错误: %v", err)
			log.Printf("请求结构体内容: %+v", req)
			c.JSON(consts.StatusBadRequest, map[string]interface{}{
				"error":    err.Error(),
				"received": string(body),
				"message":  "请求数据格式错误",
			})
			return
		}

		// 打印绑定后的数据
		log.Printf("绑定后的请求数据: %+v", req)

		result, err := orderService.CreateOrder(ctx, &req)
		if err != nil {
			c.JSON(consts.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(consts.StatusOK, result)
	})

	// 更新订单 - 修改路由格式
	h.PUT("/orders/:orderID", func(ctx context.Context, c *app.RequestContext) {
		orderID := c.Param("orderID")            // 获取路由参数
		log.Printf("收到更新订单请求，订单ID: %s", orderID) // 添加日志

		var updates map[string]interface{}

		// 打印接收到的原始数据
		body := c.Request.Body()
		log.Printf("收到的更新请求体: %s", string(body))

		if err := c.Bind(&updates); err != nil {
			log.Printf("请求数据绑定失败: %v", err) // 添加错误日志
			c.JSON(consts.StatusBadRequest, map[string]interface{}{
				"error":   "无效的请求数据",
				"details": err.Error(),
			})
			return
		}

		if err := orderService.UpdateOrder(ctx, orderID, updates); err != nil {
			log.Printf("更新订单失败: %v", err) // 添加错误日志
			c.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		c.JSON(consts.StatusOK, map[string]interface{}{
			"message": "订单更新成功",
			"orderID": orderID,
		})
	})

	// 需要添加获取订单详情的路由
	h.GET("/orders/:orderID", func(ctx context.Context, c *app.RequestContext) {
		orderID := c.Param("orderID")
		order, err := orderService.GetOrder(ctx, orderID)
		if err != nil {
			c.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.JSON(consts.StatusOK, order)
	})

	h.GET("/hello", func(ctx context.Context, c *app.RequestContext) {
		c.String(consts.StatusOK, "Hello hertz!")
	})

	h.Spin()
}
