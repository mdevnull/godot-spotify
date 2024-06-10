extends CanvasLayer

@onready var spotify_node: Node = get_node_or_null("../GodotSpotify")
@onready var auth_text: TextEdit = $CenterContainer/PanelContainer/VBoxContainer/AuthURLContainer/TextEdit
@onready var arist_name: Label = $CenterContainer/PanelContainer/VBoxContainer/ArtistContainer/AristName
@onready var song_name: Label = $CenterContainer/PanelContainer/VBoxContainer/SongContainer/SongName
@onready var album_name: Label = $CenterContainer/PanelContainer/VBoxContainer/AlbumContainer/AlbumName
@onready var is_playing_cb: CheckBox = $CenterContainer/PanelContainer/VBoxContainer/PlayingContainer/CheckBox

func _process(_delta):
  if spotify_node == null:
    return
  
  if auth_text.text != spotify_node.spotify_auth_url:
    auth_text.text = spotify_node.spotify_auth_url
    
  arist_name.text = spotify_node.artist_names
  song_name.text = spotify_node.track_name
  album_name.text = spotify_node.album_name
  is_playing_cb.button_pressed = spotify_node.is_playing
