#!/usr/bin/env bash
set -euo pipefail

project_root="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
go_bin="${GO:-go}"

usage() {
    cat <<'EOF'
Использование: ./run.sh <команда> [аргументы]

Команды:
  run [аргументы CLI]  Собрать и запустить ассемблер
  build                Собрать build/assembler
  test [флаги go test]  Запустить тесты
  check [флаги тестов]  Запустить тесты и go vet
  fmt                  Отформатировать Go-код
  clean                Удалить каталог build
  help                 Показать эту справку

Пример:
  ./run.sh run -input tests/fixtures/load_const.json -output program.bin -test

Флаги CLI можно передавать и без команды run. Относительные пути к данным
отсчитываются от текущего каталога. GO задаёт путь к исполняемому файлу Go.
EOF
}

require_go() {
    if ! command -v "$go_bin" >/dev/null 2>&1; then
        printf 'Ошибка: Go не найден. Установите Go 1.22+ или задайте GO.\n' >&2
        exit 1
    fi
}

without_arguments() {
    if [ "$#" -ne 0 ]; then
        printf 'Ошибка: команда не принимает аргументы. См. ./run.sh help.\n' >&2
        exit 2
    fi
}

project_go() (
    cd -- "$project_root"
    "$go_bin" "$@"
)

build() {
    mkdir -p -- "$project_root/build"
    project_go build -o "$project_root/build/assembler" ./src
}

action="${1:-help}"
if [ "$#" -gt 0 ]; then
    shift
fi

case "$action" in
    help|-h|--help)
        without_arguments "$@"
        usage
        exit 0
        ;;
    clean)
        without_arguments "$@"
        rm -rf -- "$project_root/build"
        exit 0
        ;;
    build|fmt)
        without_arguments "$@"
        ;;
    run|test|check)
        ;;
    -*)
        set -- "$action" "$@"
        action=run
        ;;
    *)
        printf 'Ошибка: неизвестная команда "%s". См. ./run.sh help.\n' "$action" >&2
        exit 2
        ;;
esac

require_go
case "$action" in
    run)
        build
        exec "$project_root/build/assembler" "$@"
        ;;
    build)
        build
        ;;
    test)
        project_go test "$@" ./...
        ;;
    check)
        project_go test "$@" ./...
        project_go vet ./...
        ;;
    fmt)
        project_go fmt ./...
        ;;
esac
