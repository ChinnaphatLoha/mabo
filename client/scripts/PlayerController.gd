extends Node2D
class_name PlayerController

var player_id := ""
var speed := 5.0

var current_sequence := 0
var pending_inputs := []

# Reference to the network singleton
@onready var network = get_node("/root/Network")

func _ready() -> void:
	if network:
		network.snapshot_received.connect(_on_snapshot_received)

func _physics_process(delta: float) -> void:
	if not is_multiplayer_authority():
		return
		
	# 1. Sample input
	var input_vector := Input.get_vector("ui_left", "ui_right", "ui_up", "ui_down")
	var rotation_val := rotation
	
	current_sequence += 1
	
	var input_cmd = {
		"sequence": current_sequence,
		"move_x": input_vector.x,
		"move_y": input_vector.y,
		"rotation": rotation_val,
		"delta": delta
	}
	
	pending_inputs.append(input_cmd)
	
	# 2. Client Prediction: apply immediately
	_apply_input(input_cmd)
	
	# 3. Send to server
	if network:
		network.send_movement(current_sequence, input_vector.x, input_vector.y, rotation_val)

func _apply_input(cmd: Dictionary) -> void:
	var move_dir = Vector2(cmd["move_x"], cmd["move_y"]).normalized()
	position += move_dir * speed * cmd["delta"]
	rotation = cmd["rotation"]

func _on_snapshot_received(snapshot: Dictionary) -> void:
	# Find my player data
	var my_state = null
	var players = snapshot.get("players", [])
	for p in players:
		if p.get("player_id", "") == player_id:
			my_state = p
			break
			
	if my_state == nil:
		return
		
	var server_pos = Vector2(my_state["position"]["x"], my_state["position"]["y"])
	var last_processed = my_state.get("last_processed_sequence", 0)
	
	# 4. State Reconciliation
	# Remove confirmed inputs
	var i = 0
	while i < pending_inputs.size():
		if pending_inputs[i]["sequence"] <= last_processed:
			pending_inputs.remove_at(i)
		else:
			i += 1
			
	# Snap to server position
	position = server_pos
	
	# Replay remaining unacknowledged inputs
	for pending in pending_inputs:
		_apply_input(pending)
