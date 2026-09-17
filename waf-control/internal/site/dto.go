package site

import "time"

// CreateRequest is the DTO for creating a site.
type CreateRequest struct {
	Name           string  `json:"name"`
	Domain         string  `json:"domain"`
	UpstreamHost   string  `json:"upstream_host"`
	UpstreamPort   int     `json:"upstream_port"`
	Protocol       string  `json:"protocol"`
	ProtectionMode string  `json:"protection_mode"`
	PolicyID       *string `json:"policy_id"`
	NodeGroupID    *string `json:"node_group_id"`
	Status         string  `json:"status"`
}

// UpdateRequest is the DTO for updating a site.
type UpdateRequest struct {
	Name           *string `json:"name"`
	Domain         *string `json:"domain"`
	UpstreamHost   *string `json:"upstream_host"`
	UpstreamPort   *int    `json:"upstream_port"`
	Protocol       *string `json:"protocol"`
	ProtectionMode *string `json:"protection_mode"`
	PolicyID       *string `json:"policy_id"`
	NodeGroupID    *string `json:"node_group_id"`
	Status         *string `json:"status"`
}

// Response is the API response DTO.
type Response struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Domain         string    `json:"domain"`
	UpstreamHost   string    `json:"upstream_host"`
	UpstreamPort   int       `json:"upstream_port"`
	Protocol       string    `json:"protocol"`
	ProtectionMode string    `json:"protection_mode"`
	PolicyID       *string   `json:"policy_id"`
	NodeGroupID    *string   `json:"node_group_id"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func ToResponse(s *Site) Response {
	r := Response{
		ID:             s.ID.String(),
		Name:           s.Name,
		Domain:         s.Domain,
		UpstreamHost:   s.UpstreamHost,
		UpstreamPort:   s.UpstreamPort,
		Protocol:       s.Protocol,
		ProtectionMode: s.ProtectionMode,
		Status:         s.Status,
		CreatedAt:      s.CreatedAt.UTC(),
		UpdatedAt:      s.UpdatedAt.UTC(),
	}
	if s.PolicyID != nil {
		v := s.PolicyID.String()
		r.PolicyID = &v
	}
	if s.NodeGroupID != nil {
		v := s.NodeGroupID.String()
		r.NodeGroupID = &v
	}
	return r
}

func ToResponseList(sites []Site) []Response {
	out := make([]Response, 0, len(sites))
	for i := range sites {
		out = append(out, ToResponse(&sites[i]))
	}
	return out
}
