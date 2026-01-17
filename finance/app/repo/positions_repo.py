"""
Репозиторий для работы с positions.csv и positions_history.csv
"""
import csv
import os
from typing import List, Dict
from datetime import datetime
from logger import logger
from repo.sync_repo import update_positions_sync


REQUIRED_FIELDS = [
    'date', 'ticker', 'figi', 'account', 'quantity', 
    'current_price', 'currency', 'total_value', 'value_in_rub', 'value_in_usd'
]

HISTORY_FIELDS = [
    'snapshot_date', 'date', 'ticker', 'figi', 'account', 'quantity',
    'current_price', 'currency', 'total_value', 'value_in_rub', 'value_in_usd'
]


def load_positions_snapshot(filepath: str) -> List[Dict]:
    """
    Загружает позиции из CSV файла в плоский список
    
    Args:
        filepath: Путь к CSV файлу (positions.csv или positions_history.csv)
        
    Returns:
        List[Dict] с позициями
        
    Raises:
        FileNotFoundError: Если файл не найден
        Exception: При ошибках чтения
    """
    if not os.path.exists(filepath):
        raise FileNotFoundError(f"file not found: {filepath}")
    
    positions = []
    
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            reader = csv.DictReader(f)
            
            for row in reader:
                position = {
                    'ticker': row['ticker'],
                    'figi': row['figi'],
                    'account': row['account'],
                    'quantity': float(row['quantity']),
                    'current_price': float(row['current_price']),
                    'currency': row['currency'],
                    'total_value': float(row['total_value']),
                    'value_in_rub': float(row['value_in_rub']),
                    'value_in_usd': float(row['value_in_usd'])
                }
                
                positions.append(position)
        
        logger.debug(f"loaded {len(positions)} positions from {filepath}")
        return positions
        
    except Exception as e:
        logger.error(f"failed to load positions from {filepath}: {e}")
        raise


def save_positions(positions: List[Dict], data_dir: str, source: str = 'manual') -> bool:
    """
    Сохраняет позиции в positions.csv (полная перезапись)
    
    Args:
        positions: Список позиций с полями REQUIRED_FIELDS
        data_dir: Путь к директории с данными
        source: Источник данных (tcs-api, manual)
        
    Returns:
        True если успешно, False при ошибке
    """
    if not positions:
        logger.error("positions list is empty")
        return False
    
    if not _validate_positions(positions):
        return False
    
    filepath = os.path.join(data_dir, 'positions.csv')
    
    try:
        os.makedirs(data_dir, exist_ok=True)
        
        with open(filepath, 'w', newline='', encoding='utf-8') as f:
            writer = csv.DictWriter(f, fieldnames=REQUIRED_FIELDS)
            writer.writeheader()
            writer.writerows(positions)
        
        logger.info(f"saved {len(positions)} positions to {filepath}")
        
        update_positions_sync(positions, data_dir, source)
        
        return True
        
    except Exception as e:
        logger.error(f"failed to save positions: {e}")
        return False


def save_positions_history(positions: List[Dict], data_dir: str) -> bool:
    """
    Добавляет снимок позиций в positions_history.csv (append)
    Пропускает, если снимок за текущую дату уже существует
    
    Args:
        positions: Список позиций с полями из REQUIRED_FIELDS
        data_dir: Путь к директории с данными
        
    Returns:
        True если успешно (или пропущено), False при ошибке
    """
    if not positions:
        logger.error("positions list is empty")
        return False
    
    if not _validate_positions(positions):
        return False
    
    filepath = os.path.join(data_dir, 'positions_history.csv')
    snapshot_date = datetime.now().strftime('%Y-%m-%d')
    
    if _snapshot_exists(filepath, snapshot_date):
        logger.info(f"positions_history snapshot for {snapshot_date} already exists, skipping")
        return True
    
    history_records = []
    for pos in positions:
        history_record = {
            'snapshot_date': snapshot_date,
            **pos
        }
        history_records.append(history_record)
    
    try:
        os.makedirs(data_dir, exist_ok=True)
        
        file_exists = os.path.exists(filepath)
        
        with open(filepath, 'a', newline='', encoding='utf-8') as f:
            writer = csv.DictWriter(f, fieldnames=HISTORY_FIELDS)
            
            if not file_exists:
                writer.writeheader()
            
            writer.writerows(history_records)
        
        logger.info(f"appended {len(history_records)} positions to {filepath}")
        
        return True
        
    except Exception as e:
        logger.error(f"failed to save positions_history: {e}")
        return False


def _snapshot_exists(filepath: str, snapshot_date: str) -> bool:
    """
    Проверяет, существует ли уже снимок за указанную дату
    
    Args:
        filepath: Путь к positions_history.csv
        snapshot_date: Дата снимка в формате YYYY-MM-DD
        
    Returns:
        bool: True если снимок уже существует, False если нет
    """
    if not os.path.exists(filepath):
        return False
    
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            reader = csv.DictReader(f)
            
            for row in reader:
                if row.get('snapshot_date') == snapshot_date:
                    return True
        
        return False
        
    except Exception as e:
        logger.error(f"error checking snapshot existence: {e}")
        return False


def _validate_positions(positions: List[Dict]) -> bool:
    """
    Валидирует структуру данных позиций
    """
    for idx, pos in enumerate(positions):
        missing_fields = [f for f in REQUIRED_FIELDS if f not in pos]
        if missing_fields:
            logger.error(f"position {idx} missing fields: {missing_fields}")
            return False
    
    return True
