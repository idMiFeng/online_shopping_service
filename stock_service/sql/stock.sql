CREATE TABLE `xx_stock`(
                           `id` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '主键',
                           `create_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                           `create_by` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建者',
                           `update_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
                           `update_by` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建者',
                           `version` SMALLINT(5) UNSIGNED NOT NULL DEFAULT '0' COMMENT '乐观锁版本号',
                           `is_del` tinyint(4) UNSIGNED NOT NULL DEFAULT '0' COMMENT '是否删除：0正常1删除',

                           `goods_id` BIGINT(20) UNSIGNED NOT NULL DEFAULT '0' COMMENT 'goods id',
                           `num` BIGINT(20) UNSIGNED NOT NULL DEFAULT '0' COMMENT '库存',
                           `lock` BIGINT(20) UNSIGNED NOT NULL DEFAULT '0' COMMENT '预扣库存',
                           UNIQUE (goods_id),
                           INDEX (is_del)
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4 COMMENT = '库存表';