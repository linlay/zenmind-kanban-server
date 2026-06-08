INSERT INTO project (
	ID_, PARENT_ID_, SLUG_, KEY_, NAME_, DESCRIPTION_, PATH_, DEPTH_, POSITION_,
	VISIBILITY_, DEFAULT_WORKFLOW_ID_, CREATED_AT_, UPDATED_AT_
)
VALUES ('default', NULL, 'default', 'DEFAULT', 'All Projects', '', 'default', 0, 0, 'workspace', 'workflow-standard-requirement', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;

INSERT INTO project_closure (ANCESTOR_ID_, DESCENDANT_ID_, DEPTH_)
VALUES ('default', 'default', 0)
ON CONFLICT(ANCESTOR_ID_, DESCENDANT_ID_) DO NOTHING;

INSERT INTO board (ID_, PROJECT_ID_, KEY_, NAME_, CREATED_AT_, UPDATED_AT_)
VALUES ('default', 'default', 'default', 'Default Board', '__NOW__', '__NOW__')
ON CONFLICT(ID_) DO NOTHING;
