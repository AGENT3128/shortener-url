# shortener-url service

## dev

```
➜  shortener-url ✗ task
task: Available tasks for this project:
* build:                  Build the shortener binary
* clean:                  Clean build directory
* dependencies:           Install dependencies
* lint:                   Run linters
* run:                    Run the shortener binary
* test:                   Run tests
* test_increment_1:       Test increment 1
... <omited> ...
* test_increment_N:       Test increment N
* test_increment_all:       Run all tests
```


## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).
