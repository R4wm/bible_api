# Makefile for bible_api project

.PHONY: all build run test clean opensearch-up opensearch-down index-kjv \
        deploy deploy-ui deploy-backend deploy-dry

all: build

build:
	go build -o bible_api ./cmd/bible_api.go

run: build
	./bible_api

test:
	go test ./...

clean:
	rm -f bible_api

opensearch-up:
	docker-compose up -d opensearch opensearch_dashboards

opensearch-down:
	docker-compose down

index-kjv:
	python3 scripts/index_kjv_to_opensearch.py --db data/kjv.db --index kjv_v2 --url http://localhost:9200

deploy:
	./scripts/deploy.sh

deploy-ui:
	./scripts/deploy.sh --ui-only

deploy-backend:
	./scripts/deploy.sh --backend-only

deploy-dry:
	./scripts/deploy.sh --no-restart
