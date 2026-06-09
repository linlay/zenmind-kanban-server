-- Clean old demo issues
DELETE FROM issue WHERE ID_ LIKE 'demo-issue-%' OR CREATED_BY_ = 'demo' OR UPDATED_BY_ = 'demo';

WITH RECURSIVE
seq(n) AS (
	VALUES (1)
	UNION ALL
	SELECT n + 1 FROM seq WHERE n < 200
),
base36_chars(chars) AS (
	VALUES ('0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ')
),
issue_ids(n, issue_id) AS (
	SELECT
		seq.n,
		'DEMO' ||
			CASE
				WHEN seq.n < 36 THEN substr(base36_chars.chars, seq.n + 1, 1)
				ELSE substr(base36_chars.chars, (seq.n / 36) + 1, 1) || substr(base36_chars.chars, (seq.n % 36) + 1, 1)
			END
	FROM seq
	CROSS JOIN base36_chars
),
project_cycle(idx, project_id) AS (
	VALUES
		(0, 'demo-proj-001'),
		(1, 'demo-proj-003'),
		(2, 'demo-proj-006'),
		(3, 'demo-proj-009'),
		(4, 'demo-proj-013'),
		(5, 'demo-proj-014'),
		(6, 'demo-proj-017'),
		(7, 'demo-proj-018'),
		(8, 'demo-proj-021'),
		(9, 'demo-proj-022')
),
status_cycle(idx, stage_id, status_id, title, review_required) AS (
	VALUES
		(0, 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-backlog', '梳理需求入口和验收边界', 0),
		(1, 'workflow-standard-requirement-stage-requirement_clarification', 'workflow-standard-requirement-status-todo', '补充业务规则和优先级', 0),
		(2, 'workflow-standard-requirement-stage-development', 'workflow-standard-requirement-status-in_progress', '实现核心流程和接口联调', 0),
		(3, 'workflow-standard-requirement-stage-development', 'workflow-standard-requirement-status-in_review', '提交代码审查和联调记录', 1),
		(4, 'workflow-standard-requirement-stage-solution_design', 'workflow-standard-requirement-status-waiting_answer', '等待产品回答方案疑问', 1),
		(5, 'workflow-standard-requirement-stage-solution_design', 'workflow-standard-requirement-status-waiting_submit', '等待提交设计说明', 1),
		(6, 'workflow-standard-requirement-stage-solution_design', 'workflow-standard-requirement-status-waiting_approval', '等待负责人批准方案', 1),
		(7, 'workflow-standard-requirement-stage-development', 'workflow-standard-requirement-status-success', '开发任务成功收尾', 0),
		(8, 'workflow-standard-requirement-stage-development', 'workflow-standard-requirement-status-failed', '开发任务失败复盘', 0),
		(9, 'workflow-standard-requirement-stage-development', 'workflow-standard-requirement-status-interrupted', '开发任务中断挂起', 0),
		(10, 'workflow-standard-requirement-stage-testing_acceptance', 'workflow-standard-requirement-status-testing_waiting_submit', '测试阶段等待提交报告', 1),
		(11, 'workflow-standard-requirement-stage-testing_acceptance', 'workflow-standard-requirement-status-testing_in_progress', '测试阶段执行回归用例', 0),
		(12, 'workflow-standard-requirement-stage-testing_acceptance', 'workflow-standard-requirement-status-testing_waiting_approval', '测试阶段等待验收批准', 1),
		(13, 'workflow-standard-requirement-stage-testing_acceptance', 'workflow-standard-requirement-status-testing_success', '测试阶段通过验收', 0),
		(14, 'workflow-standard-requirement-stage-testing_acceptance', 'workflow-standard-requirement-status-testing_failed', '测试阶段发现阻塞缺陷', 0),
		(15, 'workflow-standard-requirement-stage-testing_acceptance', 'workflow-standard-requirement-status-testing_interrupted', '测试阶段因环境中断', 0),
		(16, 'workflow-standard-requirement-stage-release', 'workflow-standard-requirement-status-deployment_waiting_submit', '部署阶段等待发布申请', 1),
		(17, 'workflow-standard-requirement-stage-release', 'workflow-standard-requirement-status-deployment_in_progress', '部署阶段执行发布流程', 0),
		(18, 'workflow-standard-requirement-stage-release', 'workflow-standard-requirement-status-deployment_waiting_approval', '部署阶段等待上线批准', 1),
		(19, 'workflow-standard-requirement-stage-release', 'workflow-standard-requirement-status-deployment_success', '部署阶段发布成功', 0),
		(20, 'workflow-standard-requirement-stage-release', 'workflow-standard-requirement-status-deployment_failed', '部署阶段发布失败', 0),
		(21, 'workflow-standard-requirement-stage-release', 'workflow-standard-requirement-status-deployment_interrupted', '部署阶段发布中断', 0),
		(22, 'workflow-standard-requirement-stage-release', 'workflow-standard-requirement-status-completed', '完成归档并同步结果', 0)
)
INSERT INTO issue (
	ID_, PROJECT_ID_, WORKFLOW_ID_, STAGE_ID_, STATUS_ID_,
	TITLE_, DESCRIPTION_, PRIORITY_, SEVERITY_, POSITION_,
	REVIEW_REQUIRED_, REVISION_, CREATED_BY_, UPDATED_BY_, CREATED_AT_, UPDATED_AT_
)
SELECT
	issue_ids.issue_id,
	project_cycle.project_id,
	'workflow-standard-requirement',
	status_cycle.stage_id,
	status_cycle.status_id,
	CASE
		WHEN seq.n % 13 = 0 THEN status_cycle.title || '：这是一个用于验证长标题完整换行显示的演示 issue，标题故意写得很长以超过两行，并保留阶段标签、细分状态、负责人上下文和关键验收条件'
		WHEN seq.n % 11 = 0 THEN status_cycle.title || '：请同时检查 status、statusKey、statusName 展示，审批等待文案、失败中断成功结果分支，以及跨项目筛选后的卡片排序是否仍然稳定'
		WHEN seq.n % 8 = 0 THEN status_cycle.title || '：覆盖移动端窄屏、桌面端看板列拖拽、状态细分映射、历史数据兼容和异常恢复路径，确保负责人可以直接读完整上下文'
		ELSE status_cycle.title
	END || ' #' || issue_ids.issue_id,
	'',
	CASE seq.n % 3
		WHEN 0 THEN 'high'
		WHEN 1 THEN 'medium'
		ELSE 'low'
	END,
	CASE seq.n % 4
		WHEN 0 THEN 'critical'
		WHEN 1 THEN 'high'
		WHEN 2 THEN 'medium'
		ELSE 'low'
	END,
	CAST(((seq.n - 1) / 23) AS REAL) * 100 + status_cycle.idx + 1,
	status_cycle.review_required,
	0,
	'demo',
	'demo',
	'__NOW__',
	'__NOW__'
FROM seq
JOIN issue_ids ON issue_ids.n = seq.n
JOIN project_cycle ON project_cycle.idx = ((seq.n - 1) % 10)
JOIN status_cycle ON status_cycle.idx = ((seq.n - 1) % 23);
