# run templ generation in watch mode  
templ:
	templ generate --watch  --proxy="http://localhost:8090" --open-browser=false

# run air for go hot reload
# TODO: could switch this to a toml file but i have enough config files in my life
server:
	air \
	--build.cmd "go build -o ./bin ." \
	--build.bin "./bin" \
	--build.args_bin "serve" \
	--build.delay "100" \
	--build.exclude_dir "node_modules,pb_data" \
	--build.include_ext "go" \
	--build.stop_on_error "true" \
	--misc.clean_on_exit true

# watch tailwind css changes
tailwind:
	tailwindcss -i ./pb_public/assets/css/input.css -o ./pb_public/assets/css/output.css --watch

dev:
	make -j3 tailwind templ server
# pb utils
logs:
	@echo "### Pocketbase Logs ###"
	@tail -f pb_data/logs.db

admin:
	@echo "### Create Pocketbase SU ##"
	@go run main.go superuser create
