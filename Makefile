WHISPER_DIR := third_party/whisper.cpp
WHISPER_BUILD := $(WHISPER_DIR)/build

.PHONY: whisper speek test package

APP_NAME := Speek
APP_DIR := dist/$(APP_NAME).app

whisper:
	@test -d $(WHISPER_DIR)/.git || git clone --depth 1 https://github.com/ggml-org/whisper.cpp $(WHISPER_DIR)
	cmake -S $(WHISPER_DIR) -B $(WHISPER_BUILD) -DCMAKE_BUILD_TYPE=Release -DBUILD_SHARED_LIBS=OFF
	cmake --build $(WHISPER_BUILD) --target whisper -j

speek: whisper
	CGO_ENABLED=1 go build -o bin/speek ./cmd/app

test:
	CGO_ENABLED=1 go test ./...

package: speek
	rm -rf $(APP_DIR)
	mkdir -p $(APP_DIR)/Contents/MacOS $(APP_DIR)/Contents/Resources
	cp bin/speek $(APP_DIR)/Contents/MacOS/speek
	cp scripts/macos/Info.plist $(APP_DIR)/Contents/Info.plist
	codesign --force --sign - --identifier dev.speek.app $(APP_DIR)
	@echo "Packaged and signed $(APP_DIR)"
