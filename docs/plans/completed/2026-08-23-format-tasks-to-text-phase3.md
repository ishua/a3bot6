# Format tasks to text (phase 3)

> Реализация функции форматирования задач в текст для подкоманды `/note task` в сервисе notes.

Фаза 3 из 6 (парсинг → фильтрация → **форматирование** → интеграция с notes → интеграция с mcore → e2e).

## Контекст

- Пакет `notes/internal/domain/taskparser` уже содержит:
  - `parser.go` (Phase 1) — парсинг markdown-файла в `TaskFile{Sections, Warnings}`
  - `filter.go` (Phase 2) — фильтрация задач: `FilterAll`, `FilterDue`, `FilterDueToday`, `FilterToday`
- Структуры данных:
  - `TaskItem` — `Text`, `Due time.Time` (zero если не задана), `Priority string` ("A"/"B"/""), `Emojis []string`, `Wikilinks []string`, `Subtasks []TaskItem`, `InlineLinks []InlineLink`
  - `InlineLink` — `Text string`, `URL string`
  - `Warning` — `Line int`, `Text string`, `Message string`
- **Ключевое**: парсер **извлекает** wikilinks, inline links и emojis в отдельные поля и **удаляет** их из `Text`. Функции форматирования должны **восстанавливать** их при выводе.
- Тестовые хелперы: `mkItem(text string, subtasks ...TaskItem)`, `withDue(year, month, day)` метод на `TaskItem`, `loadTestTaskFile(t)` из `filter_test.go`
- Тестовый файл: `notes/internal/domain/taskparser/testdata/tasklist.md`

## Цель

Пользователь получает отформатированный список задач через `/note task` в Telegram:
- Каждая задача: приоритет, текст, wikilinks, inline links, due-дата, эмодзи
- Подзадачи с отступом (tab `\t`)
- Разбивка по лимиту (Telegram — 4096 символов)
- Warnings для `/nt warn`

## Фаза 3: Форматирование задач в текст

### Задача 1: Реализовать `format.go`

**Создать:** `notes/internal/domain/taskparser/format.go`

- [x] Использовать существующую константу `dateLayout` из `parser.go` (не переопределять!)
- [x] Реализовать `formatTask(item TaskItem) string`:
  - Порядок: приоритет → текст → wikilinks → inline links → due → эмодзи
  - Приоритет: `(A) ` или `(B) ` если есть
  - Текст: `item.Text` (уже очищен от wikilinks, inline links, эмодзи)
  - Wikilinks: `[[link1]] [[link2]]` (через пробел, в порядке извлечения)
  - Inline links: `[text](url)` (в порядке извлечения)
  - Due: `due:YYYY-MM-DD` если `!Due.IsZero()`
  - Эмодзи: конкатенация `Emojis` в порядке извлечения
  - Подзадачи: каждая строка из рекурсивного вывода `formatTask(subtask)` получает префикс `\t` (т.е. отступ применяется ко **всем** строкам подзадачи, включая её собственные подзадачи — `\t\t` для уровня 2 и т.д.)
- [x] Реализовать `FormatTasks(items []TaskItem, limit int) []string`:
  - `limit <= 0` → нет лимита, весь вывод одной частью
  - Накапливать строки в текущую часть
  - Перед добавлением новой задачи проверять: `currentPart + "\n" + newLine` превышает `limit`?
    - Если да: закрыть текущую часть, начать новую
  - Если одна задача (со всеми подзадачами) превышает `limit` — всё равно добавить как есть (Telegram обрежет)
  - Возвращать `[]string{}` (не nil) для пустого входа
- [x] Реализовать `FormatWarnings(warnings []Warning, limit int) []string`:
  - Формат: `line N: <message> | <text>`
  - Разбивка по `limit` по границам warning (поведение как в `FormatTasks`: `limit <= 0` → без лимита; одно предупреждение длиннее `limit` → всё равно одной частью)
  - Возвращать `[]string{}` для пустого входа

### Задача 2: Написать тесты `format_test.go`

**Создать:** `notes/internal/domain/taskparser/format_test.go`

