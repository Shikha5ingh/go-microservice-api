package models

// Plan represents the top-level plan object
type Plan struct {
	// ID                 string              `json:"id"`
	PlanCostShares     CostShare           `json:"planCostShares" binding:"required"`
	LinkedPlanServices []LinkedPlanService `json:"linkedPlanServices" binding:"required,dive,required"`
	Org                string              `json:"_org" binding:"required,uri"`
	ObjectID           string              `json:"objectId" binding:"required"`
	ObjectType         string              `json:"objectType" binding:"required,oneof=plan"`
	PlanType           string              `json:"planType" binding:"required,oneof=inNetwork outOfNetwork"`
	CreationDate       string              `json:"creationDate"`
}

// CostShare represents cost-sharing details
type CostShare struct {
	Deductible float64 `json:"deductible" binding:"required,min=0"`
	Copay      float64 `json:"copay" binding:"required,min=0"`
	Org        string  `json:"_org" binding:"required,uri"`
	ObjectID   string  `json:"objectId" binding:"required"`
	ObjectType string  `json:"objectType" binding:"required,oneof=membercostshare"`
}

// LinkedPlanService represents a linked service within a plan
type LinkedPlanService struct {
	LinkedService         LinkedService `json:"linkedService" binding:"required"`
	PlanserviceCostShares CostShare     `json:"planserviceCostShares" binding:"required"`
	Org                   string        `json:"_org" binding:"required,uri"`
	ObjectID              string        `json:"objectId" binding:"required"`
	ObjectType            string        `json:"objectType" binding:"required,oneof=planservice"`
}

// LinkedService represents the details of a linked service
type LinkedService struct {
	Org        string `json:"_org" binding:"required,uri"`
	ObjectID   string `json:"objectId" binding:"required"`
	ObjectType string `json:"objectType" binding:"required,oneof=service"`
	Name       string `json:"name" binding:"required"`
}
