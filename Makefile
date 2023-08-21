VERSION := $(shell cat VERSION)
NEXT_VERSION:=$(shell echo "$(VERSION)+0.1"|bc -l)

compile: generate install release
install:
	rm -rf ./build/*
	cd go; ./compile.sh
	cp ./build/docdev-darwin-arm64 /usr/local/bin/docdev
generate:
	@echo "$(NEXT_VERSION)" > VERSION
release:
	gh release create v$(NEXT_VERSION) ./build/* -t v$(NEXT_VERSION) -R https://github.com/phpdocdev/docdev
apache:
	docker buildx build -f apache/Dockerfile apache/. --platform linux/arm64,linux/amd64 --push --tag brandonkiefer/php-dev:apache
bind:
	docker buildx build -f bind/Dockerfile bind/. --platform linux/arm64,linux/amd64 --push --tag brandonkiefer/php-dev:bind
mailhog:
	docker buildx build -f mailhog/Dockerfile mailhog/. --platform linux/arm64,linux/amd64 --push --tag brandonkiefer/php-dev:mailhog
php54:
	docker buildx build -f php/54/Dockerfile php/. --platform linux/arm64,linux/amd64 --push --tag brandonkiefer/php-dev:5.4
php56:
	docker buildx build -f php/56/Dockerfile php/. --platform linux/arm64,linux/amd64 --push --tag brandonkiefer/php-dev:5.6
php72:
	docker buildx build -f php/72/Dockerfile php/. --platform linux/arm64,linux/amd64 --push --tag brandonkiefer/php-dev:7.2
php74:
	docker buildx build -f php/74/Dockerfile php/. --platform linux/arm64,linux/amd64 --push --tag brandonkiefer/php-dev:7.4
php82:
	docker buildx build -f php/82/Dockerfile php/. --platform linux/arm64,linux/amd64 --push --tag brandonkiefer/php-dev:8.2
