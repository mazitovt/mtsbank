Задание №1

Сервис генерации потока цен сделок валютных пар на основании заданного тренда

Возможные паттерны генерации:
* `TIME` - использовать текущее время
* `SEED` - использовать значение `RATE_GENERATOR_SEED`

Уровни логирования: `debug`, `info`, `warn`, `error`

TODO:
- [x] Convert Cache interface to generic
- [x] Make SimpleCache thread-safe
- [x] Use sync.Pool for response encoding
- [ ] Health checks: liveness, readiness, startup
