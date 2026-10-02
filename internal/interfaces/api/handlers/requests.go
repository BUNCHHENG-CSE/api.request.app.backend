package handlers

import "github.com/google/uuid"

type CreateCollectionRequest struct {
	WorkspaceID uuid.UUID `json:"workspace_id" binding:"required"`
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
}

type CreateFlowRequest struct {
	WorkspaceID uuid.UUID `json:"workspace_id" binding:"required"`
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	Nodes       string    `json:"nodes"`
	Edges       string    `json:"edges"`
}

type CreateStoredRequest struct {
	FlowID       uuid.UUID  `json:"flow_id" binding:"required"`
	CollectionID *uuid.UUID `json:"collection_id"`
	Name         string     `json:"name"`
	Position     int        `json:"position"`
	Method       string     `json:"method" binding:"required"`
	URL          string     `json:"url" binding:"required"`
	Headers      string     `json:"headers"`
	QueryParams  string     `json:"query_params"`
	Body         string     `json:"body"`
	BodyType     string     `json:"body_type"`
	Auth         string     `json:"auth"`
	Scripts      string     `json:"scripts"`
	Settings     string     `json:"settings"`
}

type PatchStoredRequest struct {
	FlowID      *uuid.UUID `json:"flow_id"`
	Name        *string    `json:"name"`
	Position    *int       `json:"position"`
	Method      *string    `json:"method"`
	URL         *string    `json:"url"`
	Headers     *string    `json:"headers"`
	QueryParams *string    `json:"query_params"`
	Body        *string    `json:"body"`
	BodyType    *string    `json:"body_type"`
	Auth        *string    `json:"auth"`
	Scripts     *string    `json:"scripts"`
	Settings    *string    `json:"settings"`
}
