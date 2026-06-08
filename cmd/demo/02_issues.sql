-- Clean old demo issues
DELETE FROM issue WHERE ID_ LIKE 'demo-issue-%';

INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-001', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '登录页响应式适配移动端', '', 'high', 'critical', 1, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-002', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '订单列表分页加载性能优化', '', 'medium', 'high', 2, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-003', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', '商品详情页图片懒加载', '', 'low', 'medium', 3, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-004', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', '购物车数量增减动画效果', '', 'high', 'low', 4, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-005', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', '用户头像上传裁剪功能', '', 'medium', 'critical', 5, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-006', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '首页 banner 轮播图自动播放', '', 'low', 'high', 6, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-007', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '搜索框联想词下拉列表', '', 'high', 'medium', 7, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-008', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', '个人中心修改密码安全性校验', '', 'medium', 'low', 8, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-009', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', '注册页面短信验证码倒计时', '', 'low', 'critical', 9, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-010', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', '支付结果页面的状态轮询', '', 'high', 'high', 10, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-011', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '消息通知红点数字标记', '', 'medium', 'medium', 11, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-012', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '下拉刷新加载更多内容', '', 'low', 'low', 12, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-013', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', '富文本编辑器工具栏布局', '', 'high', 'critical', 13, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-014', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', '评论列表嵌套回复展开收起', '', 'medium', 'high', 14, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-015', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', '多选删除批量操作确认框', '', 'low', 'medium', 15, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-016', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '表格列宽拖拽调整大小', '', 'high', 'low', 16, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-017', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '表单必填项红色星号提示', '', 'medium', 'critical', 17, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-018', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', '侧边栏菜单折叠展开动画', '', 'low', 'high', 18, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-019', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', '面包屑导航动态生成路径', '', 'high', 'medium', 19, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-020', 'demo-proj-001', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', '文件拖拽上传进度条展示', '', 'medium', 'low', 20, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-021', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '时间选择器日期范围联动', '', 'high', 'critical', 1, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-022', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '图表折线图数据点悬浮提示', '', 'medium', 'high', 2, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-023', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', '暗黑模式主题切换持久化', '', 'low', 'medium', 3, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-024', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', '多语言国际化文案抽取整理', '', 'high', 'low', 4, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-025', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', '水印防截图叠加层实现', '', 'medium', 'critical', 5, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-026', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '二维码生成与长按识别', '', 'low', 'high', 6, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-027', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '地图选点位置搜索周边', '', 'high', 'medium', 7, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-028', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', '语音输入转文字实时识别', '', 'medium', 'low', 8, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-029', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', '指纹/面容生物认证登录', '', 'low', 'critical', 9, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-030', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', '快捷方式桌面图标生成', '', 'high', 'high', 10, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-031', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '第三方 OAuth 授权回调处理', '', 'medium', 'medium', 11, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-032', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '接口请求超时重试策略优化', '', 'low', 'low', 12, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-033', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', 'WebSocket 断线重连心跳机制', '', 'high', 'critical', 13, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-034', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', '离线缓存数据同步冲突合并', '', 'medium', 'high', 14, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-035', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', '灰度发布流量百分比动态调整', '', 'low', 'medium', 15, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-036', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '慢 SQL 监控告警阈值配置', '', 'high', 'low', 16, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-037', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '日志采集链路追踪采样率', '', 'medium', 'critical', 17, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-038', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', '数据库连接池大小自适应', '', 'low', 'high', 18, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-039', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', 'Redis 热 key 自动发现迁移', '', 'high', 'medium', 19, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-040', 'demo-proj-009', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', '消息队列积压消费者扩容', '', 'medium', 'low', 20, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-041', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '服务降级熔断半开恢复策略', '', 'high', 'critical', 1, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-042', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', 'API 限流令牌桶算法实现', '', 'medium', 'high', 2, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-043', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', '幂等性去重防重复提交', '', 'low', 'medium', 3, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-044', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', '分布式锁续期 watchdog 机制', '', 'high', 'low', 4, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-045', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', '配置热更新不重启生效', '', 'medium', 'critical', 5, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-046', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '定时任务分片并行执行', '', 'low', 'high', 6, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-047', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '数据脱敏敏感字段掩码规则', '', 'high', 'medium', 7, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-048', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', '加密传输国密 SM4 算法支持', '', 'medium', 'low', 8, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-049', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', '审计日志操作留痕不可篡改', '', 'low', 'critical', 9, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-050', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', '白名单 IP 访问控制策略', '', 'high', 'high', 10, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-051', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '单点登录 Session 共享方案', '', 'medium', 'medium', 11, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-052', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '跨域 CORS 预检请求缓存', '', 'low', 'low', 12, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-053', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', '文件分片断点续传合并校验', '', 'high', 'critical', 13, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-054', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', 'PDF 预览水印页码渲染', '', 'medium', 'high', 14, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-055', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', 'Excel 导入导出大数据量优化', '', 'low', 'medium', 15, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-056', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '邮件模板变量替换发送队列', '', 'high', 'low', 16, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-057', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '站内信已读未读状态管理', '', 'medium', 'critical', 17, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-058', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', '操作日志按天归档清理策略', '', 'low', 'high', 18, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-059', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', '版本发布 Changelog 自动生成', '', 'high', 'medium', 19, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-060', 'demo-proj-013', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', '自动化测试用例覆盖率报表', '', 'medium', 'low', 20, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-061', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '代码扫描 Sonar 规则定制', '', 'high', 'critical', 1, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-062', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '容器镜像分层构建缓存优化', '', 'medium', 'high', 2, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-063', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', 'K8s 滚动更新就绪探针配置', '', 'low', 'medium', 3, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-064', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', '负载均衡会话保持一致性哈希', '', 'high', 'low', 4, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-065', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', 'CDN 缓存预热刷新策略', '', 'medium', 'critical', 5, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-066', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '数据库主从切换故障转移', '', 'low', 'high', 6, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-067', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '备份恢复定期演练自动化', '', 'high', 'medium', 7, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-068', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', '灰度 A/B 实验分流规则引擎', '', 'medium', 'low', 8, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-069', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', '移动端推送长连接保活优化', '', 'low', 'critical', 9, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-070', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', '应用内升级强制更新弹窗', '', 'high', 'high', 10, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-071', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '崩溃日志符号化堆栈解析', '', 'medium', 'medium', 11, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-072', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '性能埋点启动耗时阶段分析', '', 'low', 'low', 12, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-073', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', '内存泄漏循环引用检测工具', '', 'high', 'critical', 13, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-074', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', '电量优化后台任务合并调度', '', 'medium', 'high', 14, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-075', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', '网络库 HTTP/3 QUIC 协议适配', '', 'low', 'medium', 15, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-076', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '视频播放器预加载缓冲策略', '', 'high', 'low', 16, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-077', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '直播推流画质自适应码率', '', 'medium', 'critical', 17, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-078', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', '即时通讯消息已读回执', '', 'low', 'high', 18, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-079', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', '社交动态时间线 Feed 流聚合', '', 'high', 'medium', 19, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-080', 'demo-proj-017', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', '好友推荐共同联系人算法', '', 'medium', 'low', 20, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-081', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '附近的人 GeoHash 经纬度编码', '', 'high', 'critical', 1, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-082', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '红包雨高并发抢购队列设计', '', 'medium', 'high', 2, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-083', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', '秒杀库存扣减 Redis Lua 原子操作', '', 'low', 'medium', 3, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-084', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', '优惠券过期提醒定时扫描', '', 'high', 'low', 4, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-085', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', '积分商城兑换物流跟踪', '', 'medium', 'critical', 5, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-086', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '拼团活动成团超时自动退款', '', 'low', 'high', 6, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-087', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '商品评价图片审核敏感词过滤', '', 'high', 'medium', 7, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-088', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', '售后退款原路返回到账时效', '', 'medium', 'low', 8, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-089', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', '发票抬头智能识别 OCR 提取', '', 'low', 'critical', 9, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-090', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', '会员等级升降规则引擎计算', '', 'high', 'high', 10, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-091', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '签到日历连续天数补签卡', '', 'medium', 'medium', 11, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-092', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '数据导出异步任务进度查询', '', 'low', 'low', 12, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-093', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', '短信通道自动切换故障转移', '', 'high', 'critical', 13, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-094', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', '第三方回调验签防伪造重放', '', 'medium', 'high', 14, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-095', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', '接口文档 Swagger 自动生成', '', 'low', 'medium', 15, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-096', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', 'CI/CD 构建产物版本号注入', '', 'high', 'low', 16, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-097', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '域名切换灰度 DNS 权重调整', '', 'medium', 'critical', 17, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-098', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_progress', '服务注册发现健康检查剔除', '', 'low', 'high', 18, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-099', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-in_review', '链路追踪 Span 上下文透传', '', 'high', 'medium', 19, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO issue (ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_, TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_, REVIEW_REQUIRED_, REVISION_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-issue-100', 'demo-proj-021', 'workflow-standard-requirement', 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-completed', '日志级别运行时动态调整', '', 'medium', 'low', 20, 0, 0, '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
