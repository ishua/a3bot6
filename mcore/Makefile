.PHONY: help run build clean

help: ## Показать доступные команды
	@echo "Available commands:"
	@echo "  make run     - Собрать и запустить mcore"
	@echo "  make build   - Только собрать"
	@echo "  make clean   - Удалить build файлы и БД"

build: ## Собрать приложение
	@echo "Building mcore..."
	@mkdir -p build
	cd cmd/app && CGO_ENABLED=1 go build -o ../../build/mcore
	@echo "✓ Build complete"

run: build ## Запустить mcore
	@echo "Starting mcore..."
	./build/mcore

clean: ## Очистить
	@echo "Cleaning..."
	rm -rf build
	rm -f data/*.db
	@echo "✓ Clean complete"