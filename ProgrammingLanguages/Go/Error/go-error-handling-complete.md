# Go 错误处理完全指南

> 从零基础到面试通关，一份文档搞定 Go 错误处理

---

## 📖 学习路线图

```
第一阶段：基础入门（1-2天）
├── 1. 理解 error 接口
├── 2. 基本错误处理模式
└── 3. 创建和返回错误

第二阶段：进阶应用（2-3天）
├── 4. 错误包装与链式处理
├── 5. 自定义错误类型
└── 6. panic/recover 机制

第三阶段：实战提升（3-5天）
├── 7. 错误处理最佳实践
├── 8. 常见场景解决方案
└── 9. 性能优化技巧

第四阶段：面试冲刺（1-2天）
├── 10. 手撕代码 10 题
├── 11. 面试高频考点
└── 12. 实战项目案例
```

---

## 第一阶段：基础入门

### 1.1 error 接口本质

#### 📚 核心知识

```go
// error 是 Go 内置接口
type error interface {
    Error() string
}
```

**关键要点：**
1. error 是接口，任何实现 `Error() string` 的类型都是 error
2. nil error 表示没有错误
3. error 通过多返回值传递

#### 💡 最简单的例子

```go
package main

import (
    "errors"
    "fmt"
)

func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

func main() {
    // 正确用法
    result, err := divide(10, 2)
    if err != nil {
        fmt.Println("错误:", err)
        return
    }
    fmt.Println("结果:", result)

    // 错误用法
    result, err = divide(10, 0)
    if err != nil {
        fmt.Println("错误:", err) // 输出: 错误: division by zero
        return
    }
}
```

#### 🎯 练习题 1：实现简单的错误处理

```go
// 任务：实现一个函数，读取配置文件
// 要求：
// 1. 文件不存在返回错误
// 2. 文件为空返回错误
// 3. 返回文件内容

func readConfig(filename string) (string, error) {
    // TODO: 在这里实现
}

// 测试用例
func TestReadConfig(t *testing.T) {
    // 测试文件不存在
    _, err := readConfig("notexist.txt")
    if err == nil {
        t.Error("应该返回错误")
    }

    // 测试正常读取
    content, err := readConfig("config.txt")
    if err != nil {
        t.Error("不应该返回错误")
    }
    if content == "" {
        t.Error("内容不应为空")
    }
}
```

<details>
<summary>💡 参考答案</summary>

```go
import (
    "errors"
    "os"
)

func readConfig(filename string) (string, error) {
    // 检查文件是否存在
    if _, err := os.Stat(filename); os.IsNotExist(err) {
        return "", errors.New("配置文件不存在")
    }

    // 读取文件内容
    data, err := os.ReadFile(filename)
    if err != nil {
        return "", err
    }

    // 检查文件是否为空
    if len(data) == 0 {
        return "", errors.New("配置文件为空")
    }

    return string(data), nil
}
```
</details>

---

### 1.2 错误处理三种方式

#### 📚 核心知识

```go
// 方式1: errors.New - 创建简单错误
err := errors.New("something went wrong")

// 方式2: fmt.Errorf - 格式化错误
err := fmt.Errorf("failed to process user %d", userID)

// 方式3: 自定义错误类型
type MyError struct {
    Code int
    Msg  string
}

func (e *MyError) Error() string {
    return fmt.Sprintf("error %d: %s", e.Code, e.Msg)
}
```

#### 💡 对比示例

```go
package main

import (
    "errors"
    "fmt"
)

// 方式1: 简单错误（适合固定消息）
var ErrNotFound = errors.New("user not found")

func findUserSimple(id int) error {
    if id == 0 {
        return ErrNotFound
    }
    return nil
}

// 方式2: 格式化错误（适合动态消息）
func findUserFormatted(id int) error {
    if id == 0 {
        return fmt.Errorf("user %d not found", id)
    }
    return nil
}

// 方式3: 自定义类型（适合需要附加信息）
type UserError struct {
    UserID int
    Reason string
}

func (e *UserError) Error() string {
    return fmt.Sprintf("user error [%d]: %s", e.UserID, e.Reason)
}

func findUserCustom(id int) error {
    if id == 0 {
        return &UserError{
            UserID: id,
            Reason: "user not found in database",
        }
    }
    return nil
}
```

#### 🎯 练习题 2：选择合适的错误类型

```go
// 任务：实现一个验证用户输入的函数
// 要求：
// 1. 用户名为空 -> 返回固定错误
// 2. 密码长度不足 -> 返回包含最小长度的错误
// 3. 邮箱格式错误 -> 返回自定义错误类型，包含字段名和错误原因

type ValidationError struct {
    // TODO: 定义字段
}

func (e *ValidationError) Error() string {
    // TODO: 实现
}

func validateUser(username, password, email string) error {
    // TODO: 实现
}
```

<details>
<summary>💡 参考答案</summary>

```go
import (
    "errors"
    "fmt"
    "strings"
)

var ErrEmptyUsername = errors.New("用户名不能为空")

const MinPasswordLength = 8

type ValidationError struct {
    Field  string
    Reason string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("验证失败 [%s]: %s", e.Field, e.Reason)
}

func validateUser(username, password, email string) error {
    // 用户名验证
    if username == "" {
        return ErrEmptyUsername
    }

    // 密码验证
    if len(password) < MinPasswordLength {
        return fmt.Errorf("密码长度不足，至少需要 %d 个字符", MinPasswordLength)
    }

    // 邮箱验证
    if !strings.Contains(email, "@") {
        return &ValidationError{
            Field:  "email",
            Reason: "邮箱格式错误，缺少 @ 符号",
        }
    }

    return nil
}
```
</details>

---

### 1.3 错误处理的黄金法则

#### 📚 核心知识

**法则1：永远检查错误**
```go
// ❌ 错误示范
file, _ := os.Open("file.txt")

// ✅ 正确示范
file, err := os.Open("file.txt")
if err != nil {
    return err
}
defer file.Close()
```

**法则2：错误只处理一次**
```go
// ❌ 错误示范：既记录又返回
func processData() error {
    err := doSomething()
    if err != nil {
        log.Printf("error: %v", err) // 记录
        return err                    // 又返回
    }
    return nil
}

// ✅ 正确示范：只返回，让上层决定
func processData() error {
    if err := doSomething(); err != nil {
        return fmt.Errorf("process data: %w", err)
    }
    return nil
}

// 在最顶层处理
func main() {
    if err := processData(); err != nil {
        log.Printf("ERROR: %v", err) // 只记录一次
    }
}
```

**法则3：为调用者添加上下文**
```go
// ❌ 错误示范：直接返回原始错误
func loadConfig() error {
    _, err := os.ReadFile("config.json")
    if err != nil {
        return err // 不知道是哪个文件
    }
    return nil
}

// ✅ 正确示范：添加上下文
func loadConfig() error {
    _, err := os.ReadFile("config.json")
    if err != nil {
        return fmt.Errorf("load config: %w", err)
    }
    return nil
}
```

#### 🎯 练习题 3：修复错误代码

```go
// 以下代码违反了哪些黄金法则？请修复

func processOrder(orderID string) (*Order, error) {
    // 查询订单
    order, _ := db.Query("SELECT * FROM orders WHERE id = ?", orderID)

    // 验证库存
    if err := checkInventory(order.ProductID); err != nil {
        log.Printf("库存不足: %v", err)
        return nil, err
    }

    // 扣减库存
    err := reduceInventory(order.ProductID, order.Quantity)
    if err != nil {
        return nil, err
    }

    return order, nil
}
```

<details>
<summary>💡 参考答案</summary>

```go
func processOrder(orderID string) (*Order, error) {
    // 修复1：检查数据库查询错误
    order, err := db.Query("SELECT * FROM orders WHERE id = ?", orderID)
    if err != nil {
        return nil, fmt.Errorf("query order %s: %w", orderID, err)
    }

    // 修复2：不要既记录又返回，只返回即可
    if err := checkInventory(order.ProductID); err != nil {
        return nil, fmt.Errorf("check inventory for product %s: %w",
            order.ProductID, err)
    }

    // 修复3：添加上下文信息
    if err := reduceInventory(order.ProductID, order.Quantity); err != nil {
        return nil, fmt.Errorf("reduce inventory for product %s (qty: %d): %w",
            order.ProductID, order.Quantity, err)
    }

    return order, nil
}

// 在最顶层统一记录日志
func handleOrder(w http.ResponseWriter, r *http.Request) {
    orderID := r.URL.Query().Get("id")
    order, err := processOrder(orderID)
    if err != nil {
        log.Printf("ERROR: process order failed: %v", err) // 统一记录
        http.Error(w, "Internal Server Error", 500)
        return
    }

    json.NewEncoder(w).Encode(order)
}
```
</details>

