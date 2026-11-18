---
name: "Рацион"
description: "Дневная сводка КБЖУ и список приёмов пищи с раскрываемым составом."
colors:
  page: "#fcfcfa"
  sheet: "#ffffff"
  ink: "#171f2d"
  muted: "#556176"
  line: "#dde2eb"
  rule: "#8490b4"
  input-line: "#7d88a6"
  accent: "#2a4594"
  header: "#fcfcfa"
  header-ink: "#171f2d"
  header-muted: "#556176"
  header-line: "#dde2eb"
  total: "#e9eefb"
  total-divider: "#c9d4f0"
  secondary: "#f1f3fb"
  hover: "#edf0fc"
  action-ink: "#fff"
  focus: "#354dc6"
  error: "#a02e37"
  error-bg: "#f8e7ec"
  page-dark: "#171b29"
  sheet-dark: "#20263a"
  ink-dark: "#f3f4ff"
  muted-dark: "#c0c8e2"
  line-dark: "#414b68"
  rule-dark: "#6e7da8"
  input-line-dark: "#7987ad"
  accent-dark: "#a5b5ff"
  header-dark: "#171b29"
  header-ink-dark: "#f3f4ff"
  header-muted-dark: "#c0c8e2"
  header-line-dark: "#414b68"
  total-dark: "#283551"
  total-divider-dark: "#536186"
  secondary-dark: "#232c43"
  hover-dark: "#2b3450"
  action-ink-dark: "#131c45"
  focus-dark: "#b3c1ff"
  error-dark: "#ffb2be"
  error-bg-dark: "#512d3d"
typography:
  headline:
    fontFamily: "Golos Text, system-ui, sans-serif"
    fontSize: "clamp(2.05rem, 4.6vw, 3.45rem)"
    fontWeight: 760
    lineHeight: 1.04
    letterSpacing: "-.035em"
  plan-name:
    fontFamily: "Golos Text, system-ui, sans-serif"
    fontSize: "clamp(1.55rem, 3.5vw, 2.5625rem)"
    fontWeight: 740
    lineHeight: 1.15
    letterSpacing: "-.035em"
  total:
    fontFamily: "Golos Text, system-ui, sans-serif"
    fontSize: "clamp(2.65rem, 5.625vw, 4.5rem)"
    fontWeight: 730
    lineHeight: 1.03
    letterSpacing: "-.035em"
  total-compact:
    fontFamily: "Golos Text, system-ui, sans-serif"
    fontSize: "clamp(2.4rem, 13.85vw, 3.375rem)"
    fontWeight: 730
    lineHeight: 1.03
    letterSpacing: "-.035em"
  macro:
    fontFamily: "Golos Text, system-ui, sans-serif"
    fontSize: "clamp(1.25rem, 2.2vw, 1.75rem)"
    fontWeight: 720
    lineHeight: 1.15
  macro-compact:
    fontFamily: "Golos Text, system-ui, sans-serif"
    fontSize: "1.4375rem"
    fontWeight: 720
    lineHeight: 1.15
  meal-section-title:
    fontFamily: "Golos Text, system-ui, sans-serif"
    fontSize: "1.3125rem"
    fontWeight: 680
    lineHeight: 1.45
    letterSpacing: "-.02em"
  meal-section-title-compact:
    fontFamily: "Golos Text, system-ui, sans-serif"
    fontSize: "1.1875rem"
    fontWeight: 680
    lineHeight: 1.45
    letterSpacing: "-.02em"
  title:
    fontFamily: "Golos Text, system-ui, sans-serif"
    fontSize: "1.125rem"
    fontWeight: 680
    lineHeight: 1.45
    letterSpacing: "-.02em"
  body:
    fontFamily: "Golos Text, system-ui, sans-serif"
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.45
  action:
    fontFamily: "Golos Text, system-ui, sans-serif"
    fontSize: ".875rem"
    fontWeight: 650
    lineHeight: 1.25
  field-label:
    fontFamily: "Golos Text, system-ui, sans-serif"
    fontSize: ".875rem"
    fontWeight: 640
    lineHeight: 1.45
