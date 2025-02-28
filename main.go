package main

import (
	"bytedance/config"
	"bytedance/model"
	"bytedance/service"
	"context"
	"log"
	"os"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func main() {
	// 使用配置服务连接数据库
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "root"
	}
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "127.0.0.1"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "dingdan"
	}

	db, err := config.InitDB(dbUser, dbPassword, dbHost, dbPort, dbName)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 自动迁移数据库表
	err = db.AutoMigrate(&model.Order{}, &model.Cart{})
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 初始化服务
	orderService := service.NewOrderService(db)
	cartService := service.NewCartService(db)

	// 启动订单过期检查任务
	go orderService.StartOrderExpirationChecker()

	// 创建Hertz服务器
	h := server.Default()

	// ===== 订单相关路由 =====
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
			c.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.JSON(consts.StatusOK, result)
	})

	// 更新订单
	h.PUT("/orders/:orderID", func(ctx context.Context, c *app.RequestContext) {
		orderID := c.Param("orderID")
		log.Printf("收到更新订单请求，订单ID: %s", orderID)
		
		var updates map[string]interface{}
		body := c.Request.Body()
		log.Printf("收到的更新请求体: %s", string(body))
		
		if err := c.Bind(&updates); err != nil {
			log.Printf("请求数据绑定失败: %v", err)
			c.JSON(consts.StatusBadRequest, map[string]interface{}{
				"error":   "无效的请求数据",
				"details": err.Error(),
			})
			return
		}
		
		if err := orderService.UpdateOrder(ctx, orderID, updates); err != nil {
			log.Printf("更新订单失败: %v", err)
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

	// 获取订单详情
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

	// ===== 购物车相关路由 =====
	// 添加商品到购物车
	h.POST("/carts", func(ctx context.Context, c *app.RequestContext) {
		var req service.AddToCartRequest
		body := c.Request.Body()
		log.Printf("收到添加购物车请求体: %s", string(body))
		
		if err := c.Bind(&req); err != nil {
			log.Printf("请求数据绑定失败: %v", err)
			c.JSON(consts.StatusBadRequest, map[string]interface{}{
				"error":   "无效的请求数据",
				"details": err.Error(),
			})
			return
		}
		
		result, err := cartService.CreateOrUpdateCart(ctx, &req)
		if err != nil {
			c.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		
		c.JSON(consts.StatusOK, result)
	})

	// 获取购物车信息
	h.GET("/carts/:userID", func(ctx context.Context, c *app.RequestContext) {
		userID := c.Param("userID")
		cart, err := cartService.GetCart(ctx, userID)
		if err != nil {
			c.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.JSON(consts.StatusOK, cart)
	})

	// 清空购物车
	h.DELETE("/carts/:userID", func(ctx context.Context, c *app.RequestContext) {
		userID := c.Param("userID")
		err := cartService.ClearCart(ctx, userID)
		if err != nil {
			c.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		c.JSON(consts.StatusOK, map[string]interface{}{
			"message": "购物车已清空",
			"userID":  userID,
		})
	})

	// 从购物车移除商品
	h.DELETE("/carts/:userID/items/:productID", func(ctx context.Context, c *app.RequestContext) {
		userID := c.Param("userID")
		productID := c.Param("productID")
		
		err := cartService.RemoveItemFromCart(ctx, userID, productID)
		if err != nil {
			c.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		
		c.JSON(consts.StatusOK, map[string]interface{}{
			"message":    "商品已从购物车中移除",
			"userID":     userID,
			"productID":  productID,
		})
	})

	// 更新购物车商品数量
	h.PUT("/carts/:userID/items/:productID", func(ctx context.Context, c *app.RequestContext) {
		userID := c.Param("userID")
		productID := c.Param("productID")
		
		var req struct {
			Quantity int `json:"quantity"`
		}
		
		if err := c.Bind(&req); err != nil {
			c.JSON(consts.StatusBadRequest, map[string]interface{}{
				"error": "无效的请求数据",
			})
			return
		}
		
		err := cartService.UpdateCartItemQuantity(ctx, userID, productID, req.Quantity)
		if err != nil {
			c.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		
		c.JSON(consts.StatusOK, map[string]interface{}{
			"message":   "购物车商品数量已更新",
			"userID":    userID,
			"productID": productID,
			"quantity":  req.Quantity,
		})
	})

	// 将购物车转换为订单
	h.POST("/carts/:userID/checkout", func(ctx context.Context, c *app.RequestContext) {
		userID := c.Param("userID")
		
		var req struct {
			Address map[string]interface{} `json:"address"`
			Email   string                 `json:"email"`
		}
		
		if err := c.Bind(&req); err != nil {
			c.JSON(consts.StatusBadRequest, map[string]interface{}{
				"error": "无效的请求数据",
			})
			return
		}
		
		result, err := cartService.ConvertCartToOrder(ctx, userID, req.Address, req.Email)
		if err != nil {
			c.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		
		c.JSON(consts.StatusOK, map[string]interface{}{
			"message":  "购物车已成功转换为订单",
			"order_id": result.OrderID,
		})
	})

	// 测试路由
	h.GET("/hello", func(ctx context.Context, c *app.RequestContext) {
		c.String(consts.StatusOK, "Hello hertz!")
	})

	// 启动服务器
	h.Spin()
}
