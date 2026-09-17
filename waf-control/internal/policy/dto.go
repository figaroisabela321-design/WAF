package policy

import "time"

type CreateRequest struct {
	Name              string `json:"name"`
	Mode              string `json:"mode"`
	ProtectionLevel   string `json:"protection_level"`
	ParanoiaLevel     int    `json:"paranoia_level"`
	InboundThreshold  int    `json:"inbound_threshold"`
	OutboundThreshold int    `json:"outbound_threshold"`
	SQLInjection      *bool  `json:"sql_injection"`
	XSS               *bool  `json:"xss"`
	RCE               *bool  `json:"rce"`
	LFI               *bool  `json:"lfi"`
	Scanner           *bool  `json:"scanner"`
	ProtocolAttack    *bool  `json:"protocol_attack"`
}

type UpdateRequest struct {
	Name              *string `json:"name"`
	Mode              *string `json:"mode"`
	ProtectionLevel   *string `json:"protection_level"`
	ParanoiaLevel     *int    `json:"paranoia_level"`
	InboundThreshold  *int    `json:"inbound_threshold"`
	OutboundThreshold *int    `json:"outbound_threshold"`
	SQLInjection      *bool   `json:"sql_injection"`
	XSS               *bool   `json:"xss"`
	RCE               *bool   `json:"rce"`
	LFI               *bool   `json:"lfi"`
	Scanner           *bool   `json:"scanner"`
	ProtocolAttack    *bool   `json:"protocol_attack"`
}

type Response struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Mode              string    `json:"mode"`
	ProtectionLevel   string    `json:"protection_level"`
	ParanoiaLevel     int       `json:"paranoia_level"`
	InboundThreshold  int       `json:"inbound_threshold"`
	OutboundThreshold int       `json:"outbound_threshold"`
	SQLInjection      bool      `json:"sql_injection"`
	XSS               bool      `json:"xss"`
	RCE               bool      `json:"rce"`
	LFI               bool      `json:"lfi"`
	Scanner           bool      `json:"scanner"`
	ProtocolAttack    bool      `json:"protocol_attack"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func ToResponse(p *Policy) Response {
	return Response{
		ID: p.ID.String(), Name: p.Name, Mode: p.Mode, ProtectionLevel: p.ProtectionLevel,
		ParanoiaLevel: p.ParanoiaLevel, InboundThreshold: p.InboundThreshold,
		OutboundThreshold: p.OutboundThreshold, SQLInjection: p.SQLInjection, XSS: p.XSS,
		RCE: p.RCE, LFI: p.LFI, Scanner: p.Scanner, ProtocolAttack: p.ProtocolAttack,
		CreatedAt: p.CreatedAt.UTC(), UpdatedAt: p.UpdatedAt.UTC(),
	}
}

func ToResponseList(list []Policy) []Response {
	out := make([]Response, 0, len(list))
	for i := range list {
		out = append(out, ToResponse(&list[i]))
	}
	return out
}
