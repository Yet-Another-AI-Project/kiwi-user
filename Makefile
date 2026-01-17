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
	docker rmi crpi-by4agx6tziel0uqm.cn-shanghai.personal.cr.aliyuncs.com/kiwi/kiwi-user:staging || true
	docker build -f build/Dockerfile --build-arg gitlab_user=${GITLAB_USER} --build-arg gitlab_access_token=${GITLAB_ACCESS_TOKEN} -t kiwi-user:staging .
	docker tag kiwi-user:staging crpi-by4agx6tziel0uqm.cn-shanghai.personal.cr.aliyuncs.com/kiwi/kiwi-user:staging
	docker push crpi-by4agx6tziel0uqm.cn-shanghai.personal.cr.aliyuncs.com/kiwi/kiwi-user:staging

build-production-image:
	docker rmi kiwi-user:$(VERSION) || true
	docker rmi crpi-by4agx6tziel0uqm.cn-shanghai.personal.cr.aliyuncs.com/kiwi/kiwi-user:$(VERSION) || true
	docker build -f build/Dockerfile --build-arg gitlab_user=${GITLAB_USER} --build-arg gitlab_access_token=${GITLAB_ACCESS_TOKEN} -t kiwi-user:$(VERSION) .
	docker tag kiwi-user:$(VERSION) crpi-by4agx6tziel0uqm.cn-shanghai.personal.cr.aliyuncs.com/kiwi/kiwi-user:$(VERSION)
	docker push crpi-by4agx6tziel0uqm.cn-shanghai.personal.cr.aliyuncs.com/kiwi/kiwi-user:$(VERSION)

generate-client-pkg:
	swagger generate client -f docs/swagger.json -A kiwi-user -t /root/Github/kiwi-lib/client/kiwiuser