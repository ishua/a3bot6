"""
Модуль для получения курсов валют с API Центрального Банка РФ

API: http://www.cbr.ru/scripts/XML_daily.asp
Документация: https://www.cbr.ru/development/SXML/
"""

import requests
import xml.etree.ElementTree as ET
from datetime import datetime
from typing import Dict, Optional, Union


def get_usd_rub_rate(date: Union[datetime, str], cache: Optional[Dict] = None) -> float:
    """
    Получить курс USD/RUB на указанную дату из ЦБ РФ
    
    Args:
        date: Дата в формате datetime или строка 'YYYY-MM-DD'
        cache: Опциональный словарь для кеширования курсов
        
    Returns:
        float: Курс USD/RUB (рублей за 1 доллар)
        
    Raises:
        ValueError: Если дата некорректна
        Exception: Если не удалось получить курс
        
    Examples:
        >>> rate = get_usd_rub_rate(datetime(2024, 3, 26))
        >>> print(f"Курс: {rate:.4f}")
        Курс: 92.3456
        
        >>> # С кешированием
        >>> cache = {}
        >>> rate1 = get_usd_rub_rate('2024-03-26', cache=cache)
        >>> rate2 = get_usd_rub_rate('2024-03-26', cache=cache)  # Из кеша
    """
    
    # Конвертируем дату в нужный формат
    if isinstance(date, str):
        try:
            date = datetime.fromisoformat(date)
        except ValueError:
            raise ValueError(f"Некорректный формат даты: {date}. Ожидается 'YYYY-MM-DD'")
    
    date_str = date.strftime('%d/%m/%Y')
    
    # Проверяем кеш
    if cache is not None and date_str in cache:
        return cache[date_str]
    
    # Запрашиваем курсы с ЦБ РФ
    rates = _fetch_cbr_rates(date_str)
    
    # Извлекаем курс USD
    if 'USD' not in rates:
        raise Exception(f"Курс USD не найден в ответе ЦБ РФ для даты {date_str}")
    
    usd_rate = rates['USD']
    
    # Сохраняем в кеш
    if cache is not None:
        cache[date_str] = usd_rate
    
    return usd_rate


def _fetch_cbr_rates(date_str: str) -> Dict[str, float]:
    """
    Получить курсы валют с API ЦБ РФ
    
    Args:
        date_str: Дата в формате 'DD/MM/YYYY'
        
    Returns:
        dict: Словарь {currency_code: rate}
        
    Raises:
        Exception: Если запрос неудачный или парсинг провалился
    """
    
    url = f"http://www.cbr.ru/scripts/XML_daily.asp?date_req={date_str}"
    
    try:
        response = requests.get(url, timeout=10)
        response.raise_for_status()
    except requests.exceptions.RequestException as e:
        raise Exception(f"Ошибка при запросе к API ЦБ РФ: {e}")
    
    # Парсим XML
    try:
        rates = _parse_cbr_xml(response.content)
    except Exception as e:
        raise Exception(f"Ошибка при парсинге ответа ЦБ РФ: {e}")
    
    return rates


def _parse_cbr_xml(xml_content: bytes) -> Dict[str, float]:
    """
    Распарсить XML ответ от ЦБ РФ
    
    Args:
        xml_content: Содержимое XML ответа
        
    Returns:
        dict: Словарь {currency_code: rate}
        
    Example XML:
        <ValCurs Date="26.03.2024" name="Foreign Currency Market">
            <Valute ID="R01235">
                <NumCode>840</NumCode>
                <CharCode>USD</CharCode>
                <Nominal>1</Nominal>
                <Name>Доллар США</Name>
                <Value>92,3456</Value>
            </Valute>
            ...
        </ValCurs>
    """
    
    root = ET.fromstring(xml_content)
    rates = {}
    
    for valute in root.findall('Valute'):
        char_code = valute.find('CharCode')
        nominal = valute.find('Nominal')
        value = valute.find('Value')
        
        if char_code is None or nominal is None or value is None:
            continue
        
        currency_code = char_code.text
        nominal_value = float(nominal.text)
        
        # ЦБ использует запятую как разделитель десятичных
        rate_value = float(value.text.replace(',', '.'))
        
        # Курс = значение / номинал
        # Например, для USD: nominal=1, value=92.3456 → rate=92.3456
        rate = rate_value / nominal_value
        
        rates[currency_code] = rate
    
    return rates


# Для удобства: функция с кешем по умолчанию
_global_cache = {}

def get_usd_rub_rate_cached(date: Union[datetime, str]) -> float:
    """
    Получить курс USD/RUB с глобальным кешированием
    
    Args:
        date: Дата в формате datetime или строка 'YYYY-MM-DD'
        
    Returns:
        float: Курс USD/RUB
    """
    return get_usd_rub_rate(date, cache=_global_cache)


if __name__ == '__main__':
    # Тестирование
    print("=" * 60)
    print("ТЕСТ: Получение курса USD/RUB с API ЦБ РФ")
    print("=" * 60)
    print()
    
    # Тест 1: Получение курса на конкретную дату
    test_date = datetime(2024, 3, 26)
    print(f"Тест 1: Курс на {test_date.strftime('%d.%m.%Y')}")
    try:
        rate = get_usd_rub_rate(test_date)
        print(f"✓ USD/RUB = {rate:.4f}")
    except Exception as e:
        print(f"✗ Ошибка: {e}")
    print()
    
    # Тест 2: Использование строки вместо datetime
    print("Тест 2: Курс на 2024-01-15 (строка)")
    try:
        rate = get_usd_rub_rate('2024-01-15')
        print(f"✓ USD/RUB = {rate:.4f}")
    except Exception as e:
        print(f"✗ Ошибка: {e}")
    print()
    
    # Тест 3: Кеширование
    print("Тест 3: Проверка кеширования")
    cache = {}
    
    import time
    start = time.time()
    rate1 = get_usd_rub_rate('2024-02-20', cache=cache)
    time1 = time.time() - start
    
    start = time.time()
    rate2 = get_usd_rub_rate('2024-02-20', cache=cache)
    time2 = time.time() - start
    
    print(f"  Первый запрос: {time1:.3f}s → {rate1:.4f}")
    print(f"  Второй запрос (кеш): {time2:.3f}s → {rate2:.4f}")
    print(f"  ✓ Ускорение: {time1/time2:.1f}x")
    print()
    
    # Тест 4: Текущий курс (сегодня)
    print("Тест 4: Текущий курс (сегодня)")
    try:
        today_rate = get_usd_rub_rate(datetime.now())
        print(f"✓ USD/RUB сегодня = {today_rate:.4f}")
    except Exception as e:
        print(f"✗ Ошибка: {e}")
    print()
    
    print("=" * 60)
    print("Все тесты завершены!")
    print("=" * 60)