---

## 第二阶段：进阶应用

### 2.1 错误包装与链式处理（Go 1.13+）

#### 📚 核心知识

**错误包装的三个关键函数：**

```go
// 1. fmt.Errorf 配合 %w 包装错误
err := fmt.Errorf("outer: %w", innerErr)

// 2. errors.Is 检查错误链
if errors.Is(err, os.ErrNotExist) {
    // 处理文件不存在
}

// 3. errors.As 类型断言错误链
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    fmt.Println("Path:", pathErr.Path)
}

// 4. errors.Unwrap 解包错误
unwrapped := errors.Unwrap(err)
```

#### 💡 完整示例

```go
package main

import (
    "errors"
    "fmt"
    "os"
)

// 场景：多层函数调用
func readUserConfig(userID int) ([]byte, error) {
    filename := fmt.Sprintf("config_%d.json", userID)
    data, err := os.ReadFile(filename)
    if err != nil {
        // 使用 %w 包装，保留原始错误
        return nil, fmt.Errorf("read user config: %w", err)
    }
    return data, nil
}

func loadUser(userID int) (*User, error) {
    data, err := readUserConfig(userID)
    if err != nil {
        // 再次包装
        return nil, fmt.Errorf("load user %d: %w", userID, err)
    }

    user := &User{}
    // 解析配置...
    return user, nil
}

func main() {
    user, err := loadUser(123)
    if err != nil {
        // 错误链：load user 123: read user config: open config_123.json: no such file or directory
        fmt.Println("完整错误:", err)

        // 检查是否是文件不存在错误（会遍历整个错误链）
        if errors.Is(err, os.ErrNotExist) {
            fmt.Println("配置文件不存在，使用默认配置")
            return
        }

        // 获取具体的 PathError
        var pathErr *os.PathError
        if errors.As(err, &pathErr) {
            fmt.Printf("操作: %s, 路径: %s\n", pathErr.Op, pathErr.Path)
        }

        return
    }

    fmt.Println("用户加载成功:", user)
}
```

#### 🔥 关键对比：%v vs %w

```go
// ❌ 使用 %v - 丢失错误链
err1 := errors.New("original error")
err2 := fmt.Errorf("wrapped: %v", err1)

fmt.Println(errors.Is(err2, err1))        // false，错误链断了
fmt.Println(errors.Unwrap(err2) == nil)   // true，无法解包

// ✅ 使用 %w - 保留错误链
err1 := errors.New("original error")
err2 := fmt.Errorf("wrapped: %w", err1)

fmt.Println(errors.Is(err2, err1))        // true，错误链完整
fmt.Println(errors.Unwrap(err2) == err1)  // true，可以解包
```

#### 🎯 练习题 4：实现错误链追踪

```go
// 任务：实现一个文件处理系统，要求：
// 1. 打开文件 -> 读取内容 -> 解析JSON -> 验证数据
// 2. 每一层都要添加上下文
// 3. 在main函数中能够判断具体是哪一步出错

type Config struct {
    Name string `json:"name"`
    Port int    `json:"port"`
}

func parseConfig(data []byte) (*Config, error) {
    // TODO: 实现
}

func validateConfig(cfg *Config) error {
    // TODO: 实现
}

func loadConfigFile(filename string) (*Config, error) {
    // TODO: 实现并正确包装错误
}

func main() {
    cfg, err := loadConfigFile("app.json")
    if err != nil {
        // TODO: 判断是文件不存在、JSON解析错误还是验证错误
    }
}
```

<details>
<summary>💡 参考答案</summary>

```go
import (
    "encoding/json"
    "errors"
    "fmt"
    "os"
)

type Config struct {
    Name string `json:"name"`
    Port int    `json:"port"`
}

// 自定义错误类型
var (
    ErrInvalidConfig = errors.New("invalid config")
    ErrEmptyName     = errors.New("name is empty")
    ErrInvalidPort   = errors.New("port is invalid")
)

func parseConfig(data []byte) (*Config, error) {
    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parse json: %w", err)
    }
    return &cfg, nil
}

func validateConfig(cfg *Config) error {
    if cfg.Name == "" {
        return fmt.Errorf("validate config: %w", ErrEmptyName)
    }
    if cfg.Port <= 0 || cfg.Port > 65535 {
        return fmt.Errorf("validate config: %w", ErrInvalidPort)
    }
    return nil
}

func loadConfigFile(filename string) (*Config, error) {
    // 第1步：读取文件
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("load config file %s: %w", filename, err)
    }

    // 第2步：解析JSON
    cfg, err := parseConfig(data)
    if err != nil {
        return nil, fmt.Errorf("load config file %s: %w", filename, err)
    }

    // 第3步：验证配置
    if err := validateConfig(cfg); err != nil {
        return nil, fmt.Errorf("load config file %s: %w", filename, err)
    }

    return cfg, nil
}

func main() {
    cfg, err := loadConfigFile("app.json")
    if err != nil {
        fmt.Println("错误:", err)

        // 判断具体错误类型
        switch {
        case errors.Is(err, os.ErrNotExist):
            fmt.Println("-> 配置文件不存在")

        case errors.Is(err, ErrEmptyName):
            fmt.Println("-> 配置名称为空")

        case errors.Is(err, ErrInvalidPort):
            fmt.Println("-> 端口号无效")

        default:
            // 检查是否是JSON错误
            var jsonErr *json.SyntaxError
            if errors.As(err, &jsonErr) {
                fmt.Printf("-> JSON语法错误，位置: %d\n", jsonErr.Offset)
            } else {
                fmt.Println("-> 未知错误")
            }
        }

        return
    }

    fmt.Printf("配置加载成功: %+v\n", cfg)
}
```
</details>

---

### 2.2 自定义错误类型的正确姿势

#### 📚 核心知识

**什么时候需要自定义错误类型？**
1. 需要携带额外的上下文信息（如错误码、字段名等）
2. 需要提供特定的行为方法（如 HTTP 状态码映射）
3. 需要区分不同类别的错误

#### 💡 标准模板

```go
// 模板1: 带错误码的错误
type AppError struct {
    Code    int    // 错误码
    Message string // 错误消息
    Err     error  // 原始错误（用于错误链）
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
    }
    return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// 模板2: 字段验证错误
type FieldError struct {
    Field string      // 字段名
    Value interface{} // 字段值
    Tag   string      // 验证标签
}

func (e *FieldError) Error() string {
    return fmt.Sprintf("validation failed on field '%s': %s (value: %v)",
        e.Field, e.Tag, e.Value)
}

// 模板3: 多错误聚合
type MultiError struct {
    Errors []error
}

func (m *MultiError) Error() string {
    msgs := make([]string, len(m.Errors))
    for i, err := range m.Errors {
        msgs[i] = err.Error()
    }
    return strings.Join(msgs, "; ")
}

func (m *MultiError) Unwrap() []error {
    return m.Errors
}
```

#### 💡 完整实战示例

