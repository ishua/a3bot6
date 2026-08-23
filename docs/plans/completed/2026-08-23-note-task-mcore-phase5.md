# /note task — поддержка в mcore (Phase 5)

> Поддержка команды `/note task` на стороне mcore: парсинг подкоманды, проброс аргументов в `AddText` таски, добавление алиаса `nt`/`Nt`. Обработка самой таски (`TaskNoteCmdTaskList`) уже реализована в notes (Phase 4).

Без задачи (новая функциональность)

## Контекст

- Phases 1–4 уже реализованы:
  - `notes/internal/domain/taskparser/` — парсер, фильтры, форматирование (phases 1–3).
  - `notes` отвечает на таску `TaskNoteCmdTaskList`: читает файл, парсит, выбирает фильтр по `AddText`, возвращает текст (Phase 4).
  - Константа `TaskNoteCmdTaskList = "taskList"` уже добавлена в `mcore/pkg/schema/task.go`.
- Контракт `AddText` в notes (`notes/internal/domain/model.go`, switch `taskList`):
  - `""` → поведение по умолчанию из конфига (`task_default_behavior`, default `"today"`)
  - `"list"` → `FilterAll`
  - `"due"` → `FilterDue`
  - `"today"` → `FilterToday`
  - `"due today"` → `FilterDueToday`
  - `"warn"` → `FormatWarnings` (список warnings без задач)
  - любое другое значение → `TaskStatusError` с текстом `unknown task filter: <v>`
- Осталось: чтобы пользователь мог вызвать это из Telegram, mcore должен распознать `/note task ...` (и бесслешевый алиас `nt`/`Nt`, как у остальных синонимов notes), распарсить аргументы в `AddText` и создать таску `taskList`.

## Цель

Из Telegram команды `/note task [list|due|today|due today|warn]` и алиас `nt`/`Nt` создают таску `TaskNoteCmdTaskList` с корректным `AddText`, которую notes обрабатывает и возвращает отформатированный список задач.

## Фаза 5: Поддержка /note task в mcore

- [x] В `mcore/internal/taskmng/notetask.go` — вынести логику разбора в чистую функцию `parseNoteTaskArgs(words []string) (schema.TaskNote, error)`, содержащую текущий `switch words[1]` (кейсы `5bx`, `entry`, `inbox`, `pull`, `weight`, `bp`), и добавить в этот switch кейс:
  - `case "task", "tasks":` — собрать `addText = strings.Join(words[2:], " ")` (может быть пустым → default в notes), вернуть `schema.TaskNote{Command: schema.TaskNoteCmdTaskList, AddText: addText}`
  - `createNoteTask` вызывает этот хелпер и после него добавляет таску через `m.repo.AddTask`
- [x] В `mcore/internal/taskmng/notetask.go` — `createNoteTask`: поддержать голый `nt`/`Nt` (см. «Бэар `nt`» в решениях): при `len(words) < 2` и `words[0] == "nt" || words[0] == "Nt"` заменить слова на `["/note", "task"]`; для всех остальных случаев при `len(words) < 2` по-прежнему возвращать `for note need command`. Голые `n5`/`nd`/`nw`/`nbp`/`ni` не затрагиваются.
- [x] В `mcore/internal/taskmng/notetask.go` — `parseSimplifications`, добавить кейс:
  - `case "nt", "Nt": return append([]string{"/note", "task"}, words[1:]...), nil`
- [x] В `mcore/internal/taskmng/begindialog.go`:
  - В `case "/note", ...` строки диспетчера `createReply` добавить бесслешевые синонимы `"nt", "Nt"`
  - В `tnHelpText` добавить строку о `/note task` с алиасом `nt`/`Nt` и перечислением аргументов
