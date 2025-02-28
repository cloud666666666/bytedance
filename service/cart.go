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

// CartItem 购物车项目
type CartItem struct {
	ProductID   string  `json:"product_id"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	ProductName string  `json:"product_name"`
	ImageURL    string  `json:"image_url,omitempty"`
}

// CartService 购物车服务
type CartService struct {
	db *gorm.DB
}

// CartResponse 购物车响应
type CartResponse struct {
	CartID   string     `json:"cart_id"`
	UserID   string     `json:"user_id"`
	Items    []CartItem `json:"items"`
	Currency string     `json:"currency"`
	Total    float64    `json:"total"`
}

// AddToCartRequest 添加到购物车的请求
type AddToCartRequest struct {
	UserID    string   `json:"user_id"`
	Currency  string   `json:"currency"`
	CartItems CartItem `json:"item"`
}

// NewCartService 创建新的购物车服务
func NewCartService(db *gorm.DB) *CartService {
	// 确保数据库表已迁移
	db.AutoMigrate(&model.Cart{})
	return &CartService{db: db}
}

// CreateOrUpdateCart 创建或更新购物车
func (s *CartService) CreateOrUpdateCart(ctx context.Context, req *AddToCartRequest) (*CartResponse, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	if req.Currency == "" {
		return nil, fmt.Errorf("货币类型不能为空")
	}

	if req.CartItems.ProductID == "" || req.CartItems.Quantity <= 0 {
		return nil, fmt.Errorf("商品信息不完整或数量无效")
	}

	// 检查用户是否已有购物车
	var cart model.Cart
	result := s.db.Where("user_id = ?", req.UserID).First(&cart)

	var items []CartItem
	var cartID string

	// 如果用户已有购物车，则更新
	if result.Error == nil {
		cartID = cart.CartID
		// 解析现有购物车项目
		if err := json.Unmarshal([]byte(cart.Items), &items); err != nil {
			return nil, fmt.Errorf("解析购物车项目失败: %v", err)
		}

		// 检查商品是否已在购物车中
		found := false
		for i, item := range items {
			if item.ProductID == req.CartItems.ProductID {
				// 更新商品数量
				items[i].Quantity += req.CartItems.Quantity
				found = true
				break
			}
		}

		// 如果商品不在购物车中，则添加
		if !found {
			items = append(items, req.CartItems)
		}
	} else {
		// 创建新购物车
		cartID = uuid.New().String()
		items = []CartItem{req.CartItems}
	}

	// 序列化购物车项目
	itemsBytes, err := json.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("序列化购物车项目失败: %v", err)
	}

	now := time.Now()

	// 开始事务
	tx := s.db.Begin()

	if result.Error == nil {
		// 更新现有购物车
		if err := tx.Model(&model.Cart{}).
			Where("cart_id = ?", cartID).
			Updates(map[string]interface{}{
				"items":      string(itemsBytes),
				"currency":   req.Currency,
				"updated_at": now,
			}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("更新购物车失败: %v", err)
		}
	} else {
		// 创建新购物车
		newCart := &model.Cart{
			CartID:    cartID,
			UserID:    req.UserID,
			Items:     string(itemsBytes),
			Currency:  req.Currency,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := tx.Create(newCart).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("创建购物车失败: %v", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %v", err)
	}

	// 计算总价
	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}

	return &CartResponse{
		CartID:   cartID,
		UserID:   req.UserID,
		Items:    items,
		Currency: req.Currency,
		Total:    total,
	}, nil
}

// GetCart 获取购物车信息
func (s *CartService) GetCart(ctx context.Context, userID string) (*CartResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	var cart model.Cart
	if err := s.db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("未找到用户的购物车")
		}
		return nil, fmt.Errorf("获取购物车失败: %v", err)
	}

	var items []CartItem
	if err := json.Unmarshal([]byte(cart.Items), &items); err != nil {
		return nil, fmt.Errorf("解析购物车项目失败: %v", err)
	}

	// 计算总价
	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}

	return &CartResponse{
		CartID:   cart.CartID,
		UserID:   cart.UserID,
		Items:    items,
		Currency: cart.Currency,
		Total:    total,
	}, nil
}

// ClearCart 清空购物车
func (s *CartService) ClearCart(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("用户ID不能为空")
	}

	// 检查用户是否有购物车
	var cart model.Cart
	result := s.db.Where("user_id = ?", userID).First(&cart)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return fmt.Errorf("未找到用户的购物车")
		}
		return fmt.Errorf("查询购物车失败: %v", result.Error)
	}

	// 清空购物车（更新为空数组）
	emptyItems, _ := json.Marshal([]CartItem{})
	now := time.Now()

	if err := s.db.Model(&model.Cart{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"items":      string(emptyItems),
			"updated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("清空购物车失败: %v", err)
	}

	return nil
}

// RemoveItemFromCart 从购物车中移除商品
func (s *CartService) RemoveItemFromCart(ctx context.Context, userID string, productID string) error {
	if userID == "" || productID == "" {
		return fmt.Errorf("用户ID和商品ID不能为空")
	}

	// 获取用户购物车
	var cart model.Cart
	if err := s.db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		return fmt.Errorf("获取购物车失败: %v", err)
	}

	// 解析购物车项目
	var items []CartItem
	if err := json.Unmarshal([]byte(cart.Items), &items); err != nil {
		return fmt.Errorf("解析购物车项目失败: %v", err)
	}

	// 移除指定商品
	newItems := []CartItem{}
	for _, item := range items {
		if item.ProductID != productID {
			newItems = append(newItems, item)
		}
	}

	// 序列化新购物车项目
	newItemsBytes, err := json.Marshal(newItems)
	if err != nil {
		return fmt.Errorf("序列化购物车项目失败: %v", err)
	}

	// 更新购物车
	now := time.Now()
	if err := s.db.Model(&model.Cart{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"items":      string(newItemsBytes),
			"updated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("更新购物车失败: %v", err)
	}

	return nil
}

// UpdateCartItemQuantity 更新购物车商品数量
func (s *CartService) UpdateCartItemQuantity(ctx context.Context, userID string, productID string, quantity int) error {
	if userID == "" || productID == "" {
		return fmt.Errorf("用户ID和商品ID不能为空")
	}

	if quantity <= 0 {
		return fmt.Errorf("商品数量必须大于0")
	}

	// 获取用户购物车
	var cart model.Cart
	if err := s.db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		return fmt.Errorf("获取购物车失败: %v", err)
	}

	// 解析购物车项目
	var items []CartItem
	if err := json.Unmarshal([]byte(cart.Items), &items); err != nil {
		return fmt.Errorf("解析购物车项目失败: %v", err)
	}

	// 更新指定商品数量
	found := false
	for i, item := range items {
		if item.ProductID == productID {
			items[i].Quantity = quantity
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("购物车中未找到指定商品")
	}

	// 序列化新购物车项目
	itemsBytes, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("序列化购物车项目失败: %v", err)
	}

	// 更新购物车
	now := time.Now()
	if err := s.db.Model(&model.Cart{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"items":      string(itemsBytes),
			"updated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("更新购物车失败: %v", err)
	}

	return nil
}

// ConvertCartToOrder 将购物车转换为订单
func (s *CartService) ConvertCartToOrder(ctx context.Context, userID string, address map[string]interface{}, email string) (*OrderResult, error) {
	// 获取购物车
	cartResponse, err := s.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取购物车失败: %v", err)
	}

	if len(cartResponse.Items) == 0 {
		return nil, fmt.Errorf("购物车为空，无法创建订单")
	}

	// 转换为订单项
	var orderItems []map[string]interface{}
	for _, item := range cartResponse.Items {
		orderItem := map[string]interface{}{
			"product_id":   item.ProductID,
			"quantity":     item.Quantity,
			"price":        item.Price,
			"product_name": item.ProductName,
		}
		if item.ImageURL != "" {
			orderItem["image_url"] = item.ImageURL
		}
		orderItems = append(orderItems, orderItem)
	}

	// 创建订单请求
	orderReq := &PlaceOrderReq{
		UserID:       userID,
		UserCurrency: cartResponse.Currency,
		Address:      address,
		Email:        email,
		OrderItems:   orderItems,
	}

	// 创建订单服务实例
	orderService := NewOrderService(s.db)

	// 创建订单
	orderResult, err := orderService.CreateOrder(ctx, orderReq)
	if err != nil {
		return nil, fmt.Errorf("创建订单失败: %v", err)
	}

	// 清空购物车
	if err := s.ClearCart(ctx, userID); err != nil {
		fmt.Printf("警告：订单已创建但清空购物车失败: %v\n", err)
	}

	return orderResult, nil
}