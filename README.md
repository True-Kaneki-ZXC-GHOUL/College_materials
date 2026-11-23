# Портал методических материалов

Бэкенд колледжа: учебные группы, пользователи и методические материалы.
Стек: Go, PostgreSQL, MinIO. Приложение пока запускается обычной командой `go run`, без своего Docker-образа.

## Что уже работает

- Вход, регистрация студента по инвайт-коду, обновление токена, выход.
- Сессии: список, отзыв, таймаут неактивности.
- Админ создаёт преподавателя, правит пользователей, блокирует и меняет роль.
- Учебные группы, предметы и инвайт-коды.
- Материалы трёх видов: **PDF**, **внешняя ссылка**, **текст** (HTML из редактора).
- Черновик, публикация, скрытие, мягкое удаление и восстановление.
- Новая версия создаётся, когда правят уже опубликованный материал.
- Студент видит только опубликованное: публичное или материал своей группы.
- Преподаватель читает все материалы, а менять и удалять может только свои. Админ может всё.
- PDF проверяется по содержимому и кладётся в MinIO. Скачивание — по временной ссылке.
- Заявки на опечатки: студент создаёт, преподаватель видит только заявки по своим материалам, меняет статус. Заявка в статусе «новая» дольше 5 дней помечается просроченной.

## Что нужно установить

- Go 1.26 (как в `go.mod`)
- PostgreSQL 14 или новее
- MinIO, если нужна загрузка PDF

Базу и MinIO можно поднять так:

```powershell
docker compose up -d
```

## Запуск API

Из корня проекта:

```powershell
go run ./cmd/api
```

Сервер слушает `http://localhost:8080`. Проверка: `http://localhost:8080/healthz` отвечает `ok`.

При первом запуске в режиме `dev` создаются пользователи:

| Роль | Email | Пароль |
| --- | --- | --- |
| Админ | admin@example.com | admin12345 |
| Преподаватель | teacher@example.com | teacher12345 |
| Студент ИС-21 | student@example.com | student12345 |

И три примера: публичный конспект, ссылка только для ИС-21 и черновик, который студент не видит.

## Переменные окружения

Все значения ниже уже стоят по умолчанию, их можно не задавать на локальной машине.

| Переменная | По умолчанию | Зачем |
| --- | --- | --- |
| APP_ENV | dev | `dev` заполняет базу примерами |
| HTTP_ADDR | :8080 | адрес сервера |
| LOG_LEVEL | info | debug, info, warn, error |
| LOG_FORMAT | json | json или text |
| DB_HOST | localhost | |
| DB_PORT | 5432 | |
| DB_USER | postgres | |
| DB_PASSWORD | postgres | |
| DB_NAME | portal | |
| DB_SSLMODE | disable | |
| DB_MAX_CONNS | 10 | размер пула |
| JWT_SECRET | change-me-in-production | в `prod` секрет обязателен |
| JWT_ACCESS_TTL | 15m | короткий токен доступа |
| JWT_REFRESH_TTL | 168h | срок сессии, 7 дней |
| SESSION_IDLE_TIMEOUT | 30m | если сессией не пользуются дольше, refresh перестаёт работать |
| S3_ENDPOINT | localhost:9000 | MinIO |
| S3_ACCESS_KEY | minioadmin | |
| S3_SECRET_KEY | minioadmin | |
| S3_BUCKET | materials | бакет создаётся сам в dev |
| S3_USE_SSL | false | |
| S3_PRESIGN_TTL | 15m | сколько живёт ссылка на PDF |
| UPLOAD_MAX_SIZE | 52428800 | максимум файла, 50 МБ |

Миграции применяются сами при старте. Отдельная команда не нужна.

## Как ходить в API

Заголовок для закрытых ручек: `Authorization: Bearer <access_token>`.

Вход:

```powershell
curl.exe -s http://localhost:8080/api/v1/auth/login -H "Content-Type: application/json" -d "{\"email\":\"teacher@example.com\",\"password\":\"teacher12345\"}"
```

Дальше в примерах вместо `TOKEN` подставляется `access_token`.

### Материалы

Текст, сразу для всех:

```powershell
curl.exe -s http://localhost:8080/api/v1/materials -H "Authorization: Bearer TOKEN" -H "Content-Type: application/json" -d "{\"title\":\"Конспект\",\"type\":\"text\",\"status\":\"published\",\"is_public\":true,\"subject_id\":2,\"tags\":[\"go\"],\"content\":\"<p>Текст лекции</p>\"}"
```