```go
package main

import (
    "errors"
    "fmt"
    "net/http"
)

// 错误码定义
const (
    ErrCodeBadRequest     = 400
    ErrCodeUnauthorized   = 401
    ErrCodeNotFound       = 404
    ErrCodeInternalServer = 500
)

// HTTP错误类型
type HTTPError struct {
    Code    int
    Message string
    Err     error
}

func (e *HTTPError) Error() string {
    return e.Message
}

func (e *HTTPError) Unwrap() error {
    return e.Err
}

func (e *HTTPError) StatusCode() int {
    return e.Code
}

// 构造函数
func NewHTTPError(code int, message string, err error) *HTTPError {
    return &HTTPError{
        Code:    code,
        Message: message,
        Err:     err,
    }
}

// 业务逻辑
var (
    ErrUserNotFound = errors.New("user not found")
    ErrInvalidToken = errors.New("invalid token")
)

func getUserByID(id string) (*User, error) {
    if id == "" {
        return nil, NewHTTPError(ErrCodeBadRequest, "user ID is required", nil)
    }

    // 模拟数据库查询
    user := db.Find(id)
    if user == nil {
        return nil, NewHTTPError(ErrCodeNotFound, "user not found", ErrUserNotFound)
    }

    return user, nil
}

func authenticateUser(token string) (*User, error) {
    if token == "" {
        return nil, NewHTTPError(ErrCodeUnauthorized, "token is required", nil)
    }

    user, err := validateToken(token)
    if err != nil {
        return nil, NewHTTPError(ErrCodeUnauthorized, "invalid token", ErrInvalidToken)
    }

    return user, nil
}

// HTTP处理器
func handleGetUser(w http.ResponseWriter, r *http.Request) {
    // 认证
    token := r.Header.Get("Authorization")
    _, err := authenticateUser(token)
    if err != nil {
        handleError(w, err)
        return
    }

    // 获取用户
    userID := r.URL.Query().Get("id")
    user, err := getUserByID(userID)
    if err != nil {
        handleError(w, err)
        return
    }

    // 返回成功
    json.NewEncoder(w).Encode(user)
}

func handleError(w http.ResponseWriter, err error) {
    var httpErr *HTTPError
    if errors.As(err, &httpErr) {
        w.WriteHeader(httpErr.StatusCode())
        json.NewEncoder(w).Encode(map[string]string{
            "error": httpErr.Message,
        })
        return
    }

    // 未知错误，返回500
    w.WriteHeader(500)
    json.NewEncoder(w).Encode(map[string]string{
        "error": "Internal Server Error",
    })
}
```

#### 🎯 练习题 5：设计电商系统的错误类型

```go
// 任务：设计一个电商系统的错误处理
// 要求：
// 1. 定义订单相关的错误类型（库存不足、价格变动、订单已取消等）
// 2. 定义支付相关的错误类型（余额不足、支付超时等）
// 3. 实现 createOrder 函数，处理各种错误情况
// 4. 能够区分可重试错误和不可重试错误

type OrderError struct {
    // TODO: 定义字段
}

func (e *OrderError) Error() string {
    // TODO: 实现
}

// 定义可重试接口
type Retryable interface {
    error
    CanRetry() bool
}

func createOrder(userID int, productID int, quantity int) (*Order, error) {
    // TODO: 实现
}
```

<details>
<summary>💡 参考答案</summary>

```go
import (
    "errors"
    "fmt"
)

// 错误类型枚举
type ErrorType int

const (
    ErrTypeInsufficientStock ErrorType = iota // 库存不足
    ErrTypePriceChanged                       // 价格变动
    ErrTypeOrderCancelled                     // 订单已取消
    ErrTypeInsufficientBalance                // 余额不足
    ErrTypePaymentTimeout                     // 支付超时
    ErrTypeInvalidProduct                     // 无效商品
)

// 订单错误
type OrderError struct {
    Type      ErrorType
    OrderID   string
    ProductID int
    Message   string
    Err       error
}

func (e *OrderError) Error() string {
    return fmt.Sprintf("order error [%s]: %s", e.OrderID, e.Message)
}

func (e *OrderError) Unwrap() error {
    return e.Err
}

func (e *OrderError) CanRetry() bool {
    // 价格变动和支付超时可重试
    return e.Type == ErrTypePriceChanged || e.Type == ErrTypePaymentTimeout
}

// 支付错误
type PaymentError struct {
    Type    ErrorType
    UserID  int
    Amount  float64
    Message string
}

func (e *PaymentError) Error() string {
    return fmt.Sprintf("payment error [user:%d]: %s", e.UserID, e.Message)
}

func (e *PaymentError) CanRetry() bool {
    return e.Type == ErrTypePaymentTimeout
}

// 业务函数
func checkStock(productID, quantity int) error {
    stock := getStock(productID)
    if stock < quantity {
        return &OrderError{
            Type:      ErrTypeInsufficientStock,
            ProductID: productID,
            Message:   fmt.Sprintf("库存不足: 需要 %d, 可用 %d", quantity, stock),
        }
    }
    return nil
}

func checkPrice(productID int, expectedPrice float64) error {
    currentPrice := getPrice(productID)
    if currentPrice != expectedPrice {
        return &OrderError{
            Type:      ErrTypePriceChanged,
            ProductID: productID,
            Message:   fmt.Sprintf("价格已变动: 原价 %.2f, 现价 %.2f", expectedPrice, currentPrice),
        }
    }
    return nil
}

func processPayment(userID int, amount float64) error {
    balance := getBalance(userID)
    if balance < amount {
        return &PaymentError{
            Type:    ErrTypeInsufficientBalance,
            UserID:  userID,
            Amount:  amount,
            Message: fmt.Sprintf("余额不足: 需要 %.2f, 可用 %.2f", amount, balance),
        }
    }

    // 模拟支付
    if err := callPaymentGateway(userID, amount); err != nil {
        return &PaymentError{
            Type:    ErrTypePaymentTimeout,
            UserID:  userID,
            Amount:  amount,
            Message: "支付网关超时",
        }
    }

    return nil
}

func createOrder(userID, productID, quantity int, expectedPrice float64) (*Order, error) {
    // 1. 验证商品
    if !isValidProduct(productID) {
        return nil, &OrderError{
            Type:      ErrTypeInvalidProduct,
            ProductID: productID,
            Message:   "商品不存在",
        }
    }

    // 2. 检查库存
    if err := checkStock(productID, quantity); err != nil {
        return nil, err
    }

    // 3. 检查价格
    if err := checkPrice(productID, expectedPrice); err != nil {
        return nil, err
    }

    // 4. 处理支付
    totalAmount := expectedPrice * float64(quantity)
    if err := processPayment(userID, totalAmount); err != nil {
        return nil, err
    }

    // 5. 创建订单
    order := &Order{
        UserID:    userID,
        ProductID: productID,
        Quantity:  quantity,
        Amount:    totalAmount,
    }

    return order, nil
}

// 重试逻辑
func createOrderWithRetry(userID, productID, quantity int, expectedPrice float64, maxRetries int) (*Order, error) {
    var lastErr error

    for i := 0; i < maxRetries; i++ {
        order, err := createOrder(userID, productID, quantity, expectedPrice)
        if err == nil {
            return order, nil
        }

        lastErr = err

        // 检查是否可重试
        var retryable interface {
            CanRetry() bool
        }
        if errors.As(err, &retryable) && retryable.CanRetry() {
            fmt.Printf("尝试 %d 失败，重试中...\n", i+1)
            time.Sleep(time.Second * time.Duration(i+1))
            continue
        }

        // 不可重试的错误，直接返回
        return nil, err
    }

    return nil, fmt.Errorf("达到最大重试次数: %w", lastErr)
}
```
</details>

---

### 2.3 panic 和 recover 的正确使用

#### 📚 核心知识

**panic 的三个原则：**
1. panic 用于不可恢复的错误（程序员错误）
2. 库代码不应该 panic，应该返回 error
3. 每个 goroutine 都应该有 recover

#### 💡 基础用法

```go
// panic 示例
func divide(a, b int) int {
    if b == 0 {
        panic("division by zero") // panic 会终止程序
    }
    return a / b
}

// recover 示例
func safeDiv(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic recovered: %v", r)
        }
    }()

    result = divide(a, b)
    return result, nil
}

func main() {
    // 不会崩溃
    result, err := safeDiv(10, 0)
    if err != nil {
        fmt.Println("错误:", err)
    }
}
```

#### 🔥 关键场景

**场景1: goroutine 的 panic 保护**

```go
// ❌ 危险：子 goroutine panic 会导致整个程序崩溃
func worker(id int) {
    // 如果这里 panic，整个程序会崩溃
    data := processTask()
    fmt.Println(data)
}

func main() {
    for i := 0; i < 10; i++ {
        go worker(i)
    }
    time.Sleep(time.Second)
}

// ✅ 安全：每个 goroutine 都有 recover
func safeWorker(id int) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Worker %d panic: %v\n%s", id, r, debug.Stack())
        }
    }()

    data := processTask()
    fmt.Println(data)
}

// ✅ 更好：封装 safe goroutine 启动器
func safeGo(fn func()) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("Goroutine panic: %v\n%s", r, debug.Stack())
            }
        }()
        fn()
    }()
}

func main() {
    for i := 0; i < 10; i++ {
        i := i // 捕获变量
        safeGo(func() {
            processTask(i)
        })
    }
}
```

