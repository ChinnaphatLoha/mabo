extends CanvasLayer
class_name DebugUI

@onready var network = get_node("/root/Network")
@onready var ping_label = $VBoxContainer/PingLabel
@onready var rtt_label = $VBoxContainer/RTTLabel
@onready var tick_label = $VBoxContainer/TickLabel
@onready var fps_label = $VBoxContainer/FPSLabel

var last_ping_sent := 0
var ping := 0
var current_tick := 0

func _ready() -> void:
	if network:
		network.snapshot_received.connect(_on_snapshot_received)

func _process(_delta: float) -> void:
	fps_label.text = "FPS: " + str(Engine.get_frames_per_second())
	
	# Send ping every second
	if Time.get_ticks_msec() - last_ping_sent > 1000:
		last_ping_sent = Time.get_ticks_msec()
		if network and network.is_connected_to_server:
			network.send_packet(100, {"timestamp": last_ping_sent})

# A complete implementation would handle Pong packets (ID 101) to calculate true RTT.
# For demo purposes, we will just display the tick from the last snapshot.
func _on_snapshot_received(snapshot: Dictionary) -> void:
	current_tick = snapshot.get("tick", 0)
	tick_label.text = "Tick: " + str(current_tick)
	
	# RTT / Ping requires server pong handling which we omit for brevity
	# ping_label.text = "Ping: " + str(ping) + " ms"
