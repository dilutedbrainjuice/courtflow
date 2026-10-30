.PHONY: check-env check
check-env:
	bash scripts/check-env.sh
check:
	bash -n scripts/check-env.sh
	git diff --check

LOCAL_DB_URL = postgres://courtflow:courtflow_dev_only@127.0.0.1:5433/courtflow?sslmode=disable
.PHONY: dev-db migrate-local dev-api dev-web test build verify-local
dev-db:
	bash scripts/dev-db.sh start
migrate-local:
	bash scripts/migrate-local.sh
dev-api:
	cd backend && DATABASE_URL='$(LOCAL_DB_URL)' ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000 go run ./cmd/server
dev-web:
	cd frontend && npm run dev
test:
	cd backend && TEST_DATABASE_URL='$(LOCAL_DB_URL)' go test -race ./...
build:
	cd backend && go build ./...
	cd frontend && npm run build
verify-local:
	TEST_DATABASE_URL='$(LOCAL_DB_URL)' bash scripts/verify-local.sh
