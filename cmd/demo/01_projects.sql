-- Clean existing demo data
DELETE FROM issue WHERE PROJECT_ID_ LIKE 'demo-proj-%';
DELETE FROM project_closure WHERE ANCESTOR_ID_ LIKE 'demo-proj-%';
DELETE FROM project WHERE ID_ LIKE 'demo-proj-%';

INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-001', NULL, '电商平台重构', '电商平台重构', '电商平台重构', '', '电商平台重构', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-001', 'demo-proj-001', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-002', 'demo-proj-001', '前端商城', '前端商城', '前端商城', '', '电商平台重构/前端商城', 2, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-002', 'demo-proj-002', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
SELECT ANCESTOR_ID_, 'demo-proj-002', DEPTH_ + 1
FROM project_closure
WHERE DESCENDANT_ID_ = 'demo-proj-001';
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-003', 'demo-proj-002', '商品详情页', '商品详情页', '商品详情页', '', '电商平台重构/前端商城/商品详情页', 3, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-003', 'demo-proj-003', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
SELECT ANCESTOR_ID_, 'demo-proj-003', DEPTH_ + 1
FROM project_closure
WHERE DESCENDANT_ID_ = 'demo-proj-002';
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-004', 'demo-proj-002', '购物车模块', '购物车模块', '购物车模块', '', '电商平台重构/前端商城/购物车模块', 3, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-004', 'demo-proj-004', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
SELECT ANCESTOR_ID_, 'demo-proj-004', DEPTH_ + 1
FROM project_closure
WHERE DESCENDANT_ID_ = 'demo-proj-002';
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-005', 'demo-proj-001', '后端服务', '后端服务', '后端服务', '', '电商平台重构/后端服务', 2, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-005', 'demo-proj-005', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
SELECT ANCESTOR_ID_, 'demo-proj-005', DEPTH_ + 1
FROM project_closure
WHERE DESCENDANT_ID_ = 'demo-proj-001';
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-006', 'demo-proj-005', '订单系统', '订单系统', '订单系统', '', '电商平台重构/后端服务/订单系统', 3, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-006', 'demo-proj-006', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
SELECT ANCESTOR_ID_, 'demo-proj-006', DEPTH_ + 1
FROM project_closure
WHERE DESCENDANT_ID_ = 'demo-proj-005';
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-007', 'demo-proj-005', '支付网关', '支付网关', '支付网关', '', '电商平台重构/后端服务/支付网关', 3, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-007', 'demo-proj-007', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
SELECT ANCESTOR_ID_, 'demo-proj-007', DEPTH_ + 1
FROM project_closure
WHERE DESCENDANT_ID_ = 'demo-proj-005';
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-008', 'demo-proj-001', '用户中心', '用户中心', '用户中心', '', '电商平台重构/用户中心', 2, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-008', 'demo-proj-008', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
SELECT ANCESTOR_ID_, 'demo-proj-008', DEPTH_ + 1
FROM project_closure
WHERE DESCENDANT_ID_ = 'demo-proj-001';
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-009', NULL, '数据中台建设', '数据中台建设', '数据中台建设', '', '数据中台建设', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-009', 'demo-proj-009', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-010', 'demo-proj-009', '数据采集层', '数据采集层', '数据采集层', '', '数据中台建设/数据采集层', 2, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-010', 'demo-proj-010', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
SELECT ANCESTOR_ID_, 'demo-proj-010', DEPTH_ + 1
FROM project_closure
WHERE DESCENDANT_ID_ = 'demo-proj-009';
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-011', 'demo-proj-009', '数据治理', '数据治理', '数据治理', '', '数据中台建设/数据治理', 2, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-011', 'demo-proj-011', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
SELECT ANCESTOR_ID_, 'demo-proj-011', DEPTH_ + 1
FROM project_closure
WHERE DESCENDANT_ID_ = 'demo-proj-009';
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-012', 'demo-proj-009', '数据服务', '数据服务', '数据服务', '', '数据中台建设/数据服务', 2, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-012', 'demo-proj-012', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
SELECT ANCESTOR_ID_, 'demo-proj-012', DEPTH_ + 1
FROM project_closure
WHERE DESCENDANT_ID_ = 'demo-proj-009';
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-013', NULL, '智能客服机器人', '智能客服机器人', '智能客服机器人', '', '智能客服机器人', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-013', 'demo-proj-013', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-014', 'demo-proj-013', 'nlp-引擎', 'NLP-引擎', 'NLP 引擎', '', '智能客服机器人/nlp-引擎', 2, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-014', 'demo-proj-014', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
SELECT ANCESTOR_ID_, 'demo-proj-014', DEPTH_ + 1
FROM project_closure
WHERE DESCENDANT_ID_ = 'demo-proj-013';
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-015', 'demo-proj-013', '对话管理', '对话管理', '对话管理', '', '智能客服机器人/对话管理', 2, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-015', 'demo-proj-015', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
SELECT ANCESTOR_ID_, 'demo-proj-015', DEPTH_ + 1
FROM project_closure
WHERE DESCENDANT_ID_ = 'demo-proj-013';
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-016', 'demo-proj-013', '知识库', '知识库', '知识库', '', '智能客服机器人/知识库', 2, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-016', 'demo-proj-016', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
SELECT ANCESTOR_ID_, 'demo-proj-016', DEPTH_ + 1
FROM project_closure
WHERE DESCENDANT_ID_ = 'demo-proj-013';
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-017', NULL, '移动端-app-改版', '移动端-APP-改版', '移动端 App 改版', '', '移动端-app-改版', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-017', 'demo-proj-017', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-018', 'demo-proj-017', 'ios-客户端', 'IOS-客户端', 'iOS 客户端', '', '移动端-app-改版/ios-客户端', 2, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-018', 'demo-proj-018', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
SELECT ANCESTOR_ID_, 'demo-proj-018', DEPTH_ + 1
FROM project_closure
WHERE DESCENDANT_ID_ = 'demo-proj-017';
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-019', 'demo-proj-017', 'android-客户端', 'ANDROID-客户端', 'Android 客户端', '', '移动端-app-改版/android-客户端', 2, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-019', 'demo-proj-019', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
SELECT ANCESTOR_ID_, 'demo-proj-019', DEPTH_ + 1
FROM project_closure
WHERE DESCENDANT_ID_ = 'demo-proj-017';
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-020', 'demo-proj-017', '小程序端', '小程序端', '小程序端', '', '移动端-app-改版/小程序端', 2, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-020', 'demo-proj-020', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
SELECT ANCESTOR_ID_, 'demo-proj-020', DEPTH_ + 1
FROM project_closure
WHERE DESCENDANT_ID_ = 'demo-proj-017';
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-021', NULL, 'devops-流水线', 'DEVOPS-流水线', 'DevOps 流水线', '', 'devops-流水线', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-021', 'demo-proj-021', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-022', 'demo-proj-021', 'ci-cd-平台', 'CI-CD-平台', 'CI/CD 平台', '', 'devops-流水线/ci-cd-平台', 2, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-022', 'demo-proj-022', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
SELECT ANCESTOR_ID_, 'demo-proj-022', DEPTH_ + 1
FROM project_closure
WHERE DESCENDANT_ID_ = 'demo-proj-021';
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-023', 'demo-proj-021', '监控告警', '监控告警', '监控告警', '', 'devops-流水线/监控告警', 2, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-023', 'demo-proj-023', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
SELECT ANCESTOR_ID_, 'demo-proj-023', DEPTH_ + 1
FROM project_closure
WHERE DESCENDANT_ID_ = 'demo-proj-021';
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-024', 'demo-proj-021', '日志中心', '日志中心', '日志中心', '', 'devops-流水线/日志中心', 2, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-024', 'demo-proj-024', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
SELECT ANCESTOR_ID_, 'demo-proj-024', DEPTH_ + 1
FROM project_closure
WHERE DESCENDANT_ID_ = 'demo-proj-021';
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-025', NULL, 'oa-办公自动化', 'OA-办公自动化', 'OA 办公自动化', '', 'oa-办公自动化', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-025', 'demo-proj-025', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-026', NULL, '风控决策引擎', '风控决策引擎', '风控决策引擎', '', '风控决策引擎', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-026', 'demo-proj-026', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-027', NULL, '推荐系统优化', '推荐系统优化', '推荐系统优化', '', '推荐系统优化', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-027', 'demo-proj-027', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-028', NULL, '物流调度平台', '物流调度平台', '物流调度平台', '', '物流调度平台', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-028', 'demo-proj-028', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-029', NULL, '财务对账系统', '财务对账系统', '财务对账系统', '', '财务对账系统', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-029', 'demo-proj-029', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-030', NULL, '内容管理-cms', '内容管理-CMS', '内容管理 CMS', '', '内容管理-cms', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-030', 'demo-proj-030', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-031', NULL, '直播带货平台', '直播带货平台', '直播带货平台', '', '直播带货平台', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-031', 'demo-proj-031', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-032', NULL, '在线教育门户', '在线教育门户', '在线教育门户', '', '在线教育门户', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-032', 'demo-proj-032', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-033', NULL, '医疗预约挂号', '医疗预约挂号', '医疗预约挂号', '', '医疗预约挂号', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-033', 'demo-proj-033', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-034', NULL, '智慧园区管理', '智慧园区管理', '智慧园区管理', '', '智慧园区管理', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-034', 'demo-proj-034', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-035', NULL, '物联网设备平台', '物联网设备平台', '物联网设备平台', '', '物联网设备平台', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-035', 'demo-proj-035', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-036', NULL, '游戏运营后台', '游戏运营后台', '游戏运营后台', '', '游戏运营后台', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-036', 'demo-proj-036', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-037', NULL, '广告投放系统', '广告投放系统', '广告投放系统', '', '广告投放系统', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-037', 'demo-proj-037', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-038', NULL, '第三方支付网关', '第三方支付网关', '第三方支付网关', '', '第三方支付网关', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-038', 'demo-proj-038', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-039', NULL, '短信通道聚合', '短信通道聚合', '短信通道聚合', '', '短信通道聚合', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-039', 'demo-proj-039', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-040', NULL, '电子合同签署', '电子合同签署', '电子合同签署', '', '电子合同签署', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-040', 'demo-proj-040', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-041', NULL, '代码审查平台', '代码审查平台', '代码审查平台', '', '代码审查平台', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-041', 'demo-proj-041', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-042', NULL, 'api-网关升级', 'API-网关升级', 'API 网关升级', '', 'api-网关升级', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-042', 'demo-proj-042', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-043', NULL, '配置中心重构', '配置中心重构', '配置中心重构', '', '配置中心重构', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-043', 'demo-proj-043', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-044', NULL, '分布式任务调度', '分布式任务调度', '分布式任务调度', '', '分布式任务调度', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-044', 'demo-proj-044', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-045', NULL, '消息推送中台', '消息推送中台', '消息推送中台', '', '消息推送中台', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-045', 'demo-proj-045', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-046', NULL, '用户画像标签', '用户画像标签', '用户画像标签', '', '用户画像标签', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-046', 'demo-proj-046', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-047', NULL, 'a-b-实验平台', 'A-B-实验平台', 'A/B 实验平台', '', 'a-b-实验平台', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-047', 'demo-proj-047', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-048', NULL, '工单流转系统', '工单流转系统', '工单流转系统', '', '工单流转系统', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-048', 'demo-proj-048', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-049', NULL, '远程桌面网关', '远程桌面网关', '远程桌面网关', '', '远程桌面网关', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-049', 'demo-proj-049', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
INSERT INTO project (ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_, VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_)
VALUES ('demo-proj-050', NULL, '视频转码流水线', '视频转码流水线', '视频转码流水线', '', '视频转码流水线', 1, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('demo-proj-050', 'demo-proj-050', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;
