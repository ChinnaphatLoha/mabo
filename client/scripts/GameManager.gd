extends Node2D

@onready var network = get_node("/root/Network")

func _ready() -> void:
	# Spawn a visual grid so we can see the camera moving!
	for i in range(20):
		for j in range(20):
			var dot = ColorRect.new()
			dot.color = Color.DIM_GRAY
			dot.size = Vector2(4, 4)
			dot.position = Vector2(i * 100 - 1000, j * 100 - 1000)
			add_child(dot)
			
	if network:
		network.player_spawned.connect(_on_player_spawned)

func _on_player_spawned(event: Dictionary) -> void:
	var pid = event.get("player_id", "")
	var is_me = (pid == network.my_player_id)
	
	print("Spawning player: ", pid, " is_me: ", is_me)
	
	# Create a visual representation
	var visual = ColorRect.new()
	visual.size = Vector2(40, 40)
	visual.position = Vector2(-20, -20) # Center the rect
	
	if is_me:
		visual.color = Color.GREEN
		var controller = PlayerController.new()
		controller.player_id = pid
		controller.name = "LocalPlayer"
		controller.add_child(visual)
		
		# Give local player a camera
		var cam = Camera2D.new()
		controller.add_child(cam)
		
		# Add to scene tree before making camera current
		add_child(controller)
		cam.make_current()
	else:
		visual.color = Color.RED
		var remote = RemotePlayer.new()
		remote.player_id = pid
		remote.name = "RemotePlayer_" + pid
		remote.add_child(visual)
		add_child(remote)
