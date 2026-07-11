.PHONY: build debug kill attach clean run, tidy, screen

APP_NAME   := app-debug
CMD_PATH   := ./cmd/app
DLV_PORT   := 2345



# Build a debug binary with optimizations disabled (-N -l)
# so variable inspection and breakpoints work correctly.
build:
	go build -gcflags="all=-N -l" -o $(APP_NAME) $(CMD_PATH)

# Kill any dlv process currently holding DLV_PORT.
# Safe to run even if nothing is listening (xargs -r = no-op on empty input).
kill:
	@lsof -ti :$(DLV_PORT) | xargs -r kill -9
	@echo "Port $(DLV_PORT) is now free."

# Start Delve in headless mode against a fresh debug build.
# Always kills any leftover session first so this never fails
# with "address already in use".
debug: kill build
	dlv exec ./$(APP_NAME) \
		--headless \
		--listen=:$(DLV_PORT) \
		--api-version=2 \
		--accept-multiclient

# Same as `debug` but skips the rebuild step — use this if you
# haven't changed any .go files since the last build.
attach: kill
	dlv exec ./$(APP_NAME) \
		--headless \
		--listen=:$(DLV_PORT) \
		--api-version=2 \
		--accept-multiclient

# Run the app directly, no debugger attached (for quick manual testing
# of hotkeys etc. in a normal terminal).
run: build
	./$(APP_NAME)

# Remove the built binary.
clean:
	rm -f $(APP_NAME)

tidy: ## Clean up go.mod and go.sum
	@printf "${GREEN}🧹 Tidying modules...${NC}\n"
	go mod tidy
	go mod verify	

screen: ## Generate a new screen (Usage: make screen NAME=general)
	@if [ -z "$(NAME)" ]; then \
		printf "${RED}❌ Please provide a screen name: make screen NAME=general${NC}\n"; \
		exit 1; \
	fi
	@printf "${GREEN}📄 Generating screen: $(NAME)...${NC}\n"
	@go run ./cmd/retui-gen -make screen:$(NAME)	