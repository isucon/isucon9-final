package isutrain

type Station struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	IsStopExpress     bool   `json:"is_stop_express"`
	IsStopSemiExpress bool   `json:"is_stop_semi_express"`
	IsStopLocal       bool   `json:"is_stop_local"`
}