**场景2: Web 服务器的 panic 恢复**

```go
// HTTP 中间件：捕获 panic
func PanicRecovery(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("PANIC: %v\n%s", r, debug.Stack())

                http.Error(w, "Internal Server Error", 500)
            }
        }()

        next(w, r)
    }
}

// 使用
func handleRequest(w http.ResponseWriter, r *http.Request) {
    // 这里即使 panic，也不会导致服务器崩溃
    data := riskyOperation()
    json.NewEncoder(w).Encode(data)
}

func main() {
    http.HandleFunc("/api", PanicRecovery(handleRequest))
    http.ListenAndServe(":8080", nil)
}
```

#### 🎯 练习题 6：实现安全的任务调度器

```go
// 任务：实现一个任务调度器
// 要求：
// 1. 支持并发执行多个任务
// 2. 某个任务 panic 不影响其他任务
// 3. 收集所有任务的结果和错误
// 4. 支持超时控制

type Task func() (interface{}, error)

type Result struct {
    Value interface{}
    Error error
}

type Scheduler struct {
    // TODO: 定义字段
}

func (s *Scheduler) Submit(task Task) {
    // TODO: 实现
}

func (s *Scheduler) Wait() []Result {
    // TODO: 实现
}
```

<details>
<summary>💡 参考答案</summary>

```go
import (
    "context"
    "fmt"
    "runtime/debug"
    "sync"
    "time"
)

type Task func() (interface{}, error)

type Result struct {
    Value interface{}
    Error error
    Panic interface{} // 记录 panic 信息
}

type Scheduler struct {
    tasks   []Task
    results []Result
    mu      sync.Mutex
    wg      sync.WaitGroup
    timeout time.Duration
}

func NewScheduler(timeout time.Duration) *Scheduler {
    return &Scheduler{
        timeout: timeout,
    }
}

func (s *Scheduler) Submit(task Task) {
    s.mu.Lock()
    s.tasks = append(s.tasks, task)
    s.mu.Unlock()
}

func (s *Scheduler) Run() []Result {
    s.results = make([]Result, len(s.tasks))
    ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
    defer cancel()

    for i, task := range s.tasks {
        s.wg.Add(1)
        go s.runTask(ctx, i, task)
    }

    s.wg.Wait()
    return s.results
}

func (s *Scheduler) runTask(ctx context.Context, index int, task Task) {
    defer s.wg.Done()

    // panic 恢复
    defer func() {
        if r := recover(); r != nil {
            s.mu.Lock()
            s.results[index] = Result{
                Panic: r,
                Error: fmt.Errorf("task panic: %v\n%s", r, debug.Stack()),
            }
            s.mu.Unlock()
        }
    }()

    // 带超时的任务执行
    done := make(chan struct{})
    var value interface{}
    var err error

    go func() {
        value, err = task()
        close(done)
    }()

    select {
    case <-done:
        s.mu.Lock()
        s.results[index] = Result{
            Value: value,
            Error: err,
        }
        s.mu.Unlock()

    case <-ctx.Done():
        s.mu.Lock()
        s.results[index] = Result{
            Error: fmt.Errorf("task timeout"),
        }
        s.mu.Unlock()
    }
}

// 使用示例
func main() {
    scheduler := NewScheduler(5 * time.Second)

    // 正常任务
    scheduler.Submit(func() (interface{}, error) {
        time.Sleep(1 * time.Second)
        return "task 1 completed", nil
    })

    // 返回错误的任务
    scheduler.Submit(func() (interface{}, error) {
        return nil, errors.New("task 2 failed")
    })

    // 会 panic 的任务
    scheduler.Submit(func() (interface{}, error) {
        panic("task 3 panic!")
    })

    // 超时的任务
    scheduler.Submit(func() (interface{}, error) {
        time.Sleep(10 * time.Second)
        return "task 4 completed", nil
    })

    results := scheduler.Run()

    for i, result := range results {
        fmt.Printf("Task %d:\n", i+1)
        if result.Panic != nil {
            fmt.Printf("  PANIC: %v\n", result.Panic)
        } else if result.Error != nil {
            fmt.Printf("  ERROR: %v\n", result.Error)
        } else {
            fmt.Printf("  SUCCESS: %v\n", result.Value)
        }
    }
}
```
</details>

---

## 第三阶段：实战提升

### 3.1 错误处理的设计模式

#### 模式1: 哨兵错误模式

```go
// 定义
var (
    ErrNotFound      = errors.New("not found")
    ErrAlreadyExists = errors.New("already exists")
    ErrUnauthorized  = errors.New("unauthorized")
)

// 使用
func FindUser(id string) (*User, error) {
    user := db.Find(id)
    if user == nil {
        return nil, ErrNotFound
    }
    return user, nil
}

// 检查
user, err := FindUser("123")
if err == ErrNotFound {
    // 处理未找到的情况
}
```

**优点：** 简单、高效、易于比较
**缺点：** 无法携带上下文信息

#### 模式2: 错误类型模式

```go
// 定义
type NotFoundError struct {
    Resource string
    ID       string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

// 使用
func FindUser(id string) (*User, error) {
    user := db.Find(id)
    if user == nil {
        return nil, &NotFoundError{Resource: "user", ID: id}
    }
    return user, nil
}

// 检查
user, err := FindUser("123")
var notFoundErr *NotFoundError
if errors.As(err, &notFoundErr) {
    fmt.Printf("未找到资源: %s, ID: %s\n", notFoundErr.Resource, notFoundErr.ID)
}
```

**优点：** 可以携带丰富信息
**缺点：** 类型暴露，耦合度高

#### 模式3: 行为模式（接口）

```go
// 定义行为接口
type Temporary interface {
    Temporary() bool
}

type Timeout interface {
    Timeout() bool
}

// 实现
type NetworkError struct {
    Op  string
    Err error
}

func (e *NetworkError) Error() string {
    return fmt.Sprintf("network %s: %v", e.Op, e.Err)
}

func (e *NetworkError) Temporary() bool {
    return true // 网络错误通常是临时的
}

// 使用
func fetch(url string) error {
    return &NetworkError{Op: "GET", Err: errors.New("timeout")}
}

// 重试逻辑
for i := 0; i < maxRetries; i++ {
    err := fetch(url)
    if err == nil {
        break
    }

    if temp, ok := err.(Temporary); ok && temp.Temporary() {
        time.Sleep(backoff)
        continue
    }

    return err // 非临时错误，直接返回
}
```

**优点：** 解耦，基于行为而非类型
**缺点：** 需要定义接口

#### 🎯 练习题 7：选择合适的模式

```go
// 场景1: 实现一个缓存系统
// 要求：需要区分缓存未命中、缓存过期、缓存错误

// 场景2: 实现一个重试机制
// 要求：根据错误类型决定是否重试

// 请为以上两个场景选择合适的错误处理模式并实现
```

<details>
<summary>💡 参考答案</summary>

