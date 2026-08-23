# /note task — интеграция taskparser в DoTask (Phase 4)

> Подключение готового taskparser (phases 1–3) к DoTask в notes: конфиг, чтение файла, фильтрация, форматирование, возврат ответа в mcore.

Без задачи (новая функциональность)

## Контекст

- Фазы 1–3 уже реализованы в `notes/internal/domain/taskparser/` — парсер, фильтры, форматирование. Все юнит-тесты проходят.
- Надо добавить поддержку команды `taskList` в DoTask сервиса notes: получить таску, прочитать файл задач, распарсить, отфильтровать по аргументу, вернуть отформатированный текст.
- Конфиг `notes_config.yaml` управляет путём к файлу задач, стоп-маркером и поведением по умолчанию.
- `ReportTaskReq.TextMsg` — одно поле string. `FormatTasks` возвращает `[]string` (части <= 4096 символов). Для простоты склеиваем части через `\n\n---\n\n`. Ограничение 4096 — забота слоя tbot/Telegram, не notes.
- Синоним `/nt` и парсинг аргументов в mcore — это фаза 5, здесь не делаем.

## Цель

Сервис notes отвечает на таску с `TaskNoteCmdTaskList`: читает markdown-файл задач, парсит через taskparser, выбирает фильтр по AddText, возвращает отформатированный список задач (или warnings).

## Фаза 4: Поддержка /note task в notes (DoTask)

- [ ] В `mcore/pkg/schema/task.go`:
  - Добавить константу: `TaskNoteCmdTaskList TaskNoteCmd = "taskList"` в блок const после `TaskNoteCmdAddBP`
- [ ] В `notes/cmd/app/main.go`:
  - Добавить поле `TaskFilePath string` в `MyConfig` (default `"My Long Tasks List.md"`, usage `"path to task list markdown file relative to repo"`)
  - Добавить поле `TaskStopMarker string` (default `"### Long"`, usage `"stop marker line in task file"`)
  - Добавить поле `TaskDefaultBehavior string` (default `"due"`, usage `"default filter: list|due|today"`)
  - В `model := domain.NewModel(gc)` добавить передачу трёх новых полей: `domain.NewModel(gc, cfg.TaskFilePath, cfg.TaskStopMarker, cfg.TaskDefaultBehavior)`
- [ ] В `notes/internal/domain/model.go`:
  - В `Model` добавить три поля: `taskFilePath string`, `taskStopMarker string`, `taskDefaultBehavior string`
  - Изменить `NewModel(gitClient *gitapi.GitClient, taskFilePath, taskStopMarker, taskDefaultBehavior string) *Model` — заполнить новые поля
  - Добавить `case schema.TaskNoteCmdTaskList: return m.taskList(task)` в switch `DoTask`
  - Реализовать метод `taskList(task schema.Task) schema.ReportTaskReq`:
    - Вызвать `m.readFile(m.taskFilePath)`
    - При ошибке чтения — вернуть `TaskStatusError`
    - Вызвать `taskparser.ParseTaskFile(lines, m.taskStopMarker)`
    - Определить фильтр по `task.TaskData.Tn.AddText`:
      - `""` → использовать `m.taskDefaultBehavior`
      - `"list"` → `taskparser.FilterAll(tf)`
      - `"due"` → `taskparser.FilterDue(tf)`
      - `"today"` → `taskparser.FilterToday(tf, time.Now())`
      - `"due today"` → `taskparser.FilterDueToday(tf, time.Now())`
      - `"warn"` → `taskparser.FormatWarnings(tf.Warnings, 4096)` (без фильтра задач)
    - Для `"warn"` — результат `FormatWarnings` склеить через `\n\n---\n\n` в один `TextMsg`
    - Для остальных — `taskparser.FormatTasks(items, 4096)`, склеить части через `\n\n---\n\n` в один `TextMsg`
    - Если результат пустой — `TextMsg = "no tasks found"`, `Status = Done`
    - Вернуть `TaskStatusDone` с `TextMsg`
- [ ] Обновить `notes/conf/notes_config.yaml`:
  - Добавить `task_file_path: "My Long Tasks List.md"`
  - Добавить `task_stop_marker: "### Long"`
  - Добавить `task_default_behavior: "due"`
- [ ] Собрать notes: `cd notes && go build -o build/notes ./cmd/app`
- [ ] Проверить, что код компилируется без ошибок

## Решения и договорённости

- **Без тестов**: задачи phases 1–3 протестированы в taskparser. Интеграция thin, код тривиальный — пропускаем unit-тесты.
- **Один TextMsg**: `ReportTaskReq.TextMsg` — string, не массив. `FormatTasks`/`FormatWarnings` возвращают `[]string` с частями <= 4096. Склеиваем через `\n\n---\n\n` в один `TextMsg`. Ограничение Telegram обрабатывается слоем tbot, не здесь.
- **Поведение по умолчанию**: если `AddText` пуст — берём из конфига (`m.taskDefaultBehavior`). Это позволяет менять дефолт без пересборки mcore.
- **Структура Model**: три новых поля добавлены напрямую в struct, без отдельного config-объекта. Это минимизирует diff.
- **Переиспользование readFile**: `m.readFile(m.taskFilePath)` — существующий метод, переиспользуем без дублирования.
- **Фаза 5 (mcore)**: парсинг `/note task` и синоним `/nt` — отдельная фаза, здесь не делаем.