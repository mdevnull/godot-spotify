extends CanvasLayer

@onready var spotify_node: Node = get_node_or_null("../GodotSpotify")
@onready var auth_text: TextEdit = $CenterContainer/PanelContainer/VBoxContainer/AuthURLContainer/TextEdit
@onready var arist_name: Label = $CenterContainer/PanelContainer/VBoxContainer/ArtistContainer/AristName
@onready var song_name: Label = $CenterContainer/PanelContainer/VBoxContainer/SongContainer/SongName
@onready var album_name: Label = $CenterContainer/PanelContainer/VBoxContainer/AlbumContainer/AlbumName
@onready var is_playing_cb: CheckBox = $CenterContainer/PanelContainer/VBoxContainer/PlayingContainer/CheckBox
@onready var progress_bar: ProgressBar = $CenterContainer/PanelContainer/VBoxContainer/ProgressContainer/ProgressBar

var cover_url: String = ""
@onready var cover_texture_rect: TextureRect = $CenterContainer/PanelContainer/VBoxContainer/CoverContainer/TextureRect

func _process(_delta):
  if spotify_node == null:
    return
  
  if auth_text.text != spotify_node.spotify_auth_url:
    auth_text.text = spotify_node.spotify_auth_url
    
  arist_name.text = spotify_node.artist_names
  song_name.text = spotify_node.track_name
  album_name.text = spotify_node.album_name
  is_playing_cb.button_pressed = spotify_node.is_playing
  progress_bar.max_value = spotify_node.length_ms
  progress_bar.value = spotify_node.progress_ms

  if cover_url != spotify_node.cover_url:
    cover_url = spotify_node.cover_url
    change_img_from_url(spotify_node.cover_url)

func change_img_from_url(img_url: String):
  # Create an HTTP request node and connect its completion signal.
  var http_request = HTTPRequest.new()
  add_child(http_request)
  http_request.request_completed.connect(self._http_request_completed)

  # Perform the HTTP request. The URL below returns a PNG image as of writing.
  var error = http_request.request(img_url)
  if error != OK:
    push_error("An error occurred in the HTTP request.")

# Called when the HTTP request is completed.
func _http_request_completed(result, _response_code, _headers, body):
  if result != HTTPRequest.RESULT_SUCCESS:
    push_error("Image couldn't be downloaded. Try a different image.")

  var image = Image.new()
  var error = image.load_jpg_from_buffer(body)
  if error != OK:
    push_error("Couldn't load the image.")

  var texture = ImageTexture.create_from_image(image)

  cover_texture_rect.texture = texture