```go
// 场景1: 缓存系统 - 使用哨兵错误模式
var (
    ErrCacheMiss    = errors.New("cache miss")
    ErrCacheExpired = errors.New("cache expired")
)

type Cache struct {
    data map[string]cacheItem
}

type cacheItem struct {
    value      interface{}
    expiration time.Time
}

func (c *Cache) Get(key string) (interface{}, error) {
    item, ok := c.data[key]
    if !ok {
        return nil, ErrCacheMiss
    }

    if time.Now().After(item.expiration) {
        return nil, ErrCacheExpired
    }

    return item.value, nil
}

// 使用
value, err := cache.Get("user:123")
switch {
case err == ErrCacheMiss:
    // 从数据库加载
    value = loadFromDB("user:123")
    cache.Set("user:123", value)
case err == ErrCacheExpired:
    // 后台异步刷新缓存
    go refreshCache("user:123")
    // 返回过期数据
    value = cache.GetExpired("user:123")
}

// 场景2: 重试机制 - 使用行为模式
type Retryable interface {
    error
    ShouldRetry() bool
}

type HTTPError struct {
    StatusCode int
    Message    string
}

func (e *HTTPError) Error() string {
    return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

func (e *HTTPError) ShouldRetry() bool {
    // 5xx 错误和 429 Too Many Requests 可以重试
    return e.StatusCode >= 500 || e.StatusCode == 429
}

type NetworkError struct {
    Op  string
    Err error
}

func (e *NetworkError) Error() string {
    return fmt.Sprintf("network error: %s", e.Op)
}

func (e *NetworkError) ShouldRetry() bool {
    return true // 网络错误总是可以重试
}

// 通用重试函数
func retryWithBackoff(fn func() error, maxRetries int) error {
    var lastErr error

    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }

        lastErr = err

        // 检查是否可重试
        if retryable, ok := err.(Retryable); ok && retryable.ShouldRetry() {
            backoff := time.Duration(1<<uint(i)) * time.Second
            time.Sleep(backoff)
            continue
        }

        // 不可重试，直接返回
        return err
    }

    return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// 使用
err := retryWithBackoff(func() error {
    resp, err := http.Get("https://api.example.com/data")
    if err != nil {
        return &NetworkError{Op: "GET", Err: err}
    }

    if resp.StatusCode != 200 {
        return &HTTPError{
            StatusCode: resp.StatusCode,
            Message:    resp.Status,
        }
    }

    return nil
}, 3)
```
</details>

---

### 3.2 并发场景的错误处理

#### 💡 使用 errgroup

```go
import "golang.org/x/sync/errgroup"

func processFiles(files []string) error {
    var g errgroup.Group

    for _, file := range files {
        file := file // 捕获变量
        g.Go(func() error {
            return processFile(file)
        })
    }

    // 等待所有任务，返回第一个错误
    return g.Wait()
}

// 带上下文的版本
func processFilesWithContext(ctx context.Context, files []string) error {
    g, ctx := errgroup.WithContext(ctx)

    for _, file := range files {
        file := file
        g.Go(func() error {
            // 如果某个任务返回错误，ctx 会被取消
            select {
            case <-ctx.Done():
                return ctx.Err()
            default:
                return processFile(file)
            }
        })
    }

    return g.Wait()
}
```

#### 💡 收集所有错误

```go
type MultiError []error

func (m MultiError) Error() string {
    var msgs []string
    for _, err := range m {
        msgs = append(msgs, err.Error())
    }
    return strings.Join(msgs, "; ")
}

func processAllFiles(files []string) error {
    var (
        wg     sync.WaitGroup
        mu     sync.Mutex
        errors MultiError
    )

    for _, file := range files {
        wg.Add(1)
        go func(f string) {
            defer wg.Done()

            if err := processFile(f); err != nil {
                mu.Lock()
                errors = append(errors, fmt.Errorf("process %s: %w", f, err))
                mu.Unlock()
            }
        }(file)
    }

    wg.Wait()

    if len(errors) > 0 {
        return errors
    }
    return nil
}
```

#### 🎯 练习题 8：实现并发下载器

```go
// 任务：实现一个并发下载器
// 要求：
// 1. 并发下载多个文件
// 2. 限制并发数（如最多5个）
// 3. 某个文件下载失败不影响其他文件
// 4. 收集所有下载结果（成功或失败）
// 5. 支持超时和取消

type DownloadResult struct {
    URL      string
    Filename string
    Size     int64
    Error    error
}

type Downloader struct {
    // TODO: 定义字段
}

func (d *Downloader) Download(urls []string) []DownloadResult {
    // TODO: 实现
}
```

<details>
<summary>💡 参考答案</summary>

```go
import (
    "context"
    "fmt"
    "io"
    "net/http"
    "os"
    "path"
    "sync"
    "time"
)

type DownloadResult struct {
    URL      string
    Filename string
    Size     int64
    Error    error
    Duration time.Duration
}

type Downloader struct {
    MaxConcurrency int
    Timeout        time.Duration
    OutputDir      string
}

func NewDownloader(maxConcurrency int, timeout time.Duration, outputDir string) *Downloader {
    return &Downloader{
        MaxConcurrency: maxConcurrency,
        Timeout:        timeout,
        OutputDir:      outputDir,
    }
}

func (d *Downloader) Download(ctx context.Context, urls []string) []DownloadResult {
    results := make([]DownloadResult, len(urls))

    // 使用带缓冲的 channel 限制并发
    semaphore := make(chan struct{}, d.MaxConcurrency)
    var wg sync.WaitGroup

    for i, url := range urls {
        wg.Add(1)

        go func(index int, url string) {
            defer wg.Done()

            // 获取信号量
            semaphore <- struct{}{}
            defer func() { <-semaphore }()

            results[index] = d.downloadOne(ctx, url)
        }(i, url)
    }

    wg.Wait()
    return results
}

func (d *Downloader) downloadOne(ctx context.Context, url string) DownloadResult {
    start := time.Now()
    result := DownloadResult{
        URL: url,
    }

    // 创建带超时的上下文
    ctx, cancel := context.WithTimeout(ctx, d.Timeout)
    defer cancel()

    // 创建 HTTP 请求
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        result.Error = fmt.Errorf("create request: %w", err)
        return result
    }

    // 发送请求
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        result.Error = fmt.Errorf("http request: %w", err)
        return result
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        result.Error = fmt.Errorf("http status: %d", resp.StatusCode)
        return result
    }

    // 确定文件名
    filename := path.Base(url)
    if filename == "." || filename == "/" {
        filename = "downloaded_file"
    }
    filepath := path.Join(d.OutputDir, filename)
    result.Filename = filepath

    // 创建文件
    file, err := os.Create(filepath)
    if err != nil {
        result.Error = fmt.Errorf("create file: %w", err)
        return result
    }
    defer file.Close()

    // 下载并写入文件
    size, err := io.Copy(file, resp.Body)
    if err != nil {
        result.Error = fmt.Errorf("write file: %w", err)
        os.Remove(filepath) // 清理不完整的文件
        return result
    }

    result.Size = size
    result.Duration = time.Since(start)
    return result
}

// 使用示例
func main() {
    urls := []string{
        "https://example.com/file1.zip",
        "https://example.com/file2.zip",
        "https://example.com/file3.zip",
        "https://invalid-url", // 这个会失败
    }

    downloader := NewDownloader(3, 30*time.Second, "./downloads")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    results := downloader.Download(ctx, urls)

    // 统计结果
    var succeeded, failed int
    var totalSize int64

    for _, result := range results {
        if result.Error != nil {
            failed++
            fmt.Printf("❌ %s: %v\n", result.URL, result.Error)
        } else {
            succeeded++
            totalSize += result.Size
            fmt.Printf("✅ %s: %s (%.2f MB in %v)\n",
                result.URL,
                result.Filename,
                float64(result.Size)/1024/1024,
                result.Duration)
        }
    }

    fmt.Printf("\n总计: 成功 %d, 失败 %d, 总大小 %.2f MB\n",
        succeeded, failed, float64(totalSize)/1024/1024)
}
```
</details>

---

## 第四阶段：手撕代码

### 🔥 题目1: 实现带重试的 HTTP 客户端

**难度：⭐⭐**

```go
// 要求：
// 1. 实现 Do 方法，支持自动重试
// 2. 只有特定的错误才重试（网络错误、5xx、429）
// 3. 使用指数退避策略
// 4. 支持设置最大重试次数

type RetryableClient struct {
    client      *http.Client
    maxRetries  int
    initialWait time.Duration
}

func (c *RetryableClient) Do(req *http.Request) (*http.Response, error) {
    // TODO: 实现
}
```

<details>
<summary>💡 参考答案</summary>

