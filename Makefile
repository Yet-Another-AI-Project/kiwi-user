config=`pwd`/conf/config.json

test-unit:
	go test -v -count 1 -run Unit ./...

generate-repository-code:
	go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/execquery --feature sql/lock  ./internal/infrastructure/repository/ent/schema

generate-migration-repository-db:
	atlas migrate diff --dir "file://internal/infrastructure/repository/ent/migrate/migrations" --to "ent://internal/infrastructure/repository/ent/schema" --dev-url "postgres://root:pgpass@localhost:5432/atlas_dev?search_path=public&sslmode=disable"

apply-migration-repository-db:
	atlas migrate apply --dir "file://internal/infrastructure/repository/ent/migrate/migrations" --url $(shell cat $(config) | jq .postgres.connection_string)

apply-migration-repository-db-full:
	atlas migrate apply --dir "file://internal/infrastructure/repository/ent/migrate/migrations" --url $(shell cat $(config) | jq .postgres.connection_string)

generate-api-swagger:
	swag init --parseDependency --parseInternal -g internal/facade/controller/api/controller.go

build-staging-image:
	docker rmi kiwi-user:staging || true
	docker rmi kiwihub.azurecr.io/kiwi-user:staging || true
	docker buildx build --platform linux/amd64 -f build/Dockerfile --build-arg github_user=${GITHUB_USER} --build-arg github_access_token=${GITHUB_ACCESS_TOKEN} -t kiwi-user:staging .
	docker tag kiwi-user:staging kiwihub.azurecr.io/kiwi-user:staging
	docker push kiwihub.azurecr.io/kiwi-user:staging

build-production-image:
	docker rmi kiwi-user:$(VERSION) || true
	docker rmi kiwihub.azurecr.io/kiwi-user:$(VERSION) || true
	docker buildx build --platform linux/amd64 -f build/Dockerfile --build-arg github_user=${GITHUB_USER} --build-arg github_access_token=${GITHUB_ACCESS_TOKEN} -t kiwi-user:$(VERSION) .
	docker tag kiwi-user:$(VERSION) kiwihub.azurecr.io/kiwi-user:$(VERSION)
	docker push kiwihub.azurecr.io/kiwi-user:$(VERSION)

generate-client-pkg:
	swagger generate client -f docs/swagger.json -A kiwi-user -t /root/Github/kiwi-lib/client/kiwiuser