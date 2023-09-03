# online_shopping_service
## 基于微服务架构的购物服务demo
三个服务：商品微服务 库存微服务 订单微服务
### 项目依赖

1. MySQL
   1. 建库建表
   2. 测试数据
   3. 不同的服务可以用不同的数据库

2. Redis
   1. 库存服务分布式锁

3. Consul
   1. 服务注册与服务发现
   2. 程序启动时建立负载均衡连接

4. RocketMQ
   1. 事务消息（扣减库存）
   2. 延迟消息（超时未支付取消）

#### 进到每个微服务执行下述命令，访问locolhost:8500可以在consul控制台看到注册的微服务
   ```bash
   go mod tidy
   go run main.go
   ```
   
