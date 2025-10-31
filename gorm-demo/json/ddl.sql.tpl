create table approval (
  id bigint unsigned auto_increment primary key,
  created_at datetime null,
  updated_at datetime null,
  deleted_at datetime null,
  instance_id varchar(255) not null comment '审批实例 ID, 飞书: uuid',
  approval_code varchar(255) not null comment '审批实例 Code, 即表单定义的唯一 ID, 飞书: approval_code',
  type varchar(20) not null comment '审批实例类型, 可选值: lark, dingtalk',
  is_written_es tinyint(1) not null default 0 comment '是否已写入ES：0-未写入，1-已写入',
  -- 这里把分号改成逗号
  lark_data json not null comment '单个飞书审批实例数据',
  -- 为 instance_id 创建唯一索引
  constraint uk_instance_id unique (instance_id)
);
