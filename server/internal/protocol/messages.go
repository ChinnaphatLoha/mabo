package protocol

type ErrorCode string

const (
	ErrorInvalidRoom      ErrorCode = "INVALID_ROOM"
	ErrorRoomFull         ErrorCode = "ROOM_FULL"
	ErrorDuplicateJoin    ErrorCode = "DUPLICATE_JOIN"
	ErrorAlreadyConnected ErrorCode = "ALREADY_CONNECTED"
	ErrorDisconnected     ErrorCode = "DISCONNECTED"
	ErrorInvalidPacket    ErrorCode = "INVALID_PACKET"
)

type LoginRequest struct {
	GuestName string `json:"guest_name,omitempty"`
}

type LoginResponse struct {
	SessionID string `json:"session_id"`
	PlayerID  string `json:"player_id"`
}

type CreateRoomRequest struct {
	Capacity int `json:"capacity,omitempty"`
}

type JoinRoomRequest struct {
	RoomID string `json:"room_id"`
}

type LeaveRoomRequest struct {
	RoomID string `json:"room_id"`
}

type MovementInput struct {
	Sequence uint32  `json:"sequence"`
	MoveX    float64 `json:"move_x"`
	MoveY    float64 `json:"move_y"`
	Rotation float64 `json:"rotation"`
}

type RoomResponse struct {
	RoomID   string   `json:"room_id"`
	PlayerID string   `json:"player_id"`
	Players  []string `json:"players"`
	Capacity int      `json:"capacity"`
	State    string   `json:"state"`
}

type Vec2 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type PlayerSpawned struct {
	MatchID        string `json:"match_id"`
	RoomID         string `json:"room_id"`
	EntityID       string `json:"entity_id"`
	PlayerID       string `json:"player_id"`
	Team           int    `json:"team"`
	Position       Vec2   `json:"position"`
	Rotation       float64 `json:"rotation"`
	HP             int    `json:"hp"`
	AnimationState string `json:"animation_state"`
}

type PlayerDisconnected struct {
	RoomID   string `json:"room_id"`
	MatchID  string `json:"match_id"`
	PlayerID string `json:"player_id"`
}

type Snapshot struct {
	Tick    uint64           `json:"tick"`
	MatchID string           `json:"match_id"`
	Players []SnapshotPlayer `json:"players"`
}

type SnapshotPlayer struct {
	PlayerID       string `json:"player_id"`
	EntityID       string `json:"entity_id"`
	Team           int    `json:"team"`
	Position       Vec2   `json:"position"`
	Rotation       float64 `json:"rotation"`
	Velocity       Vec2   `json:"velocity"`
	AnimationState string `json:"animation_state"`
	HP             int    `json:"hp"`
}

type ErrorResponse struct {
	RequestPacketID uint16    `json:"request_packet_id"`
	Code            ErrorCode `json:"code"`
	Message         string    `json:"message"`
}
