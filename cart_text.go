# 设置基础 URL
$baseUrl = "http://localhost:8888"

# 设置请求头，添加UTF-8编码
$headers = @{
    "Content-Type" = "application/json; charset=utf-8"
    "Accept" = "application/json"
}

# 确保PowerShell使用UTF-8编码
$PSDefaultParameterValues['Out-File:Encoding'] = 'utf8'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# 测试用户ID
$userID = "user456"

# 1. 添加商品到购物车
$addToCartBody = @{
    user_id = $userID
    currency = "CNY"
    item = @{
        product_id = "prod456"
        quantity = 3
        price = 199.99
        product_name = "高级耳机"
        image_url = "https://example.com/headphones.jpg"
    }
} | ConvertTo-Json -Depth 10

Write-Host "正在添加商品到购物车..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/carts" `
                                -Method Post `
                                -Body ([System.Text.Encoding]::UTF8.GetBytes($addToCartBody)) `
                                -Headers $headers
    
    Write-Host "商品已添加到购物车！" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 10
} catch {
    Write-Host "添加商品到购物车失败: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# 2. 获取购物车信息
Write-Host "`n正在获取购物车信息..." -ForegroundColor Yellow
try {
    $cartInfo = Invoke-RestMethod -Uri "$baseUrl/carts/$userID" `
                                -Method Get `
                                -Headers $headers
    
    Write-Host "购物车信息获取成功！" -ForegroundColor Green
    $cartInfo | ConvertTo-Json -Depth 10
} catch {
    Write-Host "获取购物车信息失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 3. 添加另一个商品到购物车
$addSecondItemBody = @{
    user_id = $userID
    currency = "CNY"
    item = @{
        product_id = "prod789"
        quantity = 1
        price = 499.99
        product_name = "智能手表"
        image_url = "https://example.com/smartwatch.jpg"
    }
} | ConvertTo-Json -Depth 10

Write-Host "`n正在添加第二个商品到购物车..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/carts" `
                                -Method Post `
                                -Body ([System.Text.Encoding]::UTF8.GetBytes($addSecondItemBody)) `
                                -Headers $headers
    
    Write-Host "第二个商品已添加到购物车！" -ForegroundColor Green
} catch {
    Write-Host "添加第二个商品到购物车失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 4. 更新购物车商品数量
$updateQuantityBody = @{
    quantity = 5
} | ConvertTo-Json

Write-Host "`n正在更新购物车商品数量..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/carts/$userID/items/prod456" `
                                -Method Put `
                                -Body ([System.Text.Encoding]::UTF8.GetBytes($updateQuantityBody)) `
                                -Headers $headers
    
    Write-Host "商品数量已更新！" -ForegroundColor Green
    $response | ConvertTo-Json
} catch {
    Write-Host "更新商品数量失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 5. 再次获取购物车信息
Write-Host "`n正在获取更新后的购物车信息..." -ForegroundColor Yellow
try {
    $updatedCartInfo = Invoke-RestMethod -Uri "$baseUrl/carts/$userID" `
                                      -Method Get `
                                      -Headers $headers
    
    Write-Host "更新后的购物车信息获取成功！" -ForegroundColor Green
    $updatedCartInfo | ConvertTo-Json -Depth 10
} catch {
    Write-Host "获取更新后的购物车信息失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 6. 结算购物车（转换为订单）
$checkoutBody = @{
    address = @{
        street = "测试大道123号"
        city = "北京市"
        district = "海淀区"
        country = "中国"
        postal_code = "100081"
    }
    email = "test_user@example.com"
} | ConvertTo-Json -Depth 10

Write-Host "`n正在结算购物车..." -ForegroundColor Yellow
try {
    $checkoutResponse = Invoke-RestMethod -Uri "$baseUrl/carts/$userID/checkout" `
                                        -Method Post `
                