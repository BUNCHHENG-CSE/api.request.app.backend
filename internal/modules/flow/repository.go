package flow

type Repository interface {
	CreateFlow(workspaceID string, flowName string)
}