rounded:
  field: ".2rem"
  button: ".5rem"
  summary: ".75rem"
spacing:
  control-min: "2.75rem"
  control-block: ".45rem"
  field-inline: ".65rem"
  button-inline: ".85rem"
  form-gap: "1rem"
  frame-inset: "clamp(1.5rem, 4.0625vw, 3.25rem)"
  sheet-inset: "clamp(.85rem, 2.4vw, 1.75rem)"
  summary-inset: "clamp(1.25rem, 2.4vw, 1.875rem)"
components:
  button-primary:
    backgroundColor: "{colors.accent}"
    textColor: "{colors.action-ink}"
    typography: "{typography.action}"
    rounded: "{rounded.button}"
    padding: ".45rem .85rem"
  button-secondary:
    backgroundColor: "{colors.sheet}"
    textColor: "{colors.ink}"
    typography: "{typography.action}"
    rounded: "{rounded.button}"
    padding: ".45rem .85rem"
  button-quiet:
    backgroundColor: "transparent"
    textColor: "{colors.accent}"
    typography: "{typography.action}"
    rounded: "{rounded.button}"
    padding: ".45rem .85rem"
  button-danger:
    backgroundColor: "{colors.sheet}"
    textColor: "{colors.error}"
    typography: "{typography.action}"
    rounded: "{rounded.button}"
    padding: ".45rem .85rem"
  input:
    backgroundColor: "{colors.sheet}"
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    rounded: "{rounded.field}"
    padding: ".45rem .65rem"
  navigation:
    backgroundColor: "{colors.header}"
    textColor: "{colors.header-muted}"
  plan-switcher:
    backgroundColor: "transparent"
    textColor: "{colors.ink}"
    typography: "{typography.plan-name}"
    rounded: "{rounded.field}"
    padding: ".15rem 1.6rem .15rem 0"
  day-total:
    backgroundColor: "{colors.total}"
    textColor: "{colors.accent}"
    typography: "{typography.total}"
    rounded: "{rounded.summary}"
    padding: "{spacing.summary-inset}"
  records:
    backgroundColor: "{colors.sheet}"
    textColor: "{colors.ink}"
---

# Дизайн «Рациона»

## Overview

**Визуальный ориентир — дневная сводка и простой список.** На нейтральном фоне выделяется один блок «Итого за день». Пользователь сверяет активный рацион и четыре дневных показателя, затем открывает нужный приём.

## Colors

- Синий акцент выделяет дневные калории, название сводки, основные действия и активный раздел. Фон шапки совпадает с фоном страницы в обеих темах.
- Сводка имеет голубую заливку в светлой теме и синевато-серую в тёмной. Строки приёмов лежат на фоне страницы, детали выделены отдельной заливкой. Приглушённый цвет используется для подписей, единиц, калорий в строках приёмов и состояния продукта.
- Цвет ошибки выделяет подпись у поля, его границу и действие удаления. Общая ошибка показана на отдельной заливке. Тема следует настройке устройства через `prefers-color-scheme`.

## Typography

Локальный вариативный Golos Text (400–900), затем системный шрифт. Название рациона задаёт иерархию заголовка. Дневные калории — самое крупное число, БЖУ образуют второй уровень. Числа используют табличные линейные цифры.

## Layout

- Общий контейнер шириной до 70 rem с адаптивными полями. Шапка занимает ширину окна, её содержимое выровнено по основному контейнеру. На широком экране название рациона слева над сводкой, «Изменить» справа.
- На широком экране сводка слева и список справа делят доступную ширину в пропорции 43:57. В сводке крупные калории расположены над разделителем, ниже три равные колонки БЖУ с обозначением «г» у значений.
- При ширине до 46 rem сводка встаёт над списком, «Изменить» переходит в верхнюю шапку, а навигация закрепляется снизу в три равные колонки. Нижние отступы учитывают безопасную область устройства.
- Формы продукта и рациона ограничены шириной 38 и 54 rem соответственно. В форме продукта три поля БЖУ переходят в две колонки при ширине до 33 rem. Расчётная калорийность показана под полями. Действия переносятся на новую строку, длинные названия в содержимом переносятся внутри строки.

