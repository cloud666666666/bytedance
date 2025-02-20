# 设置基础 URL
$baseUrl = "http://localhost:8888"

# 准备创建订单的请求体
$createOrderBody = @{
    user_id = "user123"
    user_currency = "CNY"
    address = @{
        street = "测试街道"
        city = "测试城市"
        country = "中国"
    }
    email = "test@example.com"
    order_items = @(
        @{
            product_id = "prod123"
            quantity = 2
            price = 99.99
        }
    )
} | ConvertTo-Json -Depth 10  # 添加 -Depth 参数确保嵌套对象被正确序列化

# 设置请求头，添加UTF-8编码
$headers = @{
    "Content-Type" = "application/json; charset=utf-8"
    "Accept" = "application/json"
}

# 确保PowerShell使用UTF-8编码
$PSDefaultParameterValues['Out-File:Encoding'] = 'utf8'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# 发送创建订单请求
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/orders" `
                                -Method Post `
                                -Body ([System.Text.Encoding]::UTF8.GetBytes($createOrderBody)) `
                                -Headers $headers `
                                -ContentType "application/json; charset=utf-8"
    
    Write-Host "订单创建成功！订单ID: $($response.order_id)" -ForegroundColor Green
    Write-Host "请求体: $createOrderBody" -ForegroundColor Yellow
} catch {
    Write-Host "创建订单失败: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "状态码: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Red
    Write-Host "请求体: $createOrderBody" -ForegroundColor Yellow
    exit 1
}