```go
import (
    "errors"
    "fmt"
    "io"
    "math"
    "net/http"
    "time"
)

type RetryableClient struct {
    client      *http.Client
    maxRetries  int
    initialWait time.Duration
}

func NewRetryableClient(maxRetries int, initialWait time.Duration) *RetryableClient {
    return &RetryableClient{
        client:      &http.Client{Timeout: 30 * time.Second},
        maxRetries:  maxRetries,
        initialWait: initialWait,
    }
}

func (c *RetryableClient) Do(req *http.Request) (*http.Response, error) {
    var lastErr error

    for attempt := 0; attempt <= c.maxRetries; attempt++ {
        // 克隆请求（因为 Body 只能读一次）
        clonedReq := c.cloneRequest(req)

        resp, err := c.client.Do(clonedReq)

        // 成功
        if err == nil && !c.shouldRetry(resp.StatusCode) {
            return resp, nil
        }

        // 记录错误
        if err != nil {
            lastErr = err
        } else {
            lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
            resp.Body.Close()
        }

        // 已经是最后一次尝试
        if attempt == c.maxRetries {
            break
        }

        // 计算退避时间（指数退避）
        wait := c.initialWait * time.Duration(math.Pow(2, float64(attempt)))
        time.Sleep(wait)
    }

    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *RetryableClient) shouldRetry(statusCode int) bool {
    // 5xx 和 429 可以重试
    return statusCode >= 500 || statusCode == 429
}

func (c *RetryableClient) cloneRequest(req *http.Request) *http.Request {
    cloned := req.Clone(req.Context())

    // 如果有 Body，需要重新读取
    if req.Body != nil {
        // 注意：这要求原始 Body 可以被多次读取
        // 实际应用中可能需要先读取到内存中
        body, _ := io.ReadAll(req.Body)
        req.Body = io.NopCloser(bytes.NewReader(body))
        cloned.Body = io.NopCloser(bytes.NewReader(body))
    }

    return cloned
}

// 测试
func main() {
    client := NewRetryableClient(3, 1*time.Second)

    req, _ := http.NewRequest("GET", "https://httpbin.org/status/500", nil)

    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer resp.Body.Close()

    fmt.Println("Success:", resp.Status)
}
```
</details>

---

### 🔥 题目2: 实现错误聚合器

**难度：⭐⭐**

```go
// 要求：
// 1. 可以添加多个错误
// 2. 可以按类型分组错误
// 3. 可以获取第一个错误
// 4. 可以判断是否包含特定错误

type ErrorAggregator struct {
    // TODO: 实现
}

func (ea *ErrorAggregator) Add(err error) {
    // TODO: 实现
}

func (ea *ErrorAggregator) HasErrors() bool {
    // TODO: 实现
}

func (ea *ErrorAggregator) First() error {
    // TODO: 实现
}

func (ea *ErrorAggregator) Contains(target error) bool {
    // TODO: 实现
}

func (ea *ErrorAggregator) Error() string {
    // TODO: 实现
}
```

<details>
<summary>💡 参考答案</summary>

```go
import (
    "errors"
    "fmt"
    "strings"
)

type ErrorAggregator struct {
    errors []error
    groups map[string][]error
}

func NewErrorAggregator() *ErrorAggregator {
    return &ErrorAggregator{
        errors: make([]error, 0),
        groups: make(map[string][]error),
    }
}

func (ea *ErrorAggregator) Add(err error) {
    if err == nil {
        return
    }

    ea.errors = append(ea.errors, err)
}

func (ea *ErrorAggregator) AddWithGroup(group string, err error) {
    if err == nil {
        return
    }

    ea.errors = append(ea.errors, err)
    ea.groups[group] = append(ea.groups[group], err)
}

func (ea *ErrorAggregator) HasErrors() bool {
    return len(ea.errors) > 0
}

func (ea *ErrorAggregator) First() error {
    if len(ea.errors) == 0 {
        return nil
    }
    return ea.errors[0]
}

func (ea *ErrorAggregator) Contains(target error) bool {
    for _, err := range ea.errors {
        if errors.Is(err, target) {
            return true
        }
    }
    return false
}

func (ea *ErrorAggregator) GetGroup(group string) []error {
    return ea.groups[group]
}

func (ea *ErrorAggregator) Error() string {
    if len(ea.errors) == 0 {
        return ""
    }

    if len(ea.errors) == 1 {
        return ea.errors[0].Error()
    }

    var sb strings.Builder
    sb.WriteString(fmt.Sprintf("%d errors occurred:\n", len(ea.errors)))

    for i, err := range ea.errors {
        sb.WriteString(fmt.Sprintf("  [%d] %v\n", i+1, err))
    }

    return sb.String()
}

func (ea *ErrorAggregator) Unwrap() []error {
    return ea.errors
}

// 使用示例
func validateUser(user *User) error {
    agg := NewErrorAggregator()

    if user.Name == "" {
        agg.AddWithGroup("validation", errors.New("name is required"))
    }

    if user.Email == "" {
        agg.AddWithGroup("validation", errors.New("email is required"))
    }

    if user.Age < 0 {
        agg.AddWithGroup("validation", errors.New("age must be positive"))
    }

    if !agg.HasErrors() {
        return nil
    }

    return agg
}
```
</details>

---

### 🔥 题目3: 实现 Circuit Breaker（熔断器）

**难度：⭐⭐⭐**

```go
// 要求：
// 1. 三种状态：Closed(正常)、Open(熔断)、HalfOpen(半开)
// 2. 失败次数超过阈值 -> Open
// 3. Open 状态等待一段时间 -> HalfOpen
// 4. HalfOpen 状态如果成功 -> Closed，如果失败 -> Open

type CircuitBreaker struct {
    // TODO: 实现
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    // TODO: 实现
}
```

<details>
<summary>💡 参考答案</summary>

```go
import (
    "errors"
    "fmt"
    "sync"
    "time"
)

type State int

const (
    StateClosed State = iota
    StateOpen
    StateHalfOpen
)

var (
    ErrCircuitOpen = errors.New("circuit breaker is open")
)

type CircuitBreaker struct {
    maxFailures  int
    timeout      time.Duration

    mu           sync.RWMutex
    state        State
    failures     int
    lastFailTime time.Time
}

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        maxFailures: maxFailures,
        timeout:     timeout,
        state:       StateClosed,
    }
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    // 检查状态
    if err := cb.beforeCall(); err != nil {
        return err
    }

    // 执行函数
    err := fn()

    // 更新状态
    cb.afterCall(err)

    return err
}

func (cb *CircuitBreaker) beforeCall() error {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    switch cb.state {
    case StateClosed:
        return nil

    case StateOpen:
        // 检查是否可以切换到 HalfOpen
        if time.Since(cb.lastFailTime) > cb.timeout {
            cb.state = StateHalfOpen
            return nil
        }
        return ErrCircuitOpen

    case StateHalfOpen:
        return nil
    }

    return nil
}

func (cb *CircuitBreaker) afterCall(err error) {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    if err != nil {
        cb.onFailure()
    } else {
        cb.onSuccess()
    }
}

func (cb *CircuitBreaker) onSuccess() {
    cb.failures = 0
    cb.state = StateClosed
}

func (cb *CircuitBreaker) onFailure() {
    cb.failures++
    cb.lastFailTime = time.Now()

    if cb.failures >= cb.maxFailures {
        cb.state = StateOpen
    }
}

func (cb *CircuitBreaker) State() State {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    return cb.state
}

// 测试
func main() {
    cb := NewCircuitBreaker(3, 5*time.Second)

    callAPI := func() error {
        // 模拟 API 调用
        if rand.Float64() < 0.7 {
            return errors.New("API error")
        }
        return nil
    }

    for i := 0; i < 10; i++ {
        err := cb.Call(callAPI)

        fmt.Printf("Call %d: state=%v, err=%v\n", i+1, cb.State(), err)

        time.Sleep(1 * time.Second)
    }
}
```
</details>

---

### 🔥 题目4: 实现错误恢复中间件

**难度：⭐⭐**

```go
// 要求：
// 1. 捕获 handler 中的 panic
// 2. 记录错误日志和堆栈
// 3. 返回友好的错误响应
// 4. 区分不同类型的 panic（string、error、其他）

func RecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
    // TODO: 实现
}
```

<details>
<summary>💡 参考答案</summary>

