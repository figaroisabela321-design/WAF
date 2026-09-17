package audit

import "time"

type Response struct {
	ID         string    `json:"id"`
	ActorID    *string   `json:"actor_id"`
	ActorName  string    `json:"actor_name"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id"`
	StatusCode int       `json:"status_code"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
}

func ToResponse(l *Log) Response {
	r := Response{
		ID: l.ID.String(), ActorName: l.ActorName, Method: l.Method, Path: l.Path,
		Resource: l.Resource, ResourceID: l.ResourceID, StatusCode: l.StatusCode,
		IP: l.IP, UserAgent: l.UserAgent, CreatedAt: l.CreatedAt.UTC(),
	}
	if l.ActorID != nil {
		v := l.ActorID.String()
		r.ActorID = &v
	}
	return r
}

func ToResponseList(list []Log) []Response {
	out := make([]Response, 0, len(list))
	for i := range list {
		out = append(out, ToResponse(&list[i]))
	}
	return out
}
