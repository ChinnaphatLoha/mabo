extends Node

signal connected_to_server
signal snapshot_received(snapshot: Dictionary)
signal player_spawned(event: Dictionary)
signal error_received(code: String, message: String)

const PACKET_CONNECT = 1
const PACKET_LOGIN_REQ = 10
const PACKET_LOGIN_RESP = 11
const PACKET_ERROR_RESP = 49
const PACKET_MOVEMENT = 150
const PACKET_SNAPSHOT = 200
const PACKET_PLAYER_SPAWNED = 250

var udp := PacketPeerUDP.new()
var is_connected_to_server := false
var server_ip := "127.0.0.1"
var server_port := 9000

var session_id := ""
var my_player_id := ""

func _ready() -> void:
	# For simplicity, connect immediately
	connect_to_server()

func connect_to_server():
	udp.connect_to_host(server_ip, server_port)
	# Send a login request
	var req = {"guest_name": "GodotClient"}
	send_packet(PACKET_LOGIN_REQ, req)

func _process(delta: float) -> void:
	if udp.get_available_packet_count() > 0:
		for i in range(udp.get_available_packet_count()):
			var packet_bytes = udp.get_packet()
			if packet_bytes.size() > 0:
				handle_packet(packet_bytes)

func send_packet(id: int, payload: Dictionary = {}):
	var data = JSON.stringify(payload).to_utf8_buffer()
	var packet_buffer = PackedByteArray()
	packet_buffer.append(id)
	packet_buffer.append_array(data)
	udp.put_packet(packet_buffer)

func send_movement(sequence: int, move_x: float, move_y: float, rotation: float):
	send_packet(PACKET_MOVEMENT, {
		"sequence": sequence,
		"move_x": move_x,
		"move_y": move_y,
		"rotation": rotation
	})

func handle_packet(bytes: PackedByteArray):
	var id = bytes[0]
	var payload_bytes = bytes.slice(1)
	var payload = {}
	if payload_bytes.size() > 0:
		var json_string = payload_bytes.get_string_from_utf8()
		var json = JSON.new()
		if json.parse(json_string) == OK:
			payload = json.data
	
	match id:
		PACKET_LOGIN_RESP:
			session_id = payload.get("session_id", "")
			my_player_id = payload.get("player_id", "")
			is_connected_to_server = true
			
			# Auto-create room for demo
			send_packet(20, {"capacity": 10}) 
			connected_to_server.emit()
		PACKET_ROOM_CREATED, 21:
			print("Room created")
		PACKET_ROOM_JOINED, 23:
			print("Room joined")
		PACKET_SNAPSHOT:
			snapshot_received.emit(payload)
		PACKET_PLAYER_SPAWNED:
			player_spawned.emit(payload)
		PACKET_ERROR_RESP:
			var code = payload.get("code", "UNKNOWN")
			var msg = payload.get("message", "")
			error_received.emit(code, msg)
			print("Server Error: ", code, " - ", msg)