- [x] Импортировать `testing`, `time`, `strings`
- [x] Использовать `mkItem` и `withDue` из `filter_test.go` где удобно; для тестов с приоритетом/эмодзи/wikilinks/inline links — использовать `TaskItem{...}` литералы напрямую (т.к. `mkItem` не покрывает эти поля)
- [x] `TestFormatTasks_Empty` — пустой вход возвращает `[]string{}`
- [x] `TestFormatTasks_NilInput` — `nil` вход возвращает `[]string{}` (non-nil)
- [x] `TestFormatTasks_SingleTask` — одна задача с текстом
- [x] `TestFormatTasks_Priority` — `(A)` / `(B)` в начале
- [x] `TestFormatTasks_DueDate` — `due:YYYY-MM-DD` если задана
- [x] `TestFormatTasks_NoDueDate` — без due date нет токена `due:`
- [x] `TestFormatTasks_Emojis` — эмодзи в конце
- [x] `TestFormatTasks_Wikilinks` — `[[link]]` восстановлены
- [x] `TestFormatTasks_InlineLinks` — `[text](url)` восстановлены
- [x] `TestFormatTasks_EmptyTextWithAttributes` — задача с пустым `Text` но с wikilinks/emojis: нет лишних пробелов
- [x] `TestFormatTasks_Subtasks` — подзадачи с `\t` отступом, рекурсивно
- [x] `TestFormatTasks_NestedSubtasks` — 3 уровня вложенности: `\t`, `\t\t` отступы
- [x] `TestFormatTasks_AllFields` — все поля: приоритет, текст, wikilinks, inline links, due, эмодзи, подзадачи
- [x] `TestFormatTasks_LimitSplit` — несколько задач разбиваются по лимиту
- [x] `TestFormatTasks_LimitSingleTaskExceeds` — задача длиннее лимита всё равно в одном part
- [x] `TestFormatTasks_LimitExact` — часть точно по границе лимита
- [x] `TestFormatTasks_ZeroLimit` — `limit = 0` → весь вывод одной частью
- [x] `TestFormatTasks_NegativeLimit` — `limit = -1` → весь вывод одной частью
- [x] `TestFormatWarnings_Empty` — пустой вход возвращает `[]string{}`
- [x] `TestFormatWarnings_NilInput` — `nil` вход возвращает `[]string{}` (non-nil)
- [x] `TestFormatWarnings_Formatted` — `line N: message | text`
- [x] `TestFormatWarnings_LimitSplit` — warnings разбиваются по лимиту
- [x] `TestFormatWarnings_LimitSingleWarningExceeds` — одно предупреждение длиннее лимита → одной частью
- [x] `TestFormatTasks_Integration` — использовать `loadTestTaskFile(t)`, форматнуть с большим лимитом, проверить наличие конкретных подстрок: `[[link car]]`, `due:2030-07-06`, `🏡`, `\t`-отступ для подзадач. Inline links и приоритеты в testdata отсутствуют — покрываются только unit-тестами

### Задача 3: Верификация

- [x] Запустить `go test ./internal/domain/taskparser/...` — все тесты проходят
- [x] Проверить edge cases: пустой слайс, nil-slice, одна задача длиннее лимита, limit <= 0, пустой Text с атрибутами

## Решения и договорённости

- **Порядок полей переопределяет оригинальный план**: оригинальный план (Phase 3, строка 73) указывал «приоритет, текст, эмодзи, due-дата, wikilink как есть». Phase 3 фиксирует другой порядок: приоритет → текст → wikilinks → inline links → due → эмодзи. Это намеренно — wikilinks и inline links логически сгруппированы после текста, эмодзи завершают строку
- **Восстановление извлечённых атрибутов**: парсер разбирает задачу на поля, форматирование собирает обратно в читаемый текст
- **Wikilinks первыми**: `[[wikilink]]` идут перед inline links `[text](url)` и перед due/emojis
- **Due формат**: только `due:YYYY-MM-DD`, без времени. Пустая due — токен отсутствует
- **Эмодзи в конце**: конкатенация в порядке появления в исходной строке
- **Разбивка по границам задач**: никогда не разрезать задачу/подзадачу посередине; oversized задача → oversized part (Telegram обрежет)
- **Возврат**: всегда `[]string{}` (non-nil empty slice), никогда `nil`
- **Subtask indent**: один tab `\t` на уровень вложенности
- **Warning формат**: `line N: <message> | <text>` — через `|` визуально отделяет описание от оригинальной строки
- **`FormatWarnings` сигнатура**: спецификация в фазе 3 (задача 1) документально описывала `FormatWarnings(warnings []Warning) []string` без параметра `limit`, однако реализация принимает `limit int` (разбивка по лимиту — общее поведение с `FormatTasks`). Документация обновлена, чтобы соответствовать реализации: `FormatWarnings(warnings []Warning, limit int) []string`
- **Лимит считается в символах (runes)**: лимит соответствует Telegram-ограничению 4096 символов. `len()` считает байты (кириллица — 2 байта, эмодзи — 4), поэтому разбивка использует `utf8.RuneCountInString`
- **Пустая подзадача**: пустая подзадача (все поля пустые) не выводится — строка с одним `\t`-отступом пропускается[2026-08-23 14:23:47] [decision] task 2: kept TestFormatTasks_ZeroLimit and TestFormatTasks_NegativeLimit asserting plan's documented limit<=0 -> single-part behavior; they expose a hook bug in format.go, so they were NOT weakened
[2026-08-23 14:23:47] [deviation] task 2: go test for the package is NOT fully green — TestFormatTasks_ZeroLimit and TestFormatTasks_NegativeLimit FAIL because format.go does not implement the documented limit<=0 guard (missing `if limit > 0` before the split check). Task instructions prohibit editing format.go; reported bug for orchestrator to fix
