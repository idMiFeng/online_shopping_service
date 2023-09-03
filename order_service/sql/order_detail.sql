CREATE TABLE `xx_order_detail`(
                                 `id` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '主键',
                                 `create_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                 `create_by` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建者',
                                 `update_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
                                 `update_by` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建者',
                                 `version` SMALLINT(5) UNSIGNED NOT NULL DEFAULT '0' COMMENT '乐观锁版本号',
                                 `is_del` tinyint(4) UNSIGNED NOT NULL DEFAULT '0' COMMENT '是否删除：0正常1删除',

                                 `user_id` BIGINT(20) UNSIGNED NOT NULL COMMENT '用户id',
                                 `order_id` BIGINT(20) UNSIGNED NOT NULL COMMENT '订单id',
                                 `goods_id` BIGINT(20) UNSIGNED NOT NULL COMMENT '商品id',

                                 `title` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '名称',
                                 `market_price` BIGINT(20) UNSIGNED NOT NULL DEFAULT '0' COMMENT '市场价/划线价（分）',
                                 `price` BIGINT(20) UNSIGNED NOT NULL DEFAULT '0' COMMENT '售价（分）',
                                 `brief` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '简介',
                                 `head_imgs` VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '头图',
                                 `videos` VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '视频介绍',
                                 `detail` VARCHAR(2048) NOT NULL DEFAULT '' COMMENT '详情',
                                 `num` BIGINT(20) UNSIGNED NOT NULL COMMENT '商品数量',

                                 `pay_amount` BIGINT(20) UNSIGNED NOT NULL DEFAULT '0' COMMENT '支付金额（分）',
                                 `pay_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '支付时间',

                                 INDEX (order_id),
                                 INDEX (user_id),
                                 INDEX (is_del)
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4 COMMENT = '订单商品表';