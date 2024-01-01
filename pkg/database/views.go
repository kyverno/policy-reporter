package database

import "github.com/uptrace/bun"

type Category struct {
	bun.BaseModel `bun:"table:policy_report_filter,alias:f"`

	Source string
	Name   string `bun:"category"`
	Result string
	Count  int
}

type ResourceCategory struct {
	bun.BaseModel `bun:"table:policy_report_resource,alias:res"`

	Source string
	Name   string `bun:"category"`
	Pass   int
	Warn   int
	Fail   int
	Error  int
	Skip   int
}

type ResourceStatusCount struct {
	bun.BaseModel `bun:"table:policy_report_resource,alias:res"`
	Source        string
	Pass          int
	Warn          int
	Fail          int
	Error         int
	Skip          int
}

type StatusCount struct {
	bun.BaseModel `bun:"table:policy_report_filter,alias:f"`

	Source    string
	Namespace string `bun:"resource_namespace"`
	Status    string
	Count     int
}