```go
import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "runtime/debug"
)

type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
    TraceID string `json:"trace_id,omitempty"`
}

func RecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if rec := recover(); rec != nil {
                // 获取堆栈信息
                stack := debug.Stack()

                // 获取 trace ID（如果有的话）
                traceID := r.Header.Get("X-Trace-ID")

                // 记录日志
                log.Printf("PANIC recovered: %v\nTrace-ID: %s\nStack:\n%s",
                    rec, traceID, stack)

                // 构造错误响应
                var errMsg string
                switch v := rec.(type) {
                case string:
                    errMsg = v
                case error:
                    errMsg = v.Error()
                default:
                    errMsg = fmt.Sprintf("%v", v)
                }

                response := ErrorResponse{
                    Error:   "Internal Server Error",
                    Message: errMsg,
                    TraceID: traceID,
                }

                // 返回 JSON 响应
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusInternalServerError)
                json.NewEncoder(w).Encode(response)
            }
        }()

        next(w, r)
    }
}

// 使用示例
func panicHandler(w http.ResponseWriter, r *http.Request) {
    // 模拟不同类型的 panic
    switch r.URL.Query().Get("type") {
    case "string":
        panic("something went wrong")
    case "error":
        panic(errors.New("error object"))
    case "nil":
        var p *int
        _ = *p // nil pointer dereference
    default:
        panic(12345)
    }
}

func main() {
    http.HandleFunc("/api/test", RecoveryMiddleware(panicHandler))
    http.ListenAndServe(":8080", nil)
}
```
</details>

---

### 🔥 题目5: 实现超时控制

**难度：⭐⭐⭐**

```go
// 要求：
// 1. 对任意函数添加超时控制
// 2. 超时后能够取消执行中的函数
// 3. 返回超时错误
// 4. 支持泛型（或使用 interface{}）

func WithTimeout(fn func() (interface{}, error), timeout time.Duration) (interface{}, error) {
    // TODO: 实现
}
```

<details>
<summary>💡 参考答案</summary>

```go
import (
    "context"
    "errors"
    "fmt"
    "time"
)

var ErrTimeout = errors.New("operation timeout")

// 方案1: 使用 context
func WithTimeout(fn func(context.Context) (interface{}, error), timeout time.Duration) (interface{}, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    type result struct {
        value interface{}
        err   error
    }

    resultChan := make(chan result, 1)

    go func() {
        value, err := fn(ctx)
        resultChan <- result{value, err}
    }()

    select {
    case res := <-resultChan:
        return res.value, res.err
    case <-ctx.Done():
        return nil, fmt.Errorf("%w: %v", ErrTimeout, ctx.Err())
    }
}

// 方案2: 泛型版本（Go 1.18+）
func WithTimeoutGeneric[T any](fn func(context.Context) (T, error), timeout time.Duration) (T, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    type result struct {
        value T
        err   error
    }

    resultChan := make(chan result, 1)

    go func() {
        value, err := fn(ctx)
        resultChan <- result{value, err}
    }()

    select {
    case res := <-resultChan:
        return res.value, res.err
    case <-ctx.Done():
        var zero T
        return zero, fmt.Errorf("%w: %v", ErrTimeout, ctx.Err())
    }
}

// 使用示例
func slowOperation(ctx context.Context) (string, error) {
    select {
    case <-time.After(3 * time.Second):
        return "completed", nil
    case <-ctx.Done():
        return "", ctx.Err()
    }
}

func main() {
    // 会超时
    result, err := WithTimeoutGeneric(slowOperation, 1*time.Second)
    if err != nil {
        if errors.Is(err, ErrTimeout) {
            fmt.Println("操作超时")
        }
        fmt.Println("Error:", err)
        return
    }

    fmt.Println("Result:", result)
}
```
</details>

---

## 第五阶段：面试高频考点

### 考点1: error 接口的本质

**问题：** Go 的 error 是什么？为什么设计成接口？

**答案：**
```go
// error 是内置接口
type error interface {
    Error() string
}

// 为什么是接口？
// 1. 灵活性：任何类型都可以实现 error
// 2. 可扩展：可以携带额外信息
// 3. 统一：标准化错误处理方式

// 示例
type MyError struct {
    Code int
    Msg  string
}

func (e *MyError) Error() string {
    return fmt.Sprintf("[%d] %s", e.Code, e.Msg)
}
```

**追问：** nil error 的陷阱是什么？

```go
// 陷阱示例
func returnsError() error {
    var p *MyError = nil
    if p == nil {
        return p // 陷阱！
    }
    return nil
}

func main() {
    err := returnsError()
    fmt.Println(err == nil) // false！！！
}

// 原因：interface 包含类型和值两部分
// (*MyError, nil) != (nil, nil)

// 正确做法
func returnsError() error {
    var p *MyError = nil
    if p == nil {
        return nil // 返回 nil interface
    }
    return p
}
```

---

### 考点2: %v 和 %w 的区别

**问题：** fmt.Errorf 中 %v 和 %w 有什么区别？

**答案：**
```go
err1 := errors.New("original")

// %v: 格式化字符串，丢失错误链
err2 := fmt.Errorf("wrapped: %v", err1)
errors.Is(err2, err1)        // false
errors.Unwrap(err2)          // nil

// %w: 包装错误，保留错误链
err3 := fmt.Errorf("wrapped: %w", err1)
errors.Is(err3, err1)        // true
errors.Unwrap(err3) == err1  // true

// 使用场景
// 1. 只需要错误信息 -> %v
// 2. 需要保留错误链，用于 errors.Is/As -> %w
```

---

### 考点3: errors.Is 和 errors.As 的区别

**问题：** 什么时候用 errors.Is，什么时候用 errors.As？

**答案：**
```go
// errors.Is: 判断错误链中是否包含特定错误
// 用于：哨兵错误
if errors.Is(err, os.ErrNotExist) {
    // 文件不存在
}

// errors.As: 获取错误链中特定类型的错误
// 用于：自定义错误类型
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    fmt.Println("Path:", pathErr.Path)
    fmt.Println("Op:", pathErr.Op)
}

// 实现原理
func Is(err, target error) bool {
    // 遍历错误链，比较每个错误
    for {
        if err == target {
            return true
        }
        if err = Unwrap(err); err == nil {
            return false
        }
    }
}

func As(err error, target interface{}) bool {
    // 遍历错误链，尝试类型断言
    for {
        if reflect.TypeOf(err) == reflect.TypeOf(target) {
            reflect.ValueOf(target).Elem().Set(reflect.ValueOf(err))
            return true
        }
        if err = Unwrap(err); err == nil {
            return false
        }
    }
}
```

---

### 考点4: panic 和 recover 的机制

**问题：** panic 和 recover 的工作原理是什么？有哪些注意事项？

**答案：**
```go
// 1. panic 触发时会立即停止当前函数
// 2. 执行所有 defer（按LIFO顺序）
// 3. 向上传播，直到程序崩溃或被 recover

func example() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("recovered:", r)
        }
    }()

    fmt.Println("before panic")
    panic("oops")
    fmt.Println("after panic") // 不会执行
}

// 注意事项：
// 1. recover 只能在 defer 中调用
// 2. recover 只能捕获当前 goroutine 的 panic
// 3. 多次 panic，只有最后一个会被 recover

// 陷阱示例
func wrong() {
    recover() // 无效！不在 defer 中
    panic("oops")
}

func wrongGoroutine() {
    defer func() {
        recover() // 无效！不同的 goroutine
    }()

    go func() {
        panic("oops")
    }()
}

// 正确做法
func correctGoroutine() {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("goroutine panic: %v", r)
            }
        }()

        // 可能 panic 的代码
    }()
}
```

---

### 考点5: 何时使用 panic

**问题：** 什么情况下应该使用 panic？

**答案：**
```go
// ✅ 应该使用 panic 的场景：

// 1. 不可能发生的情况（程序员错误）
func process(data []int) {
    if len(data) == 0 {
        panic("BUG: empty data should never happen")
    }
}

// 2. 初始化失败
func init() {
    cfg, err := loadConfig()
    if err != nil {
        panic(fmt.Sprintf("failed to load config: %v", err))
    }
}

// 3. 不变量被破坏
type Account struct {
    balance int
}

func (a *Account) withdraw(amount int) {
    if amount > a.balance {
        panic("invariant violated: insufficient balance")
    }
    a.balance -= amount
}

// ❌ 不应该使用 panic 的场景：

// 1. 普通的错误处理
func readFile(name string) ([]byte, error) {
    // ❌ 不要这样
    data, err := os.ReadFile(name)
    if err != nil {
        panic(err)
    }
    return data, nil

    // ✅ 应该这样
    return os.ReadFile(name)
}

// 2. 可预期的错误
func parseInt(s string) (int, error) {
    // ❌ 不要这样
    n, err := strconv.Atoi(s)
    if err != nil {
        panic(err)
    }
    return n, nil

    // ✅ 应该这样
    return strconv.Atoi(s)
}
```

