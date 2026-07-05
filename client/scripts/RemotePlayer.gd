extends Node2D
class_name RemotePlayer

var player_id := ""

# Interpolation buffer
var state_buffer = []
# Fixed delay of 100ms for interpolation (2 server ticks at 20TPS)
var interpolation_delay := 100.0 

@onready var network = get_node("/root/Network")

func _ready() -> void:
	if network:
		network.snapshot_received.connect(_on_snapshot_received)

func _on_snapshot_received(snapshot: Dictionary) -> void:
	var players = snapshot.get("players", [])
	for p in players:
		if p.get("player_id", "") == player_id:
			# Push state into buffer with local timestamp
			state_buffer.append({
				"timestamp": Time.get_ticks_msec(),
				"position": Vector2(p["position"]["x"], p["position"]["y"]),
				"rotation": p["rotation"]
			})
			break
			
	# Keep buffer size manageable
	if state_buffer.size() > 20:
		state_buffer.pop_front()

func _process(_delta: float) -> void:
	if state_buffer.size() < 2:
		# Not enough states to interpolate
		if state_buffer.size() == 1:
			position = state_buffer[0].position
			rotation = state_buffer[0].rotation
		return
		
	var render_time = Time.get_ticks_msec() - interpolation_delay
	
	# Find the two states to interpolate between
	var state_a = null
	var state_b = null
	
	# Loop backwards
	for i in range(state_buffer.size() - 1, -1, -1):
		if state_buffer[i].timestamp <= render_time:
			state_a = state_buffer[i]
			if i + 1 < state_buffer.size():
				state_b = state_buffer[i + 1]
			break
			
	if state_a and state_b:
		# We can interpolate
		var time_diff = float(state_b.timestamp - state_a.timestamp)
		var t = 0.0
		if time_diff > 0:
			t = (render_time - state_a.timestamp) / time_diff
			t = clamp(t, 0.0, 1.0)
			
		position = state_a.position.lerp(state_b.position, t)
		rotation = lerp_angle(state_a.rotation, state_b.rotation, t)
		
	elif state_a:
		# We are ahead of the buffer (or exactly on the newest state)
		# Just set it to the newest state, or extrapolate if desired.
		position = state_a.position
		rotation = state_a.rotation
