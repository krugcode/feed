# !! change this to change your install dir !! #
CLI_NAME=feed
CLI_INSTALL_PATH=/usr/local/bin/$(CLI_NAME)

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


# cli
build-cli:
	@echo "Setting up CLI tool..."
	@echo ""
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found"; \
		exit 1; \
	fi
	@echo "Current CLI name: $(CLI_NAME)"
	@read -p "Enter CLI name (press enter for default): " cli_name; \
		cli_name=$${cli_name:-$(CLI_NAME)}; \
		echo "Using CLI name: $${cli_name}"; \
		echo ""; \
		default_install_path="/usr/local/bin/$${cli_name}"; \
		echo "Default install path: $${default_install_path}"; \
		read -p "Enter install path (press enter for default): " install_path; \
		install_path=$${install_path:-$${default_install_path}}; \
		echo "Using install path: $${install_path}"; \
		echo ""; \
		install_dir=$$(dirname "$${install_path}"); \
		if [ ! -d "$${install_dir}" ]; then \
			echo "Error: Install directory $${install_dir} does not exist"; \
			exit 1; \
		fi; \
		if [ ! -w "$${install_dir}" ] && [ "$$(id -u)" != "0" ]; then \
			echo "Warning: $${install_dir} is not writable. You may need sudo."; \
		fi; \
		echo "Reading config from .env..."; \
		app_url=$$(grep '^APP_URL=' .env | cut -d '=' -f2 | tr -d '"' | sed 's/^[[:space:]]*//;s/[[:space:]]*$$//'); \
		superuser_token=$$(grep '^SUPERUSER_TOKEN=' .env | cut -d '=' -f2 | tr -d '"' | sed 's/^[[:space:]]*//;s/[[:space:]]*$$//'); \
		if [ -z "$${app_url}" ]; then \
			echo "Error: APP_URL not found in .env"; \
			exit 1; \
		fi; \
		if [ -z "$${superuser_token}" ]; then \
			echo "Error: SUPERUSER_TOKEN not found in .env"; \
			exit 1; \
		fi; \
		echo "Building CLI with embedded config..."; \
		mkdir -p ./cli; \
		go build -o "./cli/$${cli_name}" \
			-ldflags "-s -w -X 'main.appURL=$${app_url}' -X 'main.token=$${superuser_token}'" \
			./cli/cli.go; \
		if [ $$? -eq 0 ]; then \
			echo "Build successful!"; \
			echo "Installing to $${install_path}..."; \
			if [ "$$(id -u)" = "0" ] || [ -w "$${install_dir}" ]; then \
				cp "./cli/$${cli_name}" "$${install_path}"; \
			else \
				sudo cp "./cli/$${cli_name}" "$${install_path}"; \
			fi; \
			if [ "$$(id -u)" = "0" ] || [ -w "$${install_dir}" ]; then \
				chmod +x "$${install_path}"; \
			else \
				sudo chmod +x "$${install_path}"; \
			fi; \
			rm "./cli/$${cli_name}"; \
			echo "CLI installed successfully!"; \
			echo "You can now use '$${cli_name}' from anywhere"; \
			echo ""; \
			echo "Try it out:"; \
			echo "  $${cli_name} post your-blog.md"; \
		else \
			echo "Build failed!"; \
			exit 1; \
		fi
