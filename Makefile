# run templ generation in watch mode  
templ:
	templ generate --watch --proxy="http://localhost:8090" --open-browser=false

# run air for go hot reload
server:
	air \
	--build.cmd "go build -o ./tmp/bin/main ." \
	--build.bin "tmp/bin/main" \
	--build.delay "100" \
	--build.exclude_dir "node_modules" \
	--build.include_ext "go" \
	--build.stop_on_error "true" \
	--misc.clean_on_exit true

# watch tailwind css changes
tailwind:
	tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --watch

dev:
	make -j3 tailwind templ server
