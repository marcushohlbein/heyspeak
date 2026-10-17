module github.com/speek/speek

go 1.27.1

require (
	github.com/ebitengine/purego v0.11.1
	github.com/egoist/mygo v0.5.5
	github.com/gen2brain/malgo v0.11.26
	github.com/ggerganov/whisper.cpp/bindings/go v0.0.0
)

replace github.com/ggerganov/whisper.cpp/bindings/go => ./third_party/whisper.cpp/bindings/go

require (
	github.com/go-text/typesetting v0.3.5 // indirect
	golang.org/x/image v0.46.0 // indirect
	golang.org/x/text v0.43.0 // indirect
)