Ссылка только для групп:

```powershell
curl.exe -s http://localhost:8080/api/v1/materials -H "Authorization: Bearer TOKEN" -H "Content-Type: application/json" -d "{\"title\":\"Видео\",\"type\":\"link\",\"status\":\"published\",\"is_public\":false,\"external_url\":\"https://go.dev\",\"group_ids\":[1]}"
```

PDF в три шага: черновик, файл, публикация. В ответе на загрузку есть `id`: если материал уже был опубликован, новый файл становится новой версией с новым номером.

```powershell
curl.exe -s http://localhost:8080/api/v1/materials -H "Authorization: Bearer TOKEN" -H "Content-Type: application/json" -d "{\"title\":\"Методичка\",\"type\":\"pdf\",\"status\":\"draft\",\"is_public\":true,\"subject_id\":2}"
curl.exe -s http://localhost:8080/api/v1/materials/1/file -H "Authorization: Bearer TOKEN" -F "file=@C:\path\file.pdf"
curl.exe -s -X PATCH http://localhost:8080/api/v1/materials/1/status -H "Authorization: Bearer TOKEN" -H "Content-Type: application/json" -d "{\"status\":\"published\"}"
```

Каталог: `GET /api/v1/materials`. Фильтры: `q`, `subject_id`, `group_id`, `type`, `status`, `tags`, `author_id`, `mine=true`, `page`, `page_size`, `sort_by` (`created_at`, `updated_at`, `title`, `views`). Корзина автора или админа: `deleted=true`.

Остальные ручки материалов: `GET /{id}`, `PUT /{id}`, `PATCH /{id}/status`, `DELETE /{id}`, `POST /{id}/restore`, `GET /{id}/versions`, `GET /{id}/download`.

DOCX не принимается. Если нужен документ Word, его сохраняют как PDF. Короткий текст пишут прямо на сайте, ссылку кладут в `external_url`.

### Группы, предметы, пользователи

Читать группы и предметы может любой вошедший пользователь. Создавать и удалять их может только админ.

Инвайт живёт 3 дня и срабатывает один раз. Студент регистрируется с ним и сразу попадает в группу:

`POST /api/v1/auth/register` с полями `email`, `password`, `full_name`, `invite_code`.

Админ:

- `POST /api/v1/invites` с `group_id`
- `POST /api/v1/users/teachers` с `email`, `password`, `full_name`
- `GET|PUT|DELETE /api/v1/users/{id}`
- `PATCH /api/v1/users/{id}/block` с `{"blocked": true}`
- `PATCH /api/v1/users/{id}/role` с `{"role": "teacher"}`

Удаление пользователя — это блокировка. Материалы автора из-за этого не пропадают.

Свои сессии: `GET /api/v1/me/sessions` и `DELETE /api/v1/me/sessions/{id}`.

### Заявки на опечатки

Студент создаёт заявку только на материал, который ему виден:

```powershell
curl.exe -s http://localhost:8080/api/v1/requests -H "Authorization: Bearer TOKEN" -H "Content-Type: application/json" -d "{\"material_id\":1,\"comment\":\"Опечатка во втором абзаце\"}"
```

`GET /api/v1/requests` у студента возвращает его заявки. У преподавателя — заявки только на его материалы. У админа — все. Фильтры: `status` (`new`, `in_progress`, `done`, `rejected`), `material_id`, `overdue=true`, `page`, `page_size`.

Преподаватель или админ меняет статус:

```powershell
curl.exe -s -X PATCH http://localhost:8080/api/v1/requests/1/status -H "Authorization: Bearer TOKEN" -H "Content-Type: application/json" -d "{\"status\":\"in_progress\"}"
```

Просрочка считается сама при открытии списка: если заявка больше 5 дней в статусе `new`, у неё появляется `overdue: true`. Тот же проход можно вызвать явно: `POST /api/v1/requests/mark-overdue`.

## Как устроен доступ

Студент в каталоге получает только опубликованные материалы, у которых `is_public = true`, плюс материалы, привязанные к его группе. Это фильтруется в SQL, а не только на экране.

Преподаватель видит чужие черновики и закрытые материалы, но `PUT` и `DELETE` чужого материала отвечают 403. Админ может править всё.

Публичный материал нельзя одновременно ограничить списком групп: он и так виден всем.
