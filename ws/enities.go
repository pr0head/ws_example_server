package ws

type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type SetServerStatus struct {
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

type AddGameChar struct {
	UserId     string `json:"user_id"`
	ServerName string `json:"server_name"`
	CharId     string `json:"char_id"`
	CharName   string `json:"char_name"`
}

type SendGameBalance struct {
	UserId     string        `json:"user_id"`
	ServerName string        `json:"server_name"`
	Tokens     []*GameTokens `json:"tokens"`
}

type GameTokens struct {
	Id     string  `json:"id"`
	Amount float64 `json:"amount"`
}

type GetGameBalance struct {
	UserId     string `json:"user_id"`
	ServerName string `json:"server_name"`
}
