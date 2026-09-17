package node

import "time"

type CreateRequest struct {
	Name          string `json:"name"`
	IP            string `json:"ip"`
	Region        string `json:"region"`
	NodeGroup     string `json:"node_group"`
	CorazaVersion string `json:"coraza_version"`
	CRSVersion    string `json:"crs_version"`
	ConfigVersion string `json:"config_version"`
	Status        string `json:"status"`
}

type UpdateRequest struct {
	Name          *string `json:"name"`
	IP            *string `json:"ip"`
	Region        *string `json:"region"`
	NodeGroup     *string `json:"node_group"`
	CorazaVersion *string `json:"coraza_version"`
	CRSVersion    *string `json:"crs_version"`
	ConfigVersion *string `json:"config_version"`
	Status        *string `json:"status"`
}

type Response struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	IP            string     `json:"ip"`
	Region        string     `json:"region"`
	NodeGroup     string     `json:"node_group"`
	CorazaVersion string     `json:"coraza_version"`
	CRSVersion    string     `json:"crs_version"`
	ConfigVersion string     `json:"config_version"`
	Status        string     `json:"status"`
	LastHeartbeat *time.Time `json:"last_heartbeat"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func ToResponse(n *Node) Response {
	r := Response{
		ID: n.ID.String(), Name: n.Name, IP: n.IP, Region: n.Region,
		NodeGroup: n.NodeGroup, CorazaVersion: n.CorazaVersion, CRSVersion: n.CRSVersion,
		ConfigVersion: n.ConfigVersion, Status: n.Status,
		CreatedAt: n.CreatedAt.UTC(), UpdatedAt: n.UpdatedAt.UTC(),
	}
	if n.LastHeartbeat != nil {
		t := n.LastHeartbeat.UTC()
		r.LastHeartbeat = &t
	}
	return r
}

func ToResponseList(nodes []Node) []Response {
	out := make([]Response, 0, len(nodes))
	for i := range nodes {
		out = append(out, ToResponse(&nodes[i]))
	}
	return out
}