---

### 考点6: 错误处理的最佳实践

**问题：** Go 错误处理有哪些最佳实践？

**答案：**
```go
// 1. 总是检查错误
file, err := os.Open("file.txt")
if err != nil {
    return err
}
defer file.Close()

// 2. 错误只处理一次
// ❌ 不要既记录又返回
func bad() error {
    err := doSomething()
    if err != nil {
        log.Printf("error: %v", err) // 记录
        return err                    // 返回
    }
    return nil
}

// ✅ 只返回，让上层决定
func good() error {
    if err := doSomething(); err != nil {
        return fmt.Errorf("do something: %w", err)
    }
    return nil
}

// 3. 为调用者添加上下文
// ❌ 丢失上下文
return err

// ✅ 添加上下文
return fmt.Errorf("process user %d: %w", userID, err)

// 4. 使用哨兵错误
var ErrNotFound = errors.New("not found")

if err == ErrNotFound {
    // 处理
}

// 5. 自定义错误类型携带额外信息
type QueryError struct {
    Query string
    Err   error
}

// 6. 在包边界转换 panic 为 error
func SafeCall(fn func()) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic: %v", r)
        }
    }()
    fn()
    return nil
}
```

---

### 考点7: defer 在错误处理中的应用

**问题：** defer 在错误处理中有什么作用？有哪些陷阱？

**答案：**
```go
// 1. 资源清理
func processFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close() // 确保文件被关闭

    // 处理文件...
    return nil
}

// 2. 修改返回值
func example() (err error) {
    defer func() {
        if err != nil {
            err = fmt.Errorf("example failed: %w", err)
        }
    }()

    return doSomething()
}

// 3. panic 恢复
func safe() (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic: %v", r)
        }
    }()

    // 可能 panic 的代码
    return nil
}

// 陷阱1: defer 中的闭包
func trap1() error {
    var err error
    defer func() {
        // 这里捕获的是最终的 err 值
        if err != nil {
            log.Println(err)
        }
    }()

    err = errors.New("error 1")
    err = nil // err 被覆盖
    return nil // defer 中不会打印
}

// 陷阱2: defer 调用顺序（LIFO）
func trap2() {
    defer fmt.Println("1")
    defer fmt.Println("2")
    defer fmt.Println("3")
    // 输出: 3, 2, 1
}

// 陷阱3: defer 中的参数立即求值
func trap3() {
    i := 0
    defer fmt.Println(i) // 输出 0，不是 1
    i++
}
```

---

### 考点8: 并发错误处理

**问题：** 如何在并发场景下处理错误？

**答案：**
```go
// 方案1: 使用 errgroup（返回第一个错误）
import "golang.org/x/sync/errgroup"

func concurrent1(urls []string) error {
    var g errgroup.Group

    for _, url := range urls {
        url := url
        g.Go(func() error {
            return fetch(url)
        })
    }

    return g.Wait() // 返回第一个错误
}

// 方案2: 收集所有错误
func concurrent2(urls []string) error {
    var (
        wg     sync.WaitGroup
        mu     sync.Mutex
        errors []error
    )

    for _, url := range urls {
        wg.Add(1)
        go func(u string) {
            defer wg.Done()

            if err := fetch(u); err != nil {
                mu.Lock()
                errors = append(errors, err)
                mu.Unlock()
            }
        }(url)
    }

    wg.Wait()

    if len(errors) > 0 {
        return fmt.Errorf("multiple errors: %v", errors)
    }
    return nil
}

// 方案3: 使用 channel 收集错误
func concurrent3(urls []string) error {
    errChan := make(chan error, len(urls))

    for _, url := range urls {
        go func(u string) {
            errChan <- fetch(u)
        }(url)
    }

    var errors []error
    for i := 0; i < len(urls); i++ {
        if err := <-errChan; err != nil {
            errors = append(errors, err)
        }
    }

    if len(errors) > 0 {
        return fmt.Errorf("multiple errors: %v", errors)
    }
    return nil
}
```

---

### 考点9: 性能考虑

**问题：** 错误处理对性能有什么影响？如何优化？

**答案：**
```go
// 1. 哨兵错误 vs 创建新错误
var ErrNotFound = errors.New("not found") // 预分配

func find1(id string) error {
    return ErrNotFound // 无分配，快
}

func find2(id string) error {
    return errors.New("not found") // 每次分配，慢
}

// Benchmark 结果：
// find1: 0.3 ns/op, 0 allocs
// find2: 10 ns/op, 1 allocs

// 2. %v vs %w
func wrap1(err error) error {
    return fmt.Errorf("wrap: %v", err) // 不保留错误链
}

func wrap2(err error) error {
    return fmt.Errorf("wrap: %w", err) // 保留错误链，稍慢
}

// 性能差异很小，优先选择功能需求

// 3. 避免不必要的错误创建
// ❌ 慢
func validate(s string) error {
    if s == "" {
        return fmt.Errorf("empty string") // 每次分配
    }
    return nil
}

// ✅ 快
var ErrEmptyString = errors.New("empty string")

func validate(s string) error {
    if s == "" {
        return ErrEmptyString // 复用
    }
    return nil
}

// 4. 热路径避免错误分配
type FastError struct {
    Code int
}

var errorPool = sync.Pool{
    New: func() interface{} {
        return &FastError{}
    },
}

func getError(code int) *FastError {
    err := errorPool.Get().(*FastError)
    err.Code = code
    return err
}

func releaseError(err *FastError) {
    err.Code = 0
    errorPool.Put(err)
}
```

---

### 考点10: Go 2 错误处理提案

**问题：** Go 2 对错误处理有什么改进计划？

**答案：**
```go
// 当前 Go 1.x 的问题：
// 1. 重复的 if err != nil 代码
// 2. 错误处理打断代码流程
// 3. 容易忘记检查错误

// Go 2 提案 1: check/handle（已废弃）
handle err {
    return fmt.Errorf("process failed: %w", err)
}

func process() error {
    data := check readFile()
    result := check parseData(data)
    check saveResult(result)
    return nil
}

// Go 2 提案 2: try（已废弃）
func process() error {
    data := try(readFile())
    result := try(parseData(data))
    try(saveResult(result))
    return nil
}

// 当前方案: 保持现状
// 社区共识：显式错误处理虽然繁琐，但清晰明确
// 未来可能的改进：
// - 改进工具链（如 gopls 自动生成错误处理代码）
// - 泛型支持后的新模式

// 实际建议：
// 1. 使用编辑器snippets加速错误处理编写
// 2. 使用 errcheck 等工具检查未处理的错误
// 3. 接受Go的错误处理哲学，写清晰的代码
```

---

## 总结：错误处理学习路线图

```
1️⃣ 基础阶段（1-2天）
   ├── 理解 error 接口
   ├── 掌握三种创建错误的方式
   ├── 记住错误处理黄金法则
   └── 练习基本的错误检查

2️⃣ 进阶阶段（2-3天）
   ├── 掌握错误包装（%w）
   ├── 使用 errors.Is/As
   ├── 设计自定义错误类型
   ├── 理解 panic/recover
   └── 练习错误链追踪

3️⃣ 实战阶段（3-5天）
   ├── 选择合适的错误处理模式
   ├── 处理并发场景的错误
   ├── 实现重试、熔断等机制
   └── 完成5道手撕代码

4️⃣ 面试阶段（1-2天）
   ├── 背诵10个高频考点
   ├── 理解底层实现原理
   ├── 准备实战案例
   └── 模拟面试练习
```

**关键记忆点：**
1. error 是接口，`Error() string`
2. 总是检查错误，只处理一次
3. `%w` 保留错误链，配合 `errors.Is/As`
4. panic 用于不可恢复的错误
5. 每个 goroutine 都要 recover
6. defer 确保资源清理
7. 哨兵错误性能最好
8. 并发场景用 errgroup 或 channel

**学习建议：**
- 每天练习2-3道手撕代码
- 阅读标准库的错误处理实现
- 在实际项目中应用所学知识
- 定期复习面试考点

---

**祝你在 Go 错误处理的学习中取得成功！** 🚀
