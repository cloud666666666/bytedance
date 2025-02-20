package service

import (
	"bytedance/model"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderService struct {
	db *gorm.DB
}

type PlaceOrderReq struct {
	UserID       string                 `json:"user_id"`
	UserCurrency string                 `json:"user_currency"`
	Address      map[string]interface{} `json:"address"`
	Email        string                 `json:"email"`
	OrderItems   []map[string]interface{} `json:"order_items"`
}

type OrderResult struct {
	OrderID string
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{db: db}
}

// 创建订单
func (s *OrderService) CreateOrder(ctx context.Context, req *PlaceOrderReq) (*OrderResult, error) {
	// 添加日志
	fmt.Printf("收到的请求数据: %+v\n", req)

	orderID := uuid.New().String()

	// 检查必要字段
	if req.UserID == "" || req.UserCurrency == "" {
		return nil, fmt.Errorf("用户ID和货币类型不能为空")
	}

	// 检查 OrderItems 是否为空
	if req.OrderItems == nil {
		return nil, fmt.Errorf("订单项不能为空")
	}

	addressBytes, err := json.MarshalIndent(req.Address, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化地址失败: %v", err)
	}

	orderItemsBytes, err := json.MarshalIndent(req.OrderItems, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化订单项失败: %v", err)
	}

	// 打印序列化后的数据
	fmt.Printf("地址JSON: %s\n", string(addressBytes))
	fmt.Printf("订单项JSON: %s\n", string(orderItemsBytes))

	order := &model.Order{
		OrderID:    orderID,
		UserID:     req.UserID,
		Currency:   req.UserCurrency,
		Address:    string(addressBytes),
		Email:      req.Email,
		OrderItems: string(orderItemsBytes),
		Status:     "pending",
		CreatedAt:  time.Now(),
		ExpireAt:   time.Now().Add(30 * time.Minute),
	}

	// 打印最终要存储的订单数据
	fmt.Printf("要存储的订单数据: %+v\n", order)

	// 开启事务
	tx := s.db.Begin()
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建订单失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %v", err)
	}

	// 启动异步任务处理订单过期
	go s.handleOrderExpiration(orderID)

	return &OrderResult{
		OrderID: orderID,
	}, nil
}

// 修改订单信息
func (s *OrderService) UpdateOrder(ctx context.Context, orderID string, updates map[string]interface{}) error {
	// 检查并处理特殊字段
	if address, ok := updates["address"].(map[string]interface{}); ok {
		addressBytes, err := json.Marshal(address)
		if err != nil {
			return fmt.Errorf("序列化地址失败: %v", err)
		}
		updates["address"] = string(addressBytes)
	}

	if orderItems, ok := updates["order_items"].(map[string]interface{}); ok {
		orderItemsBytes, err := json.Marshal(orderItems)
		if err != nil {
			return fmt.Errorf("序列化订单项失败: %v", err)
		}
		updates["order_items"] = string(orderItemsBytes)
	}

	// 更新 updated_at 字段
	updates["updated_at"] = time.Now()

	result := s.db.Model(&model.Order{}).
		Where("order_id = ? AND status = ?", orderID, "pending").
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("更新订单失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("订单不存在或已不能修改")
	}

	return nil
}

// 处理订单过期
func (s *OrderService) handleOrderExpiration(orderID string) {
	// 创建定时器
	timer := time.NewTimer(30 * time.Minute)
	defer timer.Stop()

	<-timer.C

	// 检查订单状态并取消未支付订单
	err := s.db.Model(&model.Order{}).
		Where("order_id = ? AND status = ? AND expire_at <= ?",
			orderID, "pending", time.Now()).
		Update("status", "cancelled").Error

	if err != nil {
		// 记录错误日志
		fmt.Printf("取消订单失败: %v\n", err)
	}
}

// 获取订单详情
func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*model.Order, error) {
	var order model.Order
	if err := s.db.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return nil, fmt.Errorf("获取订单失败: %v", err)
	}

	// 解码存储的 JSON 字符串
	var address, orderItems interface{}
	if err := json.Unmarshal([]byte(order.Address), &address); err != nil {
		return nil, fmt.Errorf("解析地址失败: %v", err)
	}
	if err := json.Unmarshal([]byte(order.OrderItems), &orderItems); err != nil {
		return nil, fmt.Errorf("解析订单项失败: %v", err)
	}

	return &order, nil
}