## Elevation & Depth

Дневной итог и раскрытые детали выделены заливкой, строки списка разделены тонкими линиями. Метка текущего раздела находится под названием вкладки в шапке и над ним в нижней навигации.

## Shapes

У дневной сводки скругление .75 rem, у кнопок .5 rem, у полей .2 rem. Список приёмов и раскрытые детали имеют прямые края.

## Components

- **Навигация и выбор версии.** «День», «Версии» и «Продукты» доступны на всех экранах, текущий раздел отмечен через `aria-current`. На активном дне нативный переключатель показывает название в заголовке с подписью «Активный рацион». С JavaScript выбор отправляется сразу, без JavaScript доступна кнопка «Активировать». В списке версий можно открыть рацион и сделать его активным.
- **День.** Главная открывает активный рацион. Для другой версии заголовок содержит её название, доступно действие «Сделать активным». Свёрнутая строка показывает название приёма слева, калории и стрелку справа. Нативный `summary` раскрывает БЖУ приёма и список продуктов. Стрелка поворачивается за 130 мс, переход отключается при настройке уменьшенного движения. В строке продукта видны название, состояние взвешивания при наличии, вес справа и КБЖУ порции ниже. Вес и «г» остаются в одной строке. Первый запуск предлагает создать рацион, пустой состав сопровождается пояснением.
- **Версии.** В каждой строке ссылка на рацион, его дневные КБЖУ, текстовая отметка активной версии и действия активации или копирования по ситуации. Над списком находится «Создать рацион».
- **Продукты.** Над списком находятся поиск по названию и «Добавить продукт». Вся строка с названием, состоянием взвешивания и КБЖУ на 100 г открывает карточку продукта. В пустом справочнике первым действием служит добавление продукта. У поиска без совпадений своё сообщение.
- **Редактор рациона.** Создание, изменение и копирование используют один редактор. Единственная кнопка «Сохранить рацион» остаётся сверху при прокрутке. Позиция содержит выбор продукта и поле «Вес, г», ошибки отмечаются у соответствующего поля. В пустом справочнике «Создать первый продукт» сразу открывает форму. Пока она открыта, видны действия «Сохранить и выбрать» и «Отмена».
- **Форма продукта.** Название, состояние взвешивания и отдельная группа БЖУ на 100 г. Подписи содержат единицы. Расчётная калорийность обновляется при вводе БЖУ и отображается в ккал. Общая ошибка появляется над формой, ошибки полей — рядом с ними. Введённые данные сохраняются при ошибке. Удаление существующего продукта или рациона подтверждается внутри свёрнутого по умолчанию блока.
- **Кнопки и фокус.** Основное действие имеет акцентную заливку, обычное — рамку, тихое — подчёркнутый текст. Кнопки, поля и основные элементы навигации имеют минимальную высоту 2,75 rem (44 CSS-пикселя при базовых 16 px). Видимый фокус — контур 2 px с отступом 3 px, у строки приёма отступ 2 px. Общий экран ошибки показывает сообщение и ссылку «К активному рациону».

## Do's and Don'ts

- Сохранять порядок «ккал, белки, жиры, углеводы». Использовать полные подписи и граммы в сводке и в деталях приёма. Для строк продуктов и справочников использовать «Б, Ж, У».
- Оставлять приёмы свёрнутыми при открытии дня и показывать калории прямо в строке.
- Сохранять текстовую отметку активной версии и видимый фокус при работе с клавиатуры. Шрифты и стили обслуживать вместе с приложением, чтобы интерфейс был доступен в домашней сети.
