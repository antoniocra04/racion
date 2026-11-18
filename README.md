# Рацион

Личное веб-приложение для повторяющегося дневного рациона. Показывает КБЖУ за день, по приёмам пищи и для каждой порции. Поддерживает собственные продукты и несколько именованных версий рациона.

Карточка продукта хранит КБЖУ на 100 г. Сохранённые рационы ссылаются на карточки, поэтому правка карточки пересчитывает все связанные рационы. Граммовки хранятся отдельно для каждой версии.

## Запуск на Raspberry Pi 5

На 64-битной Raspberry Pi OS установите Docker Engine с плагином Compose. [Инструкция для ARM64](https://docs.docker.com/engine/install/debian/#install-using-the-apt-repository)

```sh
git clone https://github.com/antoniocra04/racion.git
cd racion
docker compose up -d --build
```

С телефона в домашней сети откройте `http://<локальный-IP-Pi>:8084`. При необходимости задайте порт через `RATION_PORT`, а адрес интерфейса Pi через `RATION_BIND_IP` в файле `.env` рядом с `compose.yaml`. Приложение рассчитано на доверенную домашнюю сеть. Оставьте проброс порта на роутере выключенным.

Docker Compose собирает образ под архитектуру Pi. Контейнер работает от отдельного пользователя. База `/data/ration.db` хранится в томе `ration-data` и сохраняется при пересоздании контейнера.

Обновление выполняется из папки проекта.

```sh
git pull --ff-only
docker compose up -d --build
```

Команды `docker compose stop` и `docker compose down` сохраняют том. Флаг `-v` у `down` удаляет том с базой.

## Резервная копия

Снимок можно выгрузить из работающего контейнера. Скопируйте полученный файл на другое устройство.

```sh
(
  set -e
  backup="ration-$(date +%Y%m%d-%H%M%S).db"
  partial=$(mktemp "./.$backup.XXXXXX")
  trap 'rm -f "$partial"' EXIT
  docker compose exec -T ration /usr/local/bin/ration -db /data/ration.db -backup "/tmp/$backup"
  docker compose exec -T ration cat "/tmp/$backup" > "$partial"
  ln "$partial" "./$backup"
  docker compose exec -T ration rm "/tmp/$backup"
  printf 'Резервная копия: %s\n' "$backup"
)
```

Для восстановления укажите путь к своей копии. Команда сохраняет прежнюю базу для отката.

```sh
backup=./ration-backup.db
test -s "$backup" &&
  docker compose stop ration &&
  docker compose run --rm --no-deps -T --entrypoint sh ration -ec '
    umask 077
    temp=$(mktemp /data/restore.XXXXXX)
    trap "rm -f \"$temp\"" EXIT
    cat > "$temp"
    test -s "$temp"
    if [ -f /data/ration.db ]; then
      previous=$(mktemp /data/ration-before-restore.XXXXXX)
      cp /data/ration.db "$previous"
      printf "Previous database saved at %s\n" "$previous"
    fi
    mv -f "$temp" /data/ration.db
  ' < "$backup" &&
  docker compose up -d
```

После запуска проверьте название активного рациона, его продукты и КБЖУ.

Если приложение уже было установлено как служба, остановите её и выгрузите существующую базу перед запуском Compose.

```sh
sudo systemctl stop ration
sudo /usr/local/bin/ration -db /var/lib/ration/ration.db -backup "$PWD/ration-migration.db"
sudo chown "$(id -u):$(id -g)" ration-migration.db
```

Запустите Compose и восстановите `ration-migration.db` по инструкции выше. После проверки отключите автозапуск старой службы командой `sudo systemctl disable ration`.

## Локальная разработка

Для локальной разработки нужен Go 1.26 или новее.

```sh
go run ./cmd/ration
```

Откройте [127.0.0.1:8080](http://127.0.0.1:8080). При первом запуске приложение создаст пустую базу `data/ration.db`. Для выбора файла базы или адреса используйте флаги `-db` и `-addr`.

```sh
go test ./...
```
