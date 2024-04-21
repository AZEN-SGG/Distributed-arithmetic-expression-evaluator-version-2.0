# Распределенный вычислитель арифметических выражений версия 2.0

## Обзор

Распределенный вычислитель арифметических выражений версия 2.0 представляет собой улучшенную систему для асинхронных вычислений арифметических выражений, поддерживающую работу с множеством пользователей. Система обладает расширенной функциональностью для регистрации и авторизации пользователей, управления выражениями, их вычисления и сохранения результатов в базу данных.

## Начало работы

### Требования

- Go 1.22 или выше
- СУБД SQLite для хранения данных пользователей и выражений

### Установка и запуск

1. Клонировать репозиторий:
   ```bash
   git clone https://example.com/distributed-arithmetic-expression-evaluator-v2.0.git
   ```
2. Перейти в каталог проекта:
   ```bash
   cd distributed-arithmetic-expression-evaluator-v2.0
   ```
3. Запустить сервер:
   ```bash
   go run main.go
   ```

## Компоненты системы

### Сервер
Основной компонент системы, который обеспечивает следующие функции:
- **Регистрация и авторизация пользователей**
- **Добавление, обработка и хранение арифметических выражений**
- **Управление статусами и результатами выражений**
- **Выполнение арифметических операций с учетом заданного времени**

### Клиенты
Модуль, отвечающий за взаимодействие пользователей с системой. Поддерживает функции регистрации, авторизации и отправки запросов на вычисление выражений.

## Основные HTTP интерфейсы

### Регистрация пользователя
**POST** `/register`
- Принимает параметры `username` и `password`.
- Регистрирует нового пользователя в системе.

### Авторизация пользователя
**GET** `/login`
- Принимает параметры `username` и `password`.
- Возвращает токен для доступа к защищенным маршрутам.

### Добавление арифметического выражения
**POST** `/expression`
- Принимает параметры `expression`, `id` и `username`.
- Добавляет арифметическое выражение в базу и запускает его вычисление.

### Получение результата выражения
**POST** `/get`
- Принимает параметры `id` и `username`.
- Возвращает результат выражения, если оно было вычислено.

### Список всех выражений пользователя
**GET** `/list`
- Принимает параметр `username`.
- Возвращает список всех выражений пользователя с их статусами.

### Управление временем выполнения операций
**GET/POST** `/math`
- GET возвращает текущие времена выполнения операций.
- POST позволяет обновлять времена выполнения операций (параметры `addition`, `subtraction`, `multiplication`, `division`).

### Просмотр и управление вычислительными процессами
**GET** `/processes`
- Возвращает информацию о текущих вычислительных процессах.

## Примеры использования

### Регистрация нового пользователя
```bash
curl -X POST -d "username=user1&password=pass123" http://localhost:8080/register
```

### Авторизация пользователя
```bash
curl -X GET "http://localhost:8080/login?username=user1&password=pass123"
```

### Добавление выражения
```bash
curl -X POST -d "id=expr1&expression=2*2&username=user1&token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidXNlcjEiLCJwYXNzd29yZCI6InBhc3MxMjMifQ.bxi3MK96co-CcUTQdlA0jgzDa1JMJgLGOBodS9D2iH4" http://localhost:8080/expression
```

### Получение списка выражений
```bash
curl -X GET -d "username=user1" http://localhost:8080/list
```

### Получение результата выражения
```bash
curl -X POST -d "id=expr1&username=user1" http://localhost:8080/get
```

## Взаимодействие с сервером
Сервер предоставляет REST API для взаимодействия. Запросы можно отправлять через любой HTTP-клиент, например, `curl` или через программы для тестирования API, такие как Postman.

## Масштабируемость и надежность
Система поддерживает масштабирование путем добавления дополнительных вычислительных ресурсов. Все данные надежно сохраняются в базе данных, что позволяет восстанавливать работу системы без потери данных после сбоев.
## Защита и безопасность

Система обеспечивает безопасность данных и операций благодаря следующим механизмам:

- **Шифрование паролей**: Пароли пользователей хранятся в зашифрованном виде.
- **JWT аутентификация**: Доступ к операциям, требующим авторизации, контролируется через JWT токены, обеспечивая, что каждый запрос аутентифицирован.
- **Ограничение доступа**: Пользователи могут взаимодействовать только со своими выражениями, исключая возможность доступа к чужим данным.

## Мониторинг и управление

### Процессы
Система обеспечивает контроль за выполнением выражений и операциями:

- **Мониторинг состояния выражений**: Оркестратор отслеживает все активные и ожидающие выражения, позволяя пользователям получать информацию о статусе их выражений в реальном времени.
- **Управление вычислительными ресурсами**: Система предоставляет информацию о текущих вычислительных ресурсах, используемых для обработки выражений, что помогает оптимизировать загрузку ресурсов и время выполнения операций.

### Оптимизация вычислений
Оптимизация процесса вычислений происходит за счет распределенной обработки запросов и асинхронной архитектуры:

- **Балансировка нагрузки**: Вычислительные запросы равномерно распределяются между доступными ресурсами, что увеличивает эффективность обработки и снижает время ожидания ответа.
- **Асинхронная обработка**: Выражения обрабатываются асинхронно, позволяя системе обрабатывать множество запросов одновременно без блокировки основного потока выполнения.

## Резервное копирование и восстановление

Для обеспечения непрерывности бизнес-процессов и защиты данных пользователей, система поддерживает регулярное резервное копирование данных:

- **Резервное копирование базы данных**: Все данные регулярно архивируются, что позволяет восстановить систему в случае сбоев или потери данных.
- **Восстановление после сбоев**: В случае сбоя системы или потери данных, оркестратор может быстро восстановить операционное состояние, используя последние доступные резервные копии.

## Примеры дополнительных запросов

### Обновление времени выполнения операций
```bash
curl -X POST http://localhost:8080/math -d "addition=1000&subtraction=500&multiplication=1200&division=1500"
```
**Описание:** Этот запрос обновляет времена выполнения арифметических операций: сложения до 1000 мс, вычитания до 500 мс, умножения до 1200 мс и деления до 1500 мс.

### Просмотр вычислительных ресурсов
```bash
curl http://localhost:8080/processes
```
**Описание:** Запрос возвращает список текущих вычислительных ресурсов и задач, выполняемых на них, что помогает администраторам мониторить загрузку и распределение ресурсов в ре

альном времени.

## Заключение

Распределенный вычислитель арифметических выражений версия 2.0 представляет собой мощную, масштабируемую и безопасную систему для асинхронного вычисления арифметических выражений, обеспечивающую пользователям гибкий инструмент для работы с большими объемами данных в многопользовательской среде.