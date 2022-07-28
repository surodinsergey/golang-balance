### Запуск
Склонируйте репозиторий:
```bash
git clone github.com/surodinsergey/golang-balance
```

Скопируйте `env.example` файл в `.env` файл.
```bash
$ cp .env.example .env
```

Соберите образ:
```bash
$ docker-compose up --build
```
Файл с тестовыми данными уже включен в сборку и находится в директории `db/init.sql`

### Разработка
В докер образ уже вшита программа `CompileDaemon` , которая будет автоматически пересобирать и компилировать приложение после изменения файлов на машине , пересобирать образ нет необходимости

### Примеры запросов API
Запрос получения данных баланса по id пользователя
`http://localhost:8010/balance/{id}` Methods("GET")

Body отсутствует

Запрос начисления/списания средств на балансе пользователя
`http://localhost:8010/balance/{id}` Methods("PUT")

Body `application/json`:
```json
{
  "data": {
    "sum": -1330
  }
}
```

Запрос перевода средств от пользователя к пользователю
`http://localhost:8010/balance/transfer` Methods("POST")

Body `application/json` :
```json
{
  "data": {
    "from" : 1,//от кого
    "to"   : 2,//кому
    "sum"  : 1000
  }
}
```