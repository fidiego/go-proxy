# vim: set ft=make :

docker_compose_cmd := "docker compose -f docker/docker-compose.yml"

# print this message
default:
	just --list

# run local dev environment with docker-compose
[group('dev')]
up:
	{{ docker_compose_cmd }} up --build --remove-orphans

# docker-compose down
[group('dev')]
down:
	{{ docker_compose_cmd }} down

# run a command in the given container
[group('dev')]
run container="web" args="pwd && echo '\nhello from inside the container\n\n' && ls -Alh":
	{{ docker_compose_cmd }} run {{container}} {{args}}

# resart a specific container, or all if un-specified.
[group('dev')]
restart container="":
	{{ docker_compose_cmd }} restart {{container}}

# stop a specific container, or all if un-specified.
[group('dev')]
stop container="":
	{{ docker_compose_cmd }} stop {{container}}

# stop a specific container, or all if un-specified.
[group('build')]
build container="":
	{{ docker_compose_cmd }} build {{container}}

# install a package with `go get` inside of the web container
[group('go')]
get args:
	{{ docker_compose_cmd }} run web go get {{args}}

# run tidy
[group('go')]
tidy:
	{{ docker_compose_cmd }} run web go mod tidy
