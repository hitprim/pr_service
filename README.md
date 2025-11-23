## Запуск

```shell
  docker-compose up --build
```

## Проблемы

Таблица pr_reviewers со связью many2many создавалась примерно так: pull_request_pull_request_id,
user_user_id. Долго не понимал почему. Оказалось все просто - нужно было явно указать наименование:
```go
Reviewers   []User  `gorm:"many2many:pr_reviewers;joinForeignKey:PullRequestID;JoinReferences:UserID"
```

---
Я не стал использовать явную структуру TeamMember так как заметил, что User включает те же поля + просто один лишний.
То есть User удовлетворяет TeamMember.