- [x] Написать unit-тесты в `mcore/internal/taskmng/notetask_test.go`:
  - `TestParseSimplifications_nt` — `parseSimplifications([]string{"nt", "list"})` → `["/note","task","list"]`, и `parseSimplifications([]string{"Nt", "due", "today"})` → `["/note","task","due","today"]`
  - `TestParseNoteTaskArgs` — покрыть маппинг чистого хелпера: `["/note","task","list"]` → `{Command: taskList, AddText: "list"}`, `["/note","tasks","due","today"]` → `AddText: "due today"`, `["/note","task"]` → `AddText: ""`
  - `TestCreateNoteTask_BareNt` — голый `nt`/`Nt` → `taskList` c `AddText: ""` (без регрессии прочих алиасов); для этого потребуется заглушка `repo` в тесте или покрытие через извлечённый хелпер бэар-логики
- [x] Собрать mcore: `cd mcore && go build -o build/mcore ./cmd/app`
- [x] Прогнать тесты mcore: `cd mcore && go test ./internal/taskmng/...`
- [x] Проверить, что код компилируется без ошибок

## Решения и договорённости

- **Два подпути создания таски**: `/note task ...` напрямую и алиас `nt ...` через `parseSimplifications`. Оба сходятся в один кейс `"task"/"tasks"` в `parseNoteTaskArgs`.
- **Бесслешевый алиас `nt`/`Nt`**: все существующие синонимы notes (`nd`, `n5`, `nbp`, `nir`, ...) вводятся без ведущего слеша — и в `createReply`, и в `parseSimplifications`. `nt`/`Nt` следуют той же конвенции: пользователь пишет `nt list`, `nt due today`, а не `/nt`. Это критично для корректности: слеш-форма `/nt` не совпала бы ни с одним кейсом (везде `command not found`).
- **AddText = все слова после `task`**: `strings.Join(words[2:], " ")`. Это даёт `"due today"` (два слова) и позволяет пустой `""` (→ default в notes). mcore не валидирует список аргументов сам — неизвестный фильтр обрабатывает notes (`unknown task filter: ...`). Это держит mcore тонким и избегает дублирования контракта.
- **Поддержка `"today"`**: в оригинальном плане Phase 5 перечислены только `list|due|due today|warn`, но Phase 4 добавила фильтр `today` для полноты. Включаем `"today"` как документированный аргумент — консистентно с контрактом notes.
- **Бэар `nt` (без аргументов)**: желаемое поведение — `AddText=""` → default. Текущий код `createNoteTask` проверяет `len(words) < 2` до разбора, поэтому голый `nt` дал бы `for note need command`. Решение: разрешить голый `nt`/`Nt` → `["/note","task"]` → default. Изменение **ограничено только** `nt`/`Nt`: голые `n5`/`nd`/`nw`/`nbp`/`ni` по-прежнему дают `for note need command`, чтобы не регрессировать их UX (в notes пустой `AddText` для них → `add text is empty`).
- **Чистый хелпер `parseNoteTaskArgs`**: выносим существующий `switch words[1]` из `createNoteTask` в чистую функцию без репозитория. Это делает маппинг аргументов тестируемым без мок-фреймворка (требование ревью). `createNoteTask` вызывает хелпер, затем `m.repo.AddTask`.
- **Тесты**: в `mcore/internal/taskmng` тестов нет. Добавляем фокусные тесты на новую логику: `parseSimplifications` для `nt`, чистый хелпер `parseNoteTaskArgs`, и бэар-обработку `nt`. Остальное — сборка и ручная проверка.
- **Help**: обновляем `tnHelpText`, т.к. `/note task` — новая публичная команда; пользователь должен про неё знать из `/note help`.

## Граничные случаи для проверки

- `nt` / `Nt` (без аргументов) → default `today`; `/note task` → default `today`
- `/note task list` / `/note tasks list` → все задачи
- `/note task due today` / `nt due today` → due+today
- `/note task warn` / `nt warn` → warnings
- `/note task foo` → `unknown task filter: foo` (error от notes)
- Голые `n5`/`nd`/`nw`/`nbp` не сломать: по-прежнему `for note need command`
- Голой `/note` без подкоманды → `for note need command` (не сломать существующее)
- Регистр: `/note Task` не распознаётся (case-sensitive, соответствует конвенции прочих подкоманд)
