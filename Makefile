VERSION := $(shell cat VERSION)
NEXT_VERSION:=$(shell echo "$(VERSION)+0.1"|bc -l)

compile: generate release
generate:
	cd go; ./compile.sh
	@echo "0$(NEXT_VERSION)" > VERSION
release:
	gh release create v$(VERSION) ./build/* -t v$(VERSION) -R https://github.ark.org/brandon-kiefer/docker-dev 
apache:
	docker buildx build -f apache/Dockerfile apache/. --platform linux/arm64,linux/amd64 --push --tag brandonkiefer/php-dev:apache
bind:
	docker buildx build -f bind/Dockerfile bind/. --platform linux/arm64,linux/amd64 --push --tag brandonkiefer/php-dev:bind
mailhog:
	docker buildx build -f mailhog/Dockerfile mailhog/. --platform linux/arm64,linux/amd64 --push --tag brandonkiefer/php-dev:mailhog
php54:
	docker buildx build -f php/54/Dockerfile php/. --platform linux/arm64,linux/amd64 --push --tag brandonkiefer/php-dev:54
php56:
	docker buildx build -f php/56/Dockerfile php/. --platform linux/arm64,linux/amd64 --push --tag brandonkiefer/php-dev:56
php72:
	docker buildx build -f php/72/Dockerfile php/. --platform linux/arm64,linux/amd64 --push --tag brandonkiefer/php-dev:72
php74:
	docker buildx build -f php/74/Dockerfile php/. --platform linux/arm64,linux/amd64 --push --tag brandonkiefer/php-dev:74
