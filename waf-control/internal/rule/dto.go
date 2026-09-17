package rule

import "time"

type CreateRequest struct {
	RuleID   string `json:"rule_id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Source   string `json:"source"`
	Severity string `json:"severity"`
	Enabled  *bool  `json:"enabled"`
	Scope    string `json:"scope"`
}

type UpdateRequest struct {
	RuleID   *string `json:"rule_id"`
	Name     *string `json:"name"`
	Category *string `json:"category"`
	Source   *string `json:"source"`
	Severity *string `json:"severity"`
	Enabled  *bool   `json:"enabled"`
	Scope    *string `json:"scope"`
}

type Response struct {
	ID        string    `json:"id"`
	RuleID    string    `json:"rule_id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Source    string    `json:"source"`
	Severity  string    `json:"severity"`
	Enabled   bool      `json:"enabled"`
	Scope     string    `json:"scope"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToResponse(r *Rule) Response {
	return Response{
		ID: r.ID.String(), RuleID: r.RuleID, Name: r.Name, Category: r.Category,
		Source: r.Source, Severity: r.Severity, Enabled: r.Enabled, Scope: r.Scope,
		CreatedAt: r.CreatedAt.UTC(), UpdatedAt: r.UpdatedAt.UTC(),
	}
}

func ToResponseList(list []Rule) []Response {
	out := make([]Response, 0, len(list))
	for i := range list {
		out = append(out, ToResponse(&list[i]))
	}
	return out
